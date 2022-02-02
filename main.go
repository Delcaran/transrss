package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Tubbebubbe/transmission"
	"github.com/mmcdole/gofeed"
)

type Release struct {
	title   string
	series  string
	episode string
	info    string
	link    string
	hash    string
}

func (rel *Release) checkDownload(cache *OrderedCache, releases []Release) []Release {
	if cache.exists(rel.hash) {
		log.Println("Skipped cached release ", rel.title)
		return releases
	}
	return append(releases, *rel)
}

func (rel *Release) isReplacement() bool {
	return strings.Contains(rel.info, "REPACK") || strings.Contains(rel.info, "PROPER")
}

func (rel *Release) enqueue(tc *transmission.TransmissionClient, cache *OrderedCache, config *Config) {
	// add magnet
	addcmd, err := transmission.NewAddCmdByMagnet(rel.link)
	if err != nil {
		log.Println("Failed creating add cmd: ", err)
		return
	}
	downloadDir := filepath.Join(config.Download, rel.series)
	addcmd.SetDownloadDir(downloadDir)

	_, err = tc.ExecuteAddCommand(addcmd)
	if err != nil {
		log.Println("Failed adding torrent to transmission: ", err)
		return
	}
	cache.add(rel.hash)
	log.Println("Added: ", rel.title)
	cache.commit()

	// check for PROPER/REPACK
	if rel.isReplacement() {
		torrents, err := tc.GetTorrents()
		if err != nil {
			log.Println("Failed to get torrents: ", err)
			return
		}
		for _, torrent := range torrents {
			nameMatch := strings.Contains(torrent.Name, rel.episode)
			dirMatch := (torrent.DownloadDir == downloadDir)
			if nameMatch && dirMatch {
				delcmd, err := transmission.NewDelCmd(torrent.ID, true)
				if err != nil {
					log.Println("Failed creating delete cmd: ", err)
					return
				}
				_, err = tc.ExecuteCommand(delcmd)
				if err != nil {
					log.Println("Failed removing old torrent from transmission: ", err)
				} else {
					log.Println("Removed older release of %s %s", rel.series, rel.episode)
				}
				break
			}
		}
	}
}

type CacheInfo struct {
	Path string `json: "path"`
	Size int    `json: "size"`
}

type RPCInfo struct {
	Host string `json: "host"`
	Port int    `json: "port"`
	User string `json: "user"`
	Pass string `json: "pass"`
}

func (rpc *RPCInfo) URL() string {
	return fmt.Sprintf("%s:%d/transmission/rpc", rpc.Host, rpc.Port)
}

type Config struct {
	Feed     string    `json: "feed"`
	Download string    `json: "download"`
	Cache    CacheInfo `json: "cache"`
	RPC      RPCInfo   `json: "rpc"`
}

var regexTitle = regexp.MustCompile(`^(.+)\s+(S\d{2}E\d{2})\s+(.+)`)
var regexURI = regexp.MustCompile(`^magnet:\?xt=urn:btih:(\w{40})\&.+`)

func buildRelease(item *gofeed.Item) (Release, error) {
	var rel Release
	rel.title = item.Title
	titleMatch := regexTitle.FindAllStringSubmatch(item.Title, -1)
	if len(titleMatch[0]) == 4 {
		rel.series = titleMatch[0][1]
		rel.episode = titleMatch[0][2]
		rel.info = titleMatch[0][3]
	} else {
		return rel, errors.New("No title match")
	}
	uriMatch := regexURI.FindAllStringSubmatch(item.Link, -1)
	if len(uriMatch[0]) == 2 {
		rel.link = item.Link
		rel.hash = uriMatch[0][1]
	} else {
		return rel, errors.New("No hash match")
	}
	return rel, nil
}

func loadConfig(configPath string) (Config, error) {
	var config Config
	if configPath == "" {
		return config, errors.New("Config path not provided")
	}

	log.Println("Running.")

	configFile, err := os.Open(configPath)
	if err != nil {
		return config, errors.New(fmt.Sprintf("Error with opening config file %s: %v", configPath, err))
	}
	defer configFile.Close()

	configData, err := ioutil.ReadAll(configFile)
	if err != nil {
		return config, errors.New(fmt.Sprintf("Error with reading config file %s: %v", configPath, err))
	}

	err = json.Unmarshal(configData, &config)
	if err != nil {
		return config, errors.New(fmt.Sprintf("Error parsing config file %s: %v", configPath, err))
	}
	return config, nil
}

func findReleases(config *Config, cache *OrderedCache) []Release {
	fparser := gofeed.NewParser()
	feed, err := fparser.ParseURL(config.Feed)
	if err != nil {
		log.Fatalln("Failed parsing feed: ", err)
	}
	var releases []Release

	for _, item := range feed.Items {
		rel, err := buildRelease(item)
		if err != nil {
			log.Fatalln("Failed parsing feed item: ", err)
		}
		releases = rel.checkDownload(cache, releases)
	}
	return releases
}

func main() {
	// Load configuration

	configPath := flag.String("config", "./config.json", "Configuration path")

	flag.Parse()
	config, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalln("Failed configuration loading: %v", err)
	}

	// Look for new releases

	cache := newOrderedCache(config.Cache.Path, config.Cache.Size)
	releases := findReleases(&config, cache)

	// enqueue found releases and delete pre-REPACKs and pre-PROPERs

	if len(releases) > 0 {
		tclient := transmission.New(config.RPC.URL(), config.RPC.User, config.RPC.Pass)
		for _, rel := range releases {
			rel.enqueue(&tclient, cache, &config)
		}
	}

	log.Println("Done.")
}
