package main

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

// for !orl
func handleOwner(s *discordgo.Session, m *discordgo.MessageCreate, tokens []string) error {

	if !isAdmin(m.Author.ID) || len(tokens) == 0 {
		return nil
	}

	channel, _ := s.Channel(m.ChannelID)
	guild, err := s.Guild(channel.GuildID)
	if err != nil {
		fmt.Println(err)
		return err
	}

	switch tokens[0] {
	case "tr":
		// enable/disable translations for guild
		ok, err :=  toggleTranslation(guild.ID)
		if err != nil {
			return err
		}
		s.ChannelMessageSend(channel.ID, fmt.Sprintf("translation set to: %t", ok))
		save()
	case "leave":
		// leave guild
		s.ChannelMessageSend(channel.ID, "bye bye")
		s.GuildLeave(guild.ID)
	}

	return nil
}
