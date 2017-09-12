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
		go save()
	case "leave":
		// leave guild
		guid := ""
		if len(tokens) == 2 { // check if a guild id is provided or not
			guid = tokens[1]
		} else {
			guid = channel.GuildID
		}

		s.ChannelMessageSend(channel.ID, fmt.Sprintf("`leaving %s`", guid))
		s.GuildLeave(guid)
	case "status":
		// update bot status
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
