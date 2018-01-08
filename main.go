package main

import (
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

type commandFunc func(s *discordgo.Session, m *discordgo.MessageCreate, tokens []string) error

type command struct {
	Prefix      string
	Handler     commandFunc
	Description string
	Usage       string
}

var (
	helpMessage string
	Commands    = []command{
		// overrustlelogs commands
		{"!logs", cooldown(handleLogs, 3), "returns a link to the userlogs of x person", "channel user (channel is optional)"},
		// {"!mention", dgg(cooldown(handleMentions, 3)), "returns a link to the mentions of x person", "user"},
		// translation commands
		{"!en ", cooldown(handleEnglish, 15), "translate text to english", "text"},
		{"!ja ", cooldown(handleJapanese, 15), "translate text to japanese", "text"},
		{"!tr ", cooldown(handleTranslate, 15), "translate text to ?", "language-code text\n(https://cloud.google.com/translate/docs/languages)"},
		// bot owner commands
		{"!orl translation", isOwner(handleToggleTranslation), "toggle translation features on/off (owner only)", ""},
		{"!orl leave", isOwner(handleLeaveGuild), "leave guild (owner only)", "guildid (optional)"},
		{"!orl status", isOwner(handleSetStatus), "change status message (owner only)", "text"},
		{"!orl stats", isOwner(handleStats), "show bot stats (owner only)", ""},
		// admin role commands
		{"!orl default", isAdmin(handleSetDefault), "set default channel used for !logs (admin only)", "twitchchannelname"},
		{"!orl adminrole", isAdmin(handleSetAdminRole), "set adminrole (admin only)", "roleid"},
		{"!orl ignore", isAdmin(handleIgnore), "ignore user from using commands (admin only)", "userid"},
		{"!orl unignore", isAdmin(handleUnignore), "unignore user from using commands (admin only)", "userid"},
		// bot commands
		{"!orl help", cooldown(handleHelp, 15), "help", ""},
		{"!orl get", cooldown(handleGetBot, 15), "return the join link for this bot", ""},
		// test commands
		{"!test", isOwner(handleTests), "to test shit (owner only)", ""},
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

	// create help message
	helpMessage = listCommands()
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
	defer save()
	s.UpdateStatus(0, fmt.Sprintf("!orl help"))
	time.Sleep(time.Second)
	for _, guild := range m.Guilds {
		log.Printf("joined guild: %s\n", guild.ID)
		set := getSetting(guild.ID)
		set.Name = guild.Name
		so, _ := s.User(guild.OwnerID)
		set.Owner = so.String()
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

	for _, c := range Commands {
		if !strings.HasPrefix(strings.ToLower(m.Content), c.Prefix) {
			continue
		}

		tokens := strings.Split(m.Content[len(c.Prefix):], " ")

		channel, err := getChannel(s, m.ChannelID)
		if err != nil {
			log.Println(err)
			return
		}

		// check if user is ignored on guild
		if isIgnored(channel.GuildID, m.Author.ID) {
			log.Println("ignored", m.Author)
			return
		}

		go func() {

			nm, err := m.ContentWithMoreMentionsReplaced(s)
			if err != nil {
				return
			}

			tokens = strings.Split(nm, " ")

			err = c.Handler(s, m, tokens)
			if err != nil {
				log.Printf("%s tried using command %s and failed with error: %v", m.Author.Username, tokens[0], err)
			}
			// send to master channel for debug/usage info
			go sendToMasterServer(s, m, err)
		}()
		return // should we return here or let it continue? hmmmm
	}
}

// for !test
func handleTests(s *discordgo.Session, m *discordgo.MessageCreate, tokens []string) error {

	channel, err := getChannel(s, m.ChannelID)
	if err != nil {
		log.Println(err)
		return err
	}

	guild, err := getGuild(s, channel.GuildID)
	if err != nil {
		log.Println(err)
		return err
	}

	roles := "```"
	for _, r := range guild.Roles {
		roles += fmt.Sprintf("name=%s, managed=%t, permissions=%d, id=%s\n", r.Name, r.Managed, r.Permissions, r.ID)
	}
	roles += "```"

	s.ChannelMessageSend(masterChannel, roles)
	return nil
}

func sendToMasterServer(s *discordgo.Session, m *discordgo.MessageCreate, cerr error) {

	channel, err := getChannel(s, m.ChannelID)
	if err != nil {
		log.Println(err)
		return
	}

	guild, err := getGuild(s, channel.GuildID)
	if err != nil {
		log.Println(err)
		return
	}

	message := fmt.Sprintf("```[%s][%s] %s: %s", guild.Name, channel.Name, m.Author.Username, m.Content)
	if cerr != nil {
		message += fmt.Sprintf("\n%v", cerr)
	}
	message += "```"
	s.ChannelMessageSend(masterChannel, message)
}

func getChannel(s *discordgo.Session, channelID string) (*discordgo.Channel, error) {
	channel, err := s.State.Channel(channelID)
	if err != nil {
		channel, err = s.Channel(channelID)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
	}
	return channel, nil
}

func getGuild(s *discordgo.Session, guildID string) (*discordgo.Guild, error) {
	guild, err := s.State.Guild(guildID)
	if err != nil {
		guild, err = s.Guild(guildID)
		if err != nil {
			log.Println(err)
			return nil, err
		}
	}
	return guild, nil
}

func handleHelp(s *discordgo.Session, m *discordgo.MessageCreate, tokens []string) error {
	s.ChannelMessageSend(m.ChannelID, helpMessage)
	return nil
}
func handleGetBot(s *discordgo.Session, m *discordgo.MessageCreate, tokens []string) error {
	s.ChannelMessageSend(m.ChannelID, "https://discordapp.com/oauth2/authorize?client_id=217062836189265920&scope=bot&permissions=0")
	return nil
}

func listCommands() string {
	message := "```\n"
	message += "Commands:\n"
	for _, c := range Commands {
		message += fmt.Sprintf("%s %s -- Description: %s\n", c.Prefix, c.Usage, c.Description)
	}
	message += "```"
	return message
}
