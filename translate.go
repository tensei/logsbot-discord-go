package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
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

	query := strings.Join(tokens, " ")
	query = strings.TrimSpace(query)

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

	query := strings.Join(tokens, " ")
	query = strings.TrimSpace(query)

	resp, err := translate("ja", query)
	if err != nil {
		return err
	}

	_, err = s.ChannelMessageSend(channel.ID, fmt.Sprintf("```%s```", resp))

	return err
}
func handleTranslate(s *discordgo.Session, m *discordgo.MessageCreate, tokens []string) error {

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
	if tokens[0] == "help" {
		_, err = s.ChannelMessageSend(channel.ID, "```look here for supported languages\nhttps://cloud.google.com/translate/docs/languages (need ISO-639-1 Code)\n!tr <ISO-639-1 Code> <text>```")
		return err
	}

	query := strings.Join(tokens[1:], " ")
	query = strings.TrimSpace(query)

	resp, err := translate(tokens[0], query)
	if err != nil {
		return err
	}

	_, err = s.ChannelMessageSend(channel.ID, fmt.Sprintf("```%s```", resp))

	return err
}

func translate(t, q string) (string, error) {
	url := fmt.Sprintf("https://tensei.moe/api/v1/translate?q=%s&t=%s", url.QueryEscape(strings.Replace(q, "\n", " ", -1)), t)

	resp, err := http.Get(url)
	if err != nil || resp.StatusCode == http.StatusNotFound {
		resp.Body.Close()
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("error translating")
	}

	rc, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	ns := strings.TrimSpace(string(rc))

	if ns == "" {
		return "", errors.New("empty response")
	}
	return ns, nil
}
