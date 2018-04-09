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
	"github.com/miekg/dns"
)

var applicationName, version, branch, commit string

var discordToken string

func main() {
	logrus.WithField("name", applicationName).WithField("version", version).WithField("branch", branch).WithField("commit", commit).Print("starting application...")
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
	user, err := session.User("@me")
	if err != nil {
		logrus.WithError(err).Fatal("could not get information about bot user")
	}
	resolveHandler := &discord1111resolver.ResolveHandler{
		DNSClient:&dns.Client{},
		DiscordBotUser:user,
	}
	resolveHandler.Initialize()
	session.AddHandler(resolveHandler.Handle)
	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}
