package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	victron "github.com/christianschmizz/go-victron"
)

func main() {
	username := flag.String("username", "", "VRM username")
	password := flag.String("password", "", "VRM password")
	flag.Parse()

	if *username == "" || *password == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	session, err := victron.Login(*username, *password)
	if err != nil {
		log.Fatal().Err(err).Msg("login failed")
	}

	installs, err := session.Installations(session.UserID)
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}
	for _, site := range installs.Records {
		fmt.Printf("Site: %s (ID: %d)\n", site.Name, site.SiteID)

		{
			ov, err := session.SystemOverview(site.SiteID)
			if err != nil {
				log.Fatal().Err(err).Msg("")
			}
			for _, device := range ov.Records.Devices {
				fmt.Printf("\tDev: %s\n", device.Name)
			}
		}

		{
			diag, err := session.Diagnostics(site.SiteID, 1000)
			if err != nil {
				log.Fatal().Err(err).Msg("")
			}

			for _, r := range diag.Records {
				fmt.Printf("\tDesc: %s (Device: %s, ID: %d)\n", r.Description, r.Device, r.DataAttributeID)
			}
		}
	}
}
