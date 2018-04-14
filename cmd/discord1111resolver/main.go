package main

import (
	"flag"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/miekg/dns"
	"github.com/mmichaelb/discord1111resolver/pkg"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
)

var applicationName, version, branch, commit string

var discordToken string
var discordbotsToken string

func main() {
	logrus.WithField("name", applicationName).WithField("version", version).WithField("branch", branch).WithField("commit", commit).Print("starting application...")
	flag.StringVar(&discordToken, "token", "", "The Discord Bot token which should be used to authenticate with the Discord API.")
	flag.StringVar(&discordbotsToken, "discordbotstoken", "", "The discordbots.org token which is used to update the bot's stats.")
	flag.Parse()
	// check if a discord API token is available
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
		DNSClient: &dns.Client{
			Net: "tcp-tls", // enable DNS over TLS
		},
		DiscordBotUser: user,
	}
	resolveHandler.Initialize()
	session.AddHandler(resolveHandler.Handle)
	// Wait here until CTRL-C or other term signal is received.
	logrus.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
	if err := session.Close(); err != nil {
		logrus.WithError(err).Warn("could not close discord session")
	}
	logrus.Info("bye")
	os.Exit(0)
}
