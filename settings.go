package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"sync"
)

var (
	guildSettings = settings{}
	settingsfile  = "settings.json"
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
	defer mux.Unlock()

	fd, err := ioutil.ReadFile(settingsfile)
	if err != nil {
		fmt.Printf("failed reading settings file: %v", err)
		save()
		return
	}

	err = json.Unmarshal(fd, &guildSettings)
	if err != nil {
		fmt.Printf("failed unmarshal settings file: %v", err)
	}
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

func getSetting(guid string) (guildSetting, bool) {
	mux.RLock()
	defer mux.RUnlock()

	set, ok := guildSettings[guid]
	if !ok {
		addGuild(guid)
		return guildSetting{}, false
	}
	return *set, true
}

func addGuild(guid string) {
	mux.Lock()
	defer mux.Unlock()

	guildSettings[guid] = &guildSetting{
		Guid: guid,
	}
	save()
}

func toggleTranslation(guid string) (bool, error) {
	mux.Lock()
	defer mux.Unlock()

	set, ok := guildSettings[guid]
	if !ok {
		return ok, errors.New("")
	}
	set.Translation = !set.Translation
	return ok, nil
}
