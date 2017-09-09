package main

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"strings"
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
		{"(?i)^!logs?$", handleLogs, "returns a link to the userlogs of x person", nil, nil},
		{"(?i)^!mentions?$", handleMentions, "returns a link to the mentions of x person", nil, nil},
		{"(?i)^!test$", handleTests, "to test shit", nil, nil},
	}

	admins = []string{
		"105739663192363008",
		"127292136843509760",
	}

	guildRatelimits  = map[string]time.Time{}
	defaultRatelimit = time.Second * 10
)

func main() {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		fmt.Println("Missing Token")
		return
	}
	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(messageCreate)
	dg.AddHandler(populateGuilds)

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	// Cleanly close down the Discord session.
	defer dg.Close()

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}

func populateGuilds(s *discordgo.Session, m *discordgo.Ready) {
	for _, guild := range m.Guilds {
		guildRatelimits[guild.ID] = time.Now().UTC()
	}
	fmt.Println("guilds", len(m.Guilds))
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
			go c.Handler(s, m, tokens[1:])
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
	// if guild not in ratelimits add it and ok it
	cd, ok := guildRatelimits[guildid]
	if !ok {
		guildRatelimits[guildid] = time.Now().UTC()
		return false
	}
	// check how much time since last command
	if time.Since(cd) >= defaultRatelimit {
		guildRatelimits[guildid] = time.Now().UTC()
		return false
	}

	return true
}
