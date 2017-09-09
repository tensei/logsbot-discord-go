package main

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

// for !logs
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

// for !mentions
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
