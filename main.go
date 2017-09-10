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
		Blacklist   []string
	}{
		// overrustlelogs commands
		{"(?i)^!logs?$", handleLogs, "returns a link to the userlogs of x person", nil, nil},
		{"(?i)^!mentions?$", handleMentions, "returns a link to the mentions of x person", nil, nil},
		// translation commands
		{"(?i)^!en$", handleEnglish, "translate text to english", nil, nil},
		{"(?i)^!ja$", handleJapanese, "translate text to japanese", nil, nil},
		// admin commands
		{"(?i)^!orl$", handleOwner, "for bot owner only", nil, nil},

		// test commands
		{"(?i)^!test$", handleTests, "to test shit", nil, nil},
	}

	admins = []string{
		"105739663192363008",
		"127292136843509760",
	}

	guildRatelimits  = map[string]time.Time{}
	defaultRatelimit = time.Second * 10
	rlmux            sync.RWMutex
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
		guildRatelimits[guild.ID] = time.Now().UTC()
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
			err = c.Handler(s, m, tokens[1:])
			if err != nil {
				log.Printf("%s tried using command %s and failed with error: %v", m.Author.Username, tokens[0], err)
			}
			break // should we break here or let it continue? hmmmm
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

func isRatelimited(guildid, userid string) bool {
	// if admin ignore ratelimit
	if isAdmin(userid) {
		return false
	}
	// lock for changing time
	rlmux.Lock()
	defer rlmux.Unlock()

	// if guild not in ratelimits add it and ok it
	cd, ok := guildRatelimits[guildid]
	if ok && time.Since(cd) >= defaultRatelimit {
		guildRatelimits[guildid] = time.Now().UTC()
		return false
	}

	return true
}
