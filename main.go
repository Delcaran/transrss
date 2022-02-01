package main

/*
import (
	"flag"
	"log"

	"github.com/Tubbebubbe/transmission"
	"github.com/mmcdole/gofeed"
)

func main() {
	fURL := flag.String("feed", "", "Feed URL")
	cachePath := flag.String("cache", "cache.json", "Cache path")
	cacheSize := flag.Int("cachesize", 100, "Maximum cache size")
	excludedPath := flag.String("excluded", "excluded.json", "excluded path")
	excludedSize := flag.Int("excludedsize", 100, "Maximum excluded size")
	tRPC := flag.String("transmission", "http://127.0.0.1:9091/transmission/rpc", "Full URL to transmission RPC")
	tPass := flag.String("user", "", "Transmission RPC user")
	tUser := flag.String("password", "", "Transmission RPC password")

	flag.Parse()

	if *fURL == "" {
		log.Fatalln("Feed URL not provided")
	}

	log.Println("Running.")

	cache := newOrderedCache(*cachePath, *cacheSize)
	excluded := newOrderedCache(*excludedPath, *excludedSize) // una "cache" per i link scartati

	tclient := transmission.New(*tRPC, *tUser, *tPass)

	fparser := gofeed.NewParser()
	feed, err := fparser.ParseURL(*fURL)
	if err != nil {
		log.Fatalln("Failed parsing feed: ", err)
	}

	for _, item := range feed.Items {
		if cache.exists(item.Link) {
			continue
		}
		if excluded.exists(item.Link) {
			continue
		}

		downloadFolder := ""
		// TODO: determinare nome serie
		// TODO: determinare episodio
		// TODO: determinare PROPER/REPACK
		// TODO: determinare risoluzione
		// TODO: determinare se scaricare o meno
		// TODO: definire cartella di download

		if downloadFolder != "" {
			addcmd, err := transmission.NewAddCmdByMagnet(item.Link)
			if err != nil {
				log.Fatalln("Failed creating add cmd: ", err)
			}

			_, err = tclient.ExecuteAddCommand(addcmd)
			if err != nil {
				log.Fatalln("Failed adding torrent to transmission: ", err)
			}

			cache.add(item.Link)
			log.Println("Added: ", item.Title)
			cache.commit()
		} else {
			excluded.add(item.Link)
			log.Println("Ignored: ", item.Title)
			excluded.commit()
		}
	}

	log.Println("Done.")
}
*/

import (
	"flag"
	"fmt"
	"log"
	"regexp"
	"strconv"

	"github.com/mmcdole/gofeed"
)

type Release struct {
	series  string
	season  int
	episode int
	info    string
	link    string
	hash    string
}

func buildReleaseClosure() (f func(TODO ITEM) Release) {
	var regexTitle = regexp.MustCompile(`^(.+)\s+S(\d{2})E(\d{2})\s+(.+)`)
	var regexURI = regexp.MustCompile(`^magnet:\?xt=urn:btih:(\w{40})\&.+`)
	f = func(TODO ITEM) Release {
		var rel Release
		titleMatch := regexTitle.FindAllStringSubmatch(item.Title, -1)
		if len(titleMatch[0]) == 5 {
			rel.series = titleMatch[0][1]
			rel.season, err = strconv.Atoi(titleMatch[0][2])
			if err != nil {
				log.Fatalln("Failed parsing season number: ", err)
				// TODO: return error
			}
			rel.episode, err = strconv.Atoi(titleMatch[0][3])
			if err != nil {
				log.Fatalln("Failed parsing episode number: ", err)
				// TODO: return error
			}
			rel.info = titleMatch[0][4]
		} else {
			// TODO: return error
		}
		uriMatch := regexURI.FindAllStringSubmatch(item.Link, -1)
		if len(uriMatch[0]) == 2 {
			rel.link = item.Link
			rel.hash = uriMatch[0][1]
		} else {
			// TODO: return error
		}
		return rel, // TODO: return error
	}
	return
}

func main() {
	fURL := flag.String("feed", "", "Feed URL")

	flag.Parse()

	if *fURL == "" {
		log.Fatalln("Feed URL not provided")
	}

	log.Println("Running.")

	fparser := gofeed.NewParser()
	feed, err := fparser.ParseURL(*fURL)
	if err != nil {
		log.Fatalln("Failed parsing feed: ", err)
	}

	buildRelease := buildReleaseClosure()
	for _, item := range feed.Items {
		rel, err := buildRelease(item)
		break
	}

	log.Println("Done.")
}
