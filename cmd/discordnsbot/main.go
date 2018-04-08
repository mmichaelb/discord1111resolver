package main

import (
	"flag"
	"github.com/sirupsen/logrus"
	"os"
	"github.com/bwmarrin/discordgo"
	"fmt"
)

var discordToken string

func main() {
	flag.StringVar(&discordToken, "token", "", "The Discord Bot token which should be used to authenticate with the Discord API.")
	flag.Parse()
	// check if a discord API token is provided
	if discordToken == "" {
		logrus.Print("no Discord token provided")
		flag.PrintDefaults()
		os.Exit(1)
	}
	session, err := discordgo.New(fmt.Sprintf("Bot %v", discordToken))
	if err != nil {
		logrus.WithError(err).Fatal("could not connect to Discord API")
	}
	if err := session.Open(); err != nil {
		logrus.WithError(err).Fatal("could not open Discord session")
	}
}
