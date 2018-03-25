package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func handleToggleTranslation(s *discordgo.Session, m *discordgo.MessageCreate, tokens []string) error {
	channel, err := getChannel(s, m.ChannelID)
	if err != nil {
		log.Println(err)
		return err
	}

	// enable/disable translations for guild
	ok, err := toggleTranslation(channel.GuildID)
	if err != nil {
		return err
	}
	s.ChannelMessageSend(channel.ID, fmt.Sprintf("`translation set to: %t`", ok))
	return nil
}

func handleLeaveGuild(s *discordgo.Session, m *discordgo.MessageCreate, tokens []string) error {
	channel, err := getChannel(s, m.ChannelID)
	if err != nil {
		log.Println(err)
		return err
	}

	// leave guild
	guid := ""
	if len(tokens) == 1 { // check if a guild id is provided or not
		guid = tokens[0]
	} else {
		guid = channel.GuildID
	}

	s.ChannelMessageSend(channel.ID, fmt.Sprintf("`leaving %s`", guid))
	return s.GuildLeave(guid)
}

func handleSetStatus(s *discordgo.Session, m *discordgo.MessageCreate, tokens []string) error {
	// update bot status
	return s.UpdateStatus(0, strings.Join(tokens, " "))
}

func handleStats(s *discordgo.Session, m *discordgo.MessageCreate, tokens []string) error {
	return nil
}
