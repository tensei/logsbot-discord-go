package main

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
)

func dgg(f commandFunc) commandFunc {
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

func cooldown(f commandFunc, c time.Duration) commandFunc {
	return func(s *discordgo.Session, m *discordgo.MessageCreate, tokens []string) error {
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

func isOwner(c commandFunc) commandFunc {
	return func(s *discordgo.Session, m *discordgo.MessageCreate, tokens []string) error {
		for _, owner := range owners {
			if m.Author.ID == owner {
				c(s, m, tokens)
				return nil
			}
		}
		return errors.New("not a owner")
	}
}

func isAdmin(c commandFunc) commandFunc {
	return func(s *discordgo.Session, m *discordgo.MessageCreate, tokens []string) error {

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

		if m.Author.ID == guild.OwnerID {
			c(s, m, tokens)
			return nil
		}

		setting := getSetting(channel.GuildID)
		if len(setting.AdminRoles) <= 0 {
			return errors.New("not a admin")
		}

		u, _ := s.GuildMember(channel.GuildID, m.Author.ID)
		for _, role := range u.Roles {
			for _, a := range setting.AdminRoles {
				if a == role {
					c(s, m, tokens)
					return nil
				}
			}
		}

		return errors.New("not a admin")
	}
}
