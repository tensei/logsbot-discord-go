package main

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// for !orl
func handleOwner(s *discordgo.Session, m *discordgo.MessageCreate, tokens []string) error {

	if !isAdmin(m.Author.ID) || len(tokens) == 0 {
		return nil
	}

	channel, err := s.Channel(m.ChannelID)
	if err != nil {
		fmt.Println(err)
		return err
	}

	switch tokens[0] {
	case "tr":
		// enable/disable translations for guild
		ok, err := toggleTranslation(channel.GuildID)
		if err != nil {
			return err
		}
		s.ChannelMessageSend(channel.ID, fmt.Sprintf("`translation set to: %t`", ok))
		save()
	case "leave":
		// leave guild
		s.ChannelMessageSend(channel.ID, "`bye bye`")
		s.GuildLeave(channel.GuildID)
	case "status":
		s.UpdateStatus(0, strings.Join(tokens[1:], " "))
	case "default":
		// set default channel for x guild
		err := setChannel(channel.GuildID, tokens[1])
		if err == nil {
			s.ChannelMessageSend(channel.ID, fmt.Sprintf("`set default channel to: %s`", tokens[1]))
		}
	}

	return nil
}
