package main

import (
	"errors"
	"fmt"
	"log"
	"regexp"

	"github.com/bwmarrin/discordgo"
)

var (
	idRegex              = regexp.MustCompile("[0-9]+")
	ErrMissingsArguments = errors.New("`missing arguments! !orl help`")
)

func handleSetDefault(s *discordgo.Session, m *discordgo.MessageCreate, tokens []string) error {
	channel, err := getChannel(s, m.ChannelID)
	if err != nil {
		log.Println(err)
		return err
	}

	// set default channel for x guild
	if len(tokens) >= 2 {
		err := setChannel(channel.GuildID, tokens[1])
		if err == nil {
			s.ChannelMessageSend(channel.ID, fmt.Sprintf("`set default channel to: %s`", tokens[1]))
			return nil
		} else {
			log.Println(err)
			return err
		}
	}
	return nil
}

func handleSetAdminRole(s *discordgo.Session, m *discordgo.MessageCreate, tokens []string) error {
	channel, err := getChannel(s, m.ChannelID)
	if err != nil {
		log.Println(err)
		return err
	}

	if len(tokens) == 2 {
		if !idRegex.MatchString(tokens[1]) {
			s.ChannelMessageSend(channel.ID, "`need role id`")
			return errors.New("invalid role id")
		}
		setAdminRole(channel.GuildID, tokens[1])
		s.ChannelMessageSend(channel.ID, fmt.Sprintf("`set adminrole to: %s`", tokens[1]))
		return nil
	}
	return ErrMissingsArguments
}

func handleIgnore(s *discordgo.Session, m *discordgo.MessageCreate, tokens []string) error {
	channel, err := getChannel(s, m.ChannelID)
	if err != nil {
		log.Println(err)
		return err
	}

	if len(tokens) == 2 {
		if !idRegex.MatchString(tokens[1]) {
			s.ChannelMessageSend(channel.ID, "`need user id`")
			return nil
		}
		addIgnore(channel.GuildID, tokens[1])
		s.ChannelMessageSend(channel.ID, fmt.Sprintf("`ignoring: %s`", tokens[1]))
		return nil
	}
	return ErrMissingsArguments
}

func handleUnignore(s *discordgo.Session, m *discordgo.MessageCreate, tokens []string) error {
	channel, err := getChannel(s, m.ChannelID)
	if err != nil {
		log.Println(err)
		return err
	}

	if len(tokens) == 2 {
		if !idRegex.MatchString(tokens[1]) {
			s.ChannelMessageSend(channel.ID, "`invalid user id`")
			return nil
		}
		removeIgnore(channel.GuildID, tokens[1])
		s.ChannelMessageSend(channel.ID, fmt.Sprintf("`unignored: %s`", tokens[1]))
		return nil
	}
	return ErrMissingsArguments
}
