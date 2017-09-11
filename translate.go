package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// for !en
func handleEnglish(s *discordgo.Session, m *discordgo.MessageCreate, tokens []string) error {

	channel, _ := s.Channel(m.ChannelID)
	guild, err := s.Guild(channel.GuildID)
	if err != nil {
		fmt.Println(err)
		return err
	}

	settings := getSetting(guild.ID)
	if !settings.Translation {
		return errors.New("not allowed")
	}

	query := strings.TrimLeft(m.Content, "!en ")

	resp, err := translate("en", query)
	if err != nil {
		return err
	}

	_, err = s.ChannelMessageSend(channel.ID, fmt.Sprintf("```%s```", resp))

	return err
}

// for !ja
func handleJapanese(s *discordgo.Session, m *discordgo.MessageCreate, tokens []string) error {

	channel, _ := s.Channel(m.ChannelID)
	guild, err := s.Guild(channel.GuildID)
	if err != nil {
		fmt.Println(err)
		return err
	}

	settings := getSetting(guild.ID)
	if !settings.Translation {
		return errors.New("not allowed")
	}

	query := strings.TrimLeft(m.Content, "!ja ")

	resp, err := translate("ja", query)
	if err != nil {
		return err
	}

	_, err = s.ChannelMessageSend(channel.ID, fmt.Sprintf("```%s```", resp))

	return err
}

func translate(t, q string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://tensei.moe/api/v1/translate?t=%s&q=%s", t, q))
	if err != nil || resp.StatusCode == http.StatusNotFound {
		resp.Body.Close()
		return "", err
	}

	defer resp.Body.Close()

	rc, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if string(rc) == "" {
		return "", errors.New("empty response")
	}

	return string(rc), nil
}
