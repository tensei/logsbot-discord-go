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

	channel, _ := s.Channel(m.ChannelID)
	setting := getSetting(channel.GuildID)

	switch len(tokens) {
	case 1:
		if setting.Channel == "" {
			return errors.New("channel not set")
		}
		ex, li, url, d := logsExist(setting.Channel, tokens[0])
		if ex {
			sendOrlResponse(s, channel.ID, setting.Channel, url, tokens[0], d, li)
		}

	case 2:
		ex, li, url, d := logsExist(tokens[0], tokens[1])
		if ex {
			sendOrlResponse(s, channel.ID, tokens[0], url, tokens[1], d, li)
		}
	}
	return nil
}

// for !mentions
func handleMentions(s *discordgo.Session, m *discordgo.MessageCreate, tokens []string) error {
	// todo
	return nil
}

func logsExist(channel, user string) (bool, int, string, string) {
	url := fmt.Sprintf("http://ttv.overrustlelogs.net/%s/%s.txt", channel, user)
	log.Println(url)
	resp, err := http.Get(url)
	if err != nil {
		return false, 0, "", ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, 0, "", ""
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, 0, "", ""
	}
	bodystring := string(body)

	lines := strings.Count(bodystring, "\n")
	date, err := time.Parse("2006-01-02", bodystring[1:11])
	if err != nil {
		date = time.Now().UTC()
	}

	return true, lines, url, date.Format("2006-01-02")
}

func sendOrlResponse(s *discordgo.Session, cid, channel, url, user, date string, lines int) {
	url = strings.TrimSuffix(url, ".txt")
	message := &discordgo.MessageEmbed{
		Title:     "Overrustlelogs",
		URL:       url,
		Timestamp: date,
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
		Color: 0xff5722,
	}
	s.ChannelMessageSendEmbed(cid, message)
}
