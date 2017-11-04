package main

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
)

func dgg(f command) command {
	return func(s *discordgo.Session, m *discordgo.MessageCreate, tokens []string) error {
		channel, err := s.State.Channel(m.ChannelID)
		if err != nil {
			channel, err = s.Channel(m.ChannelID)
			if err != nil {
				log.Println(err)
				return fmt.Errorf("error getting channel info: %v", err)
			}
		}
		if channel.GuildID != "265256381437706240" {
			return fmt.Errorf("not dgg discord: %s", channel.GuildID)
		}
		return f(s, m, tokens)
	}
}

func cooldown(f command, c time.Duration) command {
	return func(s *discordgo.Session, m *discordgo.MessageCreate, tokens []string) error {

		if isOwner(m.Author.ID) {
			return f(s, m, tokens)
		}
		// lock for changing time
		rlmux.Lock()
		defer rlmux.Unlock()

		// if guild not in ratelimits add it and ok it
		cd, ok := guildRatelimits[m.ChannelID]
		if !ok {
			guildRatelimits[m.ChannelID] = time.Now().UTC()
		}

		if !time.Now().UTC().After(cd.Add(time.Second * c)) {
			return errors.New("command is on cooldown")
		}

		err := f(s, m, tokens)
		if err != nil {
			return err
		}

		guildRatelimits[m.ChannelID] = time.Now().UTC()
		return nil
	}
}

func isOwner(userid string) bool {
	for _, owner := range owners {
		if userid == owner {
			return true
		}
	}
	return false
}

func isAdmin(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	channel, err := s.State.Channel(m.ChannelID)
	if err != nil {
		channel, err = s.Channel(m.ChannelID)
		if err != nil {
			log.Println(err)
			return false
		}
	}

	guild, err := s.State.Guild(channel.GuildID)
	if err != nil {
		guild, err = s.Guild(channel.GuildID)
		if err != nil {
			log.Println(err)
			return false
		}
	}

	if m.Author.ID == guild.OwnerID || isOwner(m.Author.ID) {
		return true
	}

	setting := getSetting(channel.GuildID)
	if len(setting.AdminRoles) <= 0 {
		return false
	}

	u, _ := s.GuildMember(channel.GuildID, m.Author.ID)
	for _, role := range u.Roles {
		for _, a := range setting.AdminRoles {
			if a == role {
				return true
			}
		}
	}

	return false
}
