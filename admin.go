package main

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/bwmarrin/discordgo"
)

var (
	idRegex = regexp.MustCompile("[0-9]+")
)

func handleAdmins(s *discordgo.Session, m *discordgo.MessageCreate, tokens []string) error {
	if !isAdmin(s, m) || len(tokens) == 0 {
		return errors.New("not a admin")
	}

	channel, err := s.State.Channel(m.ChannelID)
	if err != nil {
		channel, err = s.Channel(m.ChannelID)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}

	switch tokens[0] {
	case "default", "df":
		// set default channel for x guild
		if len(tokens) >= 2 {
			err := setChannel(channel.GuildID, tokens[1])
			if err == nil {
				s.ChannelMessageSend(channel.ID, fmt.Sprintf("`set default channel to: %s`", tokens[1]))
			}
			break
		}
		s.ChannelMessageSend(channel.ID, "`missing channel name`")
	case "ar", "adminrole":
		if len(tokens) == 2 {
			if !idRegex.MatchString(tokens[1]) {
				s.ChannelMessageSend(channel.ID, "`need channel id`")
				break
			}
			setAdminRole(channel.GuildID, tokens[1])
			s.ChannelMessageSend(channel.ID, fmt.Sprintf("`set adminrole to: %s`", tokens[1]))
			break
		}
		s.ChannelMessageSend(channel.ID, "missing role id")
	case "ignore":
		if len(tokens) == 2 {
			if !idRegex.MatchString(tokens[1]) {
				s.ChannelMessageSend(channel.ID, "`need user id`")
				break
			}
			addIgnore(channel.GuildID, tokens[1])
			s.ChannelMessageSend(channel.ID, fmt.Sprintf("`ignoring: %s`", tokens[1]))
			break
		}
		s.ChannelMessageSend(channel.ID, "missing user id")
	case "unignore":
		if len(tokens) == 2 {
			if !idRegex.MatchString(tokens[1]) {
				s.ChannelMessageSend(channel.ID, "`need user id`")
				break
			}
			removeIgnore(channel.GuildID, tokens[1])
			s.ChannelMessageSend(channel.ID, fmt.Sprintf("`unignored: %s`", tokens[1]))
			break
		}
		s.ChannelMessageSend(channel.ID, "missing user id")
	}
	return nil
}
