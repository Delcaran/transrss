package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	//"path/filepath"
	"regexp"
	"strconv"

	"github.com/mmcdole/gofeed"
	//"github.com/Tubbebubbe/transmission"
)

type Release struct {
	title   string
	series  string
	season  int
	episode int
	info    string
	link    string
	hash    string
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

type Config struct {
	Feed     string    `json: "feed"`
	Download string    `json: "download"`
	Cache    CacheInfo `json: "cache"`
	RPC      RPCInfo   `json: "rpc"`
}

var regexTitle = regexp.MustCompile(`^(.+)\s+S(\d{2})E(\d{2})\s+(.+)`)
var regexURI = regexp.MustCompile(`^magnet:\?xt=urn:btih:(\w{40})\&.+`)

func buildRelease(item *gofeed.Item) (Release, error) {
	var rel Release
	var err error
	rel.title = item.Title
	titleMatch := regexTitle.FindAllStringSubmatch(item.Title, -1)
	if len(titleMatch[0]) == 5 {
		rel.series = titleMatch[0][1]
		rel.season, err = strconv.Atoi(titleMatch[0][2])
		if err != nil {
			return rel, errors.New(fmt.Sprintf("Failed parsing season number: %v", err))
		}
		rel.episode, err = strconv.Atoi(titleMatch[0][3])
		if err != nil {
			return rel, errors.New(fmt.Sprintf("Failed parsing episode number: %v", err))
		}
		rel.info = titleMatch[0][4]
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

func (rel *Release) checkDownload(cache *OrderedCache, downloadList []string) []string {
	// TODO: check PROPER/REPACK
	if cache.exists(rel.hash) {
		log.Println("Skipped cached release ", rel.title)
		return downloadList
	}
	// TODO: check resolution? --> no, il feed puo' contenere le risoluzioni specifiche
	return append(downloadList, rel.link)
}

func main() {
	// Load configuration

	configPath := flag.String("config", "./config.json", "Configuration path")

	flag.Parse()
	config, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalln("Failed configuration loading: %v", err)
		return
	}
	fmt.Printf("%+v\n", config)

	// Downloading feed

	fparser := gofeed.NewParser()
	feed, err := fparser.ParseURL(config.Feed)
	if err != nil {
		log.Fatalln("Failed parsing feed: ", err)
	}

	// find new releases to download

	cache := newOrderedCache(config.Cache.Path, config.Cache.Size)
	var toDownload []string

	for _, item := range feed.Items {
		rel, err := buildRelease(item)
		if err != nil {
			log.Fatalln("Failed parsing feed item: ", err)
		} else {
			toDownload = rel.checkDownload(cache, toDownload)
		}
	}
	fmt.Printf("%+v\n", toDownload)

	// enqueue found releases

	/*
		if downloadFolder != "" {
			tclient := transmission.New(*tRPC, *tUser, *tPass) http://127.0.0.1:9091/transmission/rpc
			addcmd, err := transmission.NewAddCmdByMagnet(item.Link)
			if err != nil {
				log.Fatalln("Failed creating add cmd: ", err)
			}

			_, err = tclient.ExecuteAddCommand(addcmd)
			if err != nil {
				log.Fatalln("Failed adding torrent to transmission: ", err)
			}
		} else {
			log.Println("Ignored: ", item.Title)
		}
		Cache.add(item.Link)
		log.Println("Added: ", item.Title)
		Cache.commit()
	*/

	log.Println("Done.")
}
