package main

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

type command func(s *discordgo.Session, m *discordgo.MessageCreate, tokens []string) error

var (
	commands = map[string]command{
		"!log":      handleLogs,
		"!logs":     handleLogs,
		"!mentions": handleMentions,
		"!test":     handleTests,
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
	guilds := m.Guilds
	for _, guild := range guilds {
		guildRatelimits[guild.ID] = time.Now()
	}
	fmt.Println("guilds", len(guilds))
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

	cmnd, ok := commands[tokens[0]]
	if ok {
		cmnd(s, m, tokens)
	}
}

func handleLogs(s *discordgo.Session, m *discordgo.MessageCreate, tokens []string) error {

	channel, _ := s.Channel(m.ChannelID)
	guild, err := s.Guild(channel.GuildID)
	if err != nil {
		fmt.Println(err)
		return err
	}

	if isRatelimited(guild.ID, m.Author.ID) {
		return nil
	}

	message := fmt.Sprintf("%s %s", m.Author.Mention(), m.Content)
	s.ChannelMessageSend(m.ChannelID, message)
	return nil
}

func handleMentions(s *discordgo.Session, m *discordgo.MessageCreate, tokens []string) error {

	channel, _ := s.Channel(m.ChannelID)
	guild, err := s.Guild(channel.GuildID)
	if err != nil {
		fmt.Println(err)
		return err
	}

	if isRatelimited(guild.ID, m.Author.ID) {
		return nil
	}

	message := fmt.Sprintf("%s %s", m.Author.Mention(), m.Content)
	s.ChannelMessageSend(m.ChannelID, message)
	return nil
}

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
		guildRatelimits[guildid] = time.Now()
		return false
	}
	// check how much time since last command
	if time.Since(cd) >= defaultRatelimit {
		guildRatelimits[guildid] = time.Now()
		return false
	}

	return true
}
