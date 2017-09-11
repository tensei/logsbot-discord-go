package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
)

var (
	guildSettings = make(settings)
	settingsfile  = os.Getenv("SETTINGS_FILE")
	mux           sync.RWMutex
)

type (
	settings map[string]*guildSetting

	guildSetting struct {
		Guid        string   `json:"guid"`
		Channel     string   `json:"channel"`
		Translation bool     `json:"translation"`
		Ignorelist  []string `json:"ignore_list"`
	}
)

func load() {
	mux.Lock()

	fd, err := ioutil.ReadFile(settingsfile)
	if err != nil {
		fmt.Printf("failed reading settings file: %v", err)
		mux.Unlock()
		save()
		return
	}

	err = json.Unmarshal(fd, &guildSettings)
	if err != nil {
		fmt.Printf("failed unmarshal settings file: %v", err)
	}
	mux.Unlock()
}

func save() {
	mux.RLock()
	defer mux.RUnlock()

	fd, err := json.MarshalIndent(&guildSettings, "", "    ")
	if err != nil {
		fmt.Printf("failed marshal settings file: %v", err)
		return
	}

	err = ioutil.WriteFile(settingsfile, fd, 0755)
	if err != nil {
		fmt.Printf("failed writing settings file: %v", err)
	}
}

func getSetting(guid string) *guildSetting {
	if guid == "" {
		return nil
	}
	mux.RLock()

	set, ok := guildSettings[guid]
	if !ok {
		mux.RUnlock()
		return addGuild(guid)
	}
	mux.RUnlock()
	return set
}

func addGuild(guid string) *guildSetting {
	mux.Lock()

	gs := &guildSetting{
		Guid: guid,
	}
	guildSettings[guid] = gs

	mux.Unlock()
	save()
	return gs
}

func toggleTranslation(guid string) (bool, error) {
	mux.Lock()
	defer mux.Unlock()

	set, ok := guildSettings[guid]
	if !ok {
		return ok, errors.New("couldn't get settings for " + guid)
	}
	set.Translation = !set.Translation
	return set.Translation, nil
}

func setChannel(guid, channel string) error {
	mux.Lock()

	set, ok := guildSettings[guid]
	if !ok {
		mux.Unlock()
		return errors.New("couldn't get settings for " + guid)
	}
	set.Channel = channel
	mux.Unlock()
	save()
	return nil
}
