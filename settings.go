package main

import (
	"encoding/json"
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
		Name        string   `json:"name"`
		Owner       string   `json:"owner"`
		Guid        string   `json:"guid"`
		Channel     string   `json:"channel"`
		Translation bool     `json:"translation"`
		Ignorelist  []string `json:"ignore_list"`
		AdminRoles  []string `json:"admin_roles"`
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
	defer save()

	mux.Lock()
	gs := &guildSetting{
		Guid: guid,
	}
	guildSettings[guid] = gs
	mux.Unlock()

	return gs
}

func toggleTranslation(guid string) (bool, error) {
	defer save()
	mux.Lock()
	defer mux.Unlock()

	set, ok := guildSettings[guid]
	if !ok {
		return ok, fmt.Errorf("couldn't get settings for %s", guid)
	}
	set.Translation = !set.Translation
	return set.Translation, nil
}

func setChannel(guid, channel string) error {
	if !channelRegex.MatchString(channel) {
		return fmt.Errorf("not a valid channel name: %s", channel)
	}
	mux.Lock()
	set, ok := guildSettings[guid]
	if !ok {
		mux.Unlock()
		return fmt.Errorf("couldn't get settings for %s", guid)
	}
	set.Channel = channel
	mux.Unlock()
	save()
	return nil
}

// add user to ignore list
func addIgnore(guid, uid string) bool {
	set := getSetting(guid)
	for _, u := range set.Ignorelist {
		if u == uid {
			return false
		}
	}
	mux.Lock()
	set.Ignorelist = append(set.Ignorelist, uid)
	mux.Unlock()
	go save()
	return true
}

// remove user from ignore list
func removeIgnore(guid, uid string) bool {
	set := getSetting(guid)
	for i, u := range set.Ignorelist {
		if u == uid {
			mux.Lock()
			set.Ignorelist = append(set.Ignorelist[:i], set.Ignorelist[i+1:]...)
			mux.Unlock()
			go save()
			return true
		}
	}
	return false
}

func setAdminRole(guid, roleid string) {
	set := getSetting(guid)
	mux.Lock()
	for _, a := range set.AdminRoles {
		if a == roleid {
			mux.Unlock()
			return
		}
	}
	set.AdminRoles = append(set.AdminRoles, roleid)
	mux.Unlock()
	go save()
}

func isIgnored(guid, uid string) bool {
	set := getSetting(guid)
	for _, u := range set.Ignorelist {
		if uid == u {
			return true
		}
	}
	return false
}
