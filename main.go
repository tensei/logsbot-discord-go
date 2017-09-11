package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

type command func(s *discordgo.Session, m *discordgo.MessageCreate, tokens []string) error

var (
	commands = []struct {
		Command     string
		Handler     command
		Description string
		Whitelist   []string
		Cooldown    time.Duration //in seconds
	}{
		// overrustlelogs commands
		{"(?i)^!logs?$", handleLogs, "returns a link to the userlogs of x person", nil, 10},
		{"(?i)^!mentions?$", handleMentions, "returns a link to the mentions of x person", nil, 10},
		// translation commands
		{"(?i)^!en$", handleEnglish, "translate text to english", nil, 15},
		{"(?i)^!ja$", handleJapanese, "translate text to japanese", nil, 15},
		// admin commands
		{"(?i)^!orl$", handleOwner, "for bot owner only", nil, 0},
		// test commands
		{"(?i)^!test$", handleTests, "to test shit", nil, 0},
	}

	admins = []string{
		"105739663192363008",
		"127292136843509760",
	}
	masterChannel = "356704761732530177"

	guildRatelimits = map[string]time.Time{}
	rlmux           sync.RWMutex
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Println("missing Token")
		return
	}
	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Println("error creating Discord session,", err)
		return
	}

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(messageCreate)
	dg.AddHandler(populateGuilds)

	// load settings
	load()

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		log.Println("error opening connection,", err)
		return
	}

	// Cleanly close down the Discord session.
	defer dg.Close()

	// Wait here until CTRL-C or other term signal is received.
	log.Println("bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}

func populateGuilds(s *discordgo.Session, m *discordgo.Ready) {
	for _, guild := range m.Guilds {
		log.Println("joined guild: ", guild.Name, guild.ID)
		getSetting(guild.ID)
		for _, ch := range guild.Channels {
			rlmux.Lock()
			guildRatelimits[ch.ID] = time.Now().UTC()
			rlmux.Unlock()
		}
	}
	log.Println("guilds", len(m.Guilds))
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	tokens := strings.Split(m.Content, " ")

	if len(tokens) == 0 {
		return
	}

	for _, c := range commands {
		regex, err := regexp.Compile(c.Command)
		if err != nil {
			fmt.Println(err)
			continue
		}

		if regex.MatchString(tokens[0]) {
			if err != nil {
				log.Println(err)
				return
			}
			// check if i can post again
			if isRatelimited(m.ChannelID, m.Author.ID, c.Cooldown) {
				log.Println("rate limited")
				return
			}
			go sendToMasterServer(s, m)
			err = c.Handler(s, m, tokens[1:])
			if err != nil {
				log.Printf("%s tried using command %s and failed with error: %v", m.Author.Username, tokens[0], err)
			}
			return // should we return here or let it continue? hmmmm
		}
	}
}

// for !test
func handleTests(s *discordgo.Session, m *discordgo.MessageCreate, tokens []string) error {
	if !isAdmin(m.Author.ID) {
		return errors.New("not a admin")
	}

	channel, _ := s.Channel(m.ChannelID)
	guild, err := s.Guild(channel.GuildID)
	if err != nil {
		fmt.Println(err)
	}

	message := fmt.Sprintf("%s %s %s", m.Author.Mention(), guild.Name, channel.Name)

	s.ChannelMessageSend(m.ChannelID, message)
	return nil
}

func isAdmin(userid string) bool {
	for _, admin := range admins {
		if userid == admin {
			return true
		}
	}
	return false
}

func isRatelimited(cuid, userid string, cooldown time.Duration) bool {
	// if admin ignore ratelimit
	if isAdmin(userid) {
		return false
	}
	// lock for changing time
	rlmux.Lock()
	defer rlmux.Unlock()

	// if guild not in ratelimits add it and ok it
	cd, ok := guildRatelimits[cuid]
	if ok && time.Since(cd) >= time.Second*cooldown {
		guildRatelimits[cuid] = time.Now().UTC()
		return false
	}

	return true
}

func sendToMasterServer(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Attempt to get the channel from the state.
	// If there is an error, fall back to the restapi
	channel, err := s.State.Channel(m.ChannelID)
	if err != nil {
		channel, err = s.Channel(m.ChannelID)
		if err != nil {
			return
		}
	}

	// Attempt to get the guild from the state,
	// If there is an error, fall back to the restapi.
	guild, err := s.State.Guild(channel.GuildID)
	if err != nil {
		guild, err = s.Guild(channel.GuildID)
		if err != nil {
			return
		}
	}
	s.ChannelMessageSend(masterChannel, fmt.Sprintf("```[%s][%s] %s: %s```", guild.Name, channel.Name, m.Author.Username, m.Content))
}
