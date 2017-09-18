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
		Whitelist   []*string
	}{
		// overrustlelogs commands
		{"(?i)^!logs?$", cooldown(handleLogs, 10), "returns a link to the userlogs of x person", nil},
		{"(?i)^!mentions?$", cooldown(handleMentions, 10), "returns a link to the mentions of x person", nil},
		// translation commands
		{"(?i)^!en$", cooldown(handleEnglish, 15), "translate text to english", nil},
		{"(?i)^!ja$", cooldown(handleJapanese, 15), "translate text to japanese", nil},
		// bot owner commands
		{"(?i)^!oorl$", handleOwner, "for bot owner only", nil},
		// admin role commands
		{"(?i)^!orl$", handleAdmins, "for bot owner only", nil},
		// test commands
		{"(?i)^!test$", handleTests, "to test shit", nil},
	}

	owners = []string{
		"105739663192363008", // tensei
		"127292136843509760", // dbc
	}
	masterChannel = "356704761732530177"

	guildRatelimits = make(map[string]time.Time)
	rlmux           sync.RWMutex

	usernameRegex = regexp.MustCompile("^[a-zA-Z0-9_]+$")
	channelRegex  = regexp.MustCompile("^[a-zA-Z0-9_-]+$")
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
	// and save last time
	defer save()

	// Wait here until CTRL-C or other term signal is received.
	log.Println("bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}

func populateGuilds(s *discordgo.Session, m *discordgo.Ready) {
	s.UpdateStatus(0, fmt.Sprintf("!logs <channel?> <user>"))
	time.Sleep(time.Second)
	for _, guild := range m.Guilds {
		log.Printf("joined guild: %s\n", guild.ID)
		getSetting(guild.ID)
		for _, ch := range guild.Channels {
			rlmux.Lock()
			guildRatelimits[ch.ID] = time.Now().UTC().Add(-time.Minute)
			rlmux.Unlock()
		}
	}
	log.Printf("guilds %d\nchannels %d\n", len(m.Guilds), len(guildRatelimits))
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
			channel, err := s.State.Channel(m.ChannelID)
			if err != nil {
				channel, err = s.Channel(m.ChannelID)
				if err != nil {
					log.Println(err)
					return
				}
			}
			// check if user is ignored on guild
			if isIgnored(channel.GuildID, m.Author.ID) {
				log.Println("ignored", m.Author)
				return
			}

			err = c.Handler(s, m, tokens[1:])
			if err != nil {
				log.Printf("%s tried using command %s and failed with error: %v", m.Author.Username, tokens[0], err)
			}
			// send to master channel for debug/usage info
			go sendToMasterServer(s, m, err)
			return // should we return here or let it continue? hmmmm
		}
	}
}

// for !test
func handleTests(s *discordgo.Session, m *discordgo.MessageCreate, tokens []string) error {
	if !isOwner(m.Author.ID) {
		return errors.New("not a admin")
	}

	channel, _ := s.Channel(m.ChannelID)
	guild, err := s.Guild(channel.GuildID)
	if err != nil {
		fmt.Println(err)
	}

	roles := "```"
	for _, r := range guild.Roles {
		roles += fmt.Sprintf("name=%s, managed=%t, permissions=%d, id=%s\n", r.Name, r.Managed, r.Permissions, r.ID)
	}
	roles += "```"

	s.ChannelMessageSend(masterChannel, roles)
	return nil
}

func isOwner(userid string) bool {
	for _, owner := range owners {
		if userid == owner {
			return true
		}
	}
	return false
}

func isAdmin(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	channel, err := s.State.Channel(m.ChannelID)
	if err != nil {
		channel, err = s.Channel(m.ChannelID)
		if err != nil {
			log.Println(err)
			return false
		}
	}

	guild, err := s.State.Guild(channel.GuildID)
	if err != nil {
		guild, err = s.Guild(channel.GuildID)
		if err != nil {
			log.Println(err)
			return false
		}
	}

	if m.Author.ID == guild.OwnerID || isOwner(m.Author.ID) {
		return true
	}

	setting := getSetting(channel.GuildID)
	if len(setting.AdminRoles) <= 0 {
		return false
	}

	u, _ := s.GuildMember(channel.GuildID, m.Author.ID)
	for _, role := range u.Roles {
		for _, a := range setting.AdminRoles {
			if a == role {
				return true
			}
		}
	}

	return false
}

func sendToMasterServer(s *discordgo.Session, m *discordgo.MessageCreate, cerr error) {
	// Attempt to get the channel from the state.
	// If there is an error, fall back to the restapi
	channel, err := s.State.Channel(m.ChannelID)
	if err != nil {
		channel, err = s.Channel(m.ChannelID)
		if err != nil {
			log.Println(err)
			return
		}
	}

	// Attempt to get the guild from the state,
	// If there is an error, fall back to the restapi.
	guild, err := s.State.Guild(channel.GuildID)
	if err != nil {
		guild, err = s.Guild(channel.GuildID)
		if err != nil {
			log.Println(err)
			return
		}
	}
	message := fmt.Sprintf("```[%s][%s] %s: %s", guild.Name, channel.Name, m.Author.Username, m.Content)
	if cerr != nil {
		message += fmt.Sprintf("\n%v", cerr)
	}
	message += "```"
	s.ChannelMessageSend(masterChannel, message)
}

func cooldown(f command, c time.Duration) command {
	return func(s *discordgo.Session, m *discordgo.MessageCreate, tokens []string) error {

		if isOwner(m.Author.ID) {
			return f(s, m, tokens)
		}
		// lock for changing time
		rlmux.Lock()
		defer rlmux.Unlock()

		// if guild not in ratelimits add it and ok it
		cd, ok := guildRatelimits[m.ChannelID]
		if !ok {
			guildRatelimits[m.ChannelID] = time.Now().UTC()
		}

		if !time.Now().UTC().After(cd.Add(time.Second * c)) {
			return errors.New("command is on cooldown")
		}

		err := f(s, m, tokens)
		if err != nil {
			return err
		}

		guildRatelimits[m.ChannelID] = time.Now().UTC()
		return nil
	}
}
