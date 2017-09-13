package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

// for !logs
func handleLogs(s *discordgo.Session, m *discordgo.MessageCreate, tokens []string) error {

	channel, err := s.State.Channel(m.ChannelID)
	if err != nil {
		channel, err = s.Channel(m.ChannelID)
		if err != nil {
			log.Println(err)
			return err
		}
	}

	setting := getSetting(channel.GuildID)

	ln := len(tokens)
	switch {
	case ln == 0:
		if setting.Channel == "" {
			return errors.New("channel not set")
		}
		ex, li, url, d := logsExist(setting.Channel, m.Author.Username)
		if ex {
			sendOrlResponse(s, channel.ID, setting.Channel, url, m.Author.Username, d, li)
			return nil
		}
	case ln == 1:
		if setting.Channel == "" {
			return errors.New("channel not set")
		}
		if !usernameRegex.MatchString(tokens[0]) {
			return errors.New(fmt.Sprintf("not a valid username: %s", tokens[0]))
		}
		ex, li, url, d := logsExist(setting.Channel, tokens[0])
		if ex {
			sendOrlResponse(s, channel.ID, setting.Channel, url, tokens[0], d, li)
			return nil
		}

	case ln >= 2:
		if !channelRegex.MatchString(tokens[0]) {
			return errors.New(fmt.Sprintf("not a valid channel name: %s", tokens[1]))
		}
		ex, li, url, d := logsExist(tokens[0], tokens[1])
		if ex {
			sendOrlResponse(s, channel.ID, tokens[0], url, tokens[1], d, li)
			return nil
		}
	}
	return errors.New("no command executed")
}

// for !mentions
func handleMentions(s *discordgo.Session, m *discordgo.MessageCreate, tokens []string) error {
	// todo
	return errors.New("not implemented")
}

func logsExist(channel, user string) (bool, int, string, time.Time) {
	user = strings.TrimSpace(user)
	channel = strings.TrimSpace(channel)
	url := fmt.Sprintf("http://ttv.overrustlelogs.net/%s/%s.txt", channel, user)
	log.Println(url)
	resp, err := http.Get(url)
	if err != nil {
		return false, 0, "", time.Now()
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, 0, "", time.Now()
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, 0, "", time.Now()
	}
	bodystring := string(body)

	lines := strings.Count(bodystring, "\n")
	date, err := time.Parse("2006-01-02", bodystring[1:11])
	if err != nil {
		date = time.Now().UTC()
	}

	return true, lines, url, date
}

func sendOrlResponse(s *discordgo.Session, cid, channel, url, user string, date time.Time, lines int) {
	url = strings.TrimSuffix(url, ".txt")
	message := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			URL:     "https://overrustlelogs.net/",
			Name:    s.State.User.Username,
			IconURL: s.State.User.AvatarURL(""),
		},
		Title: "Go to logs",
		URL:   url,
		Fields: []*discordgo.MessageEmbedField{
			&discordgo.MessageEmbedField{
				Name:   "User",
				Value:  user,
				Inline: true,
			},
			&discordgo.MessageEmbedField{
				Name:   "Channel",
				Value:  channel,
				Inline: true,
			},
			&discordgo.MessageEmbedField{
				Name:   "Lines",
				Value:  strconv.Itoa(lines),
				Inline: true,
			},
		},
		// Description: url,
		Color: 0xff5722, // orange
		Footer: &discordgo.MessageEmbedFooter{
			Text: date.Format("January 2006"),
		},
	}
	s.ChannelMessageSendEmbed(cid, message)
}
