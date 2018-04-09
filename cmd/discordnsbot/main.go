package main

import (
	"flag"
	"github.com/sirupsen/logrus"
	"os"
	"github.com/bwmarrin/discordgo"
	"fmt"
	"github.com/mmichaelb/discorddnsbot/pkg"
	"os/signal"
	"syscall"
)

var discordToken string

func main() {
	flag.StringVar(&discordToken, "token", "", "The Discord Bot token which should be used to authenticate with the Discord API.")
	flag.Parse()
	discordToken = "NDMyNjY2Mzk2MDgyNTY5MjE2.Dawoow.zIROxh9-wBCRks9PlgRpIGSrh_Y"
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
	session.AddHandler(discorddnsbot.NewDNSRequestHandler("!dns ").Handle)
	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}
