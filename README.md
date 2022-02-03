# Transrss

[![Go Report Card](https://goreportcard.com/badge/github.com/grooveygr/transrss)](https://goreportcard.com/report/github.com/grooveygr/transrss)

Transrss glues torrent rss feeds to the Transmission bitorrent client


## Features

- Consume feeds as provided by [showrss.info](https://showrss.info)
  - No episode filter: your feed must provide only one release for each episode
  - REPACKs and PROPERs replace the original release if present in Transmission
- Rolling dedup cache of latest torrents
- (Very) simple cache persistence

## Building

1. Clone this repository
2. Enter the cloned directory
3. Build it

Cross compile to any supported golang platform. 
Example for RPI 1:
```
GOOS=linux GOARCH=arm GOARM=6 go build
```
Example for RPI 3:
```
GOOS=linux GOARCH=arm GOARM=7 go build
```

An example Dockerfile is provided for building an executable suitable for working under Alpine Linux running on an ARM device such as Raspberry Pi 1.

## Usage

All options are listed in a (hopefully) self-explainatory configuration file.
A command-line option can change the configuration file that will be parsed.

```
Usage of ./transrss:
  -config <configuration file>
        Cache path (default "./config.json")
```

Automate rss checking using your favorite scheduler. Crontab example:

1. Edit crontab file by running:
```
crontab -e
```

2. Add the relevant crontab entry (check every 30 minutes):
```
*/30 * * * * /absolute/path/to/transrss
```

## Possible Issues

- Only releases with episode numbers written as S##E## are accepted (where ## are digits)
- Multiple releases for the same episode will be downloaded
  - propers and repacks "cleanup" may not work reliably under this condition
- Episode history (ie, if the episode has already been downloaded) is not implemented
  - This is different from the cache: cache is based on torrent's hash, not on release properties

## TODOs

- [] Parse *all* possible episode formats: S#E#, #x##, #x#, #.##, #.#, #-##, etc...
- [] Handle the "multiple releases per episode" scenario
- [] Manage episode history
