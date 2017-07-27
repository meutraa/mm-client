package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/xyproto/recwatch"
)

var nickMap = map[string]string{}
var roomMap = map[string]string{}
var fileMap = make(map[time.Time]string)
var sortedKeys []time.Time
var lastRoom string

type times []time.Time

func (t times) Len() int {
	return len(t)
}

func (t times) Less(i, j int) bool {
	return t[i].Before(t[j])
}

func (t times) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

func walk(path string, info os.FileInfo, err error) error {
	if strings.HasPrefix(info.Name(), "$") {
		fileMap[info.ModTime()] = path
	}
	return nil
}

func parse(path string) (room, sender string) {
	for _, s := range strings.Split(path, "/") {
		if "" == s {
			continue
		}
		switch s[0] {
		case '!':
			room = mapish(s, roomMap)
		case '@':
			sender = mapish(s, nickMap)
		}
	}
	return room, sender
}

func mapish(s string, m map[string]string) (value string) {
	value = m[s]
	if "" == value {
		value = s
	}
	return value
}

func processFile(f string, time time.Time, timeFormat string) {
	message, err := ioutil.ReadFile(f)
	if nil != err {
		log.Println(f)
		return
	}
	msg := strings.TrimSpace(string(message))
	room, sender := parse(f)
	if room != lastRoom {
		fmt.Printf("\n%s\n", room)
		lastRoom = room
	}
	if "me" == sender {
		fmt.Printf("%v%s\n", time.Format(timeFormat), msg)
	} else {
		fmt.Printf("%v\033[1m%s\033[0m\n", time.Format(timeFormat), msg)
	}
}

func loadJSONConfig(config, file string) map[string]string {
	JSON, err := ioutil.ReadFile(path.Join(config, file))
	mapping := map[string]string{}
	if nil != err {
		log.Println("Failed to read mapping:", err)
		return mapping
	}
	if nil != json.Unmarshal(JSON, &mapping) {
		log.Println("Failed to parse mapping:", err)
	}
	for k, v := range mapping {
		mapping[k] = strings.Replace(v, "\\033", "\033", -1)
	}
	return mapping
}

func xdgDataDir(home string) string {
	dir := os.Getenv("XDG_DATA_HOME")
	if "" == dir {
		dir = path.Join(home, ".local", "share", "mm")
	} else {
		dir = path.Join(dir, "mm")
	}
	return dir
}

func xdgConfigDir(home string) string {
	dir := os.Getenv("XDG_CONFIG_HOME")
	if "" == dir {
		dir = path.Join(home, ".config", "mm")
	} else {
		dir = path.Join(dir, "mm")
	}
	return dir
}

func main() {
	usr, err := user.Current()
	if nil != err {
		log.Println("Unable to get current user:", err)
		os.Exit(1)
	}

	var config, data, timeFormat string
	flag.StringVar(&data, "d", xdgDataDir(usr.HomeDir), "data storage directory")
	flag.StringVar(&config, "c", xdgConfigDir(usr.HomeDir), "config directory")
	flag.StringVar(&timeFormat, "f", "15:04    ", "time format")
	flag.Parse()

	/* Try to read config files. */
	roomMap = loadJSONConfig(config, "rooms.json")
	nickMap = loadJSONConfig(config, "accounts.json")

	filepath.Walk(data, walk)
	for k := range fileMap {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Sort(times(sortedKeys))
	for _, key := range sortedKeys {
		processFile(fileMap[key], key, timeFormat)
	}

	watcher, err := recwatch.NewRecursiveWatcher(data)
	if nil != err {
		println("Failed to watch data directory for new messages", err)
		os.Exit(1)
	}

	for {
		ev := <-watcher.Events
		if fsnotify.Create != fsnotify.Event(ev).Op {
			continue
		}
		file := fsnotify.Event(ev).Name

		/* I know this is hacky but it saves a little bit. */
		if !strings.Contains(file, "$") {
			continue
		}

		info, err := os.Stat(file)
		if err != nil {
			log.Println("Failed to stat message file:", err)
			continue
		}
		if _, exists := fileMap[info.ModTime()]; nil == err && !exists {
			fileMap[info.ModTime()] = file
			processFile(file, info.ModTime(), timeFormat)
		}
	}
}
