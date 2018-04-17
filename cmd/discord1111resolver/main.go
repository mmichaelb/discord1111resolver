package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/miekg/dns"
	"github.com/mmichaelb/discord1111resolver/pkg"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var discordbotsUpdateURL = "https://discordbots.org/api/bots/%s/stats"

var applicationName, version, branch, commit string

var discordToken string
var discordbotsToken string
var discordbotsUpdateInterval time.Duration
var stringLevel string

func main() {
	logrus.WithField("name", applicationName).WithField("version", version).WithField("branch", branch).WithField("commit", commit).Print("starting application...")
	flag.StringVar(&stringLevel, "level", "info", "The logging level which should be used for log outputs.")
	flag.StringVar(&discordToken, "token", "", "The Discord Bot token which should be used to authenticate with the Discord API.")
	flag.StringVar(&discordbotsToken, "discordbotstoken", "", "The discordbots.org token which is used to update the bot's stats.")
	flag.DurationVar(&discordbotsUpdateInterval, "discordbotsinterval", time.Minute*30, "The interval in which an update is sent to the discordbots.org API.")
	flag.Parse()
	// parse level from user input
	level, err :=  logrus.ParseLevel(stringLevel)
	if err != nil {
		logrus.WithError(err).WithField("user-level", stringLevel).Fatal("could not find level")
	}
	// set logrus level
	logrus.SetLevel(level)
	// check if a discord API token is available
	if discordToken == "" {
		logrus.Warn("no Discord API token provided")
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
	// check if a discordbots.org API token is available
	var discordbotsUpdateExitChan chan interface{}
	if discordbotsToken != "" {
		discordbotsUpdateExitChan = make(chan interface{})
		discordbotsUpdateURL = fmt.Sprintf(discordbotsUpdateURL, user.ID)
		logrus.Info("running discordbots.org update thread in background...")
		discordbotsUpdater := &discordbotsUpdater{
			discordSession: session,
		}
		go discordbotsUpdater.runDiscordbotsUpdater(session, discordbotsUpdateExitChan)
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
	if discordbotsUpdateExitChan != nil {
		discordbotsUpdateExitChan <- struct{}{}
	}
	if err := session.Close(); err != nil {
		logrus.WithError(err).Warn("could not close discord session")
	}
	logrus.Info("bye")
	os.Exit(0)
}

type discordbotsUpdater struct {
	http.Client
	discordSession *discordgo.Session
}

func (discordbotsUpdater *discordbotsUpdater) runDiscordbotsUpdater(session *discordgo.Session, exitChannel chan interface{}) {
	for {
		select {
		case <-exitChannel:
			return
		case <-time.After(discordbotsUpdateInterval):
			discordbotsUpdater.updateDiscordbotsAPI(session)
			break
		}
	}
}

func (discordbotsUpdater *discordbotsUpdater) updateDiscordbotsAPI(session *discordgo.Session) {
	var guildCount int
	var afterID string
	for {
		// request user guilds
		userGuilds, err := session.UserGuilds(100, "", afterID)
		if err != nil {
			logrus.WithError(err).Warn("could not request guild list")
		}
		// increase total guild count
		guildCount += len(userGuilds)
		if len(userGuilds) < 100 {
			break
		}
		// set new afterID to request the new list
		afterID = userGuilds[99].ID
	}
	updateData := struct {
		ServerCount int `json:"server_count"`
	}{
		ServerCount: guildCount,
	}
	marshalBytes, err := json.Marshal(updateData)
	if err != nil {
		logrus.WithError(err).Warn("could not marshal discordbots.org update data")
		return
	}
	request, err := http.NewRequest(http.MethodPost, discordbotsUpdateURL, bytes.NewReader(marshalBytes))
	if err != nil {
		logrus.WithError(err).Warn("could not create http request for the discordbots.org API")
		return
	}
	// set authorization and content type
	request.Header["Authorization"] = []string{discordbotsToken}
	request.Header["Content-Type"] = []string{"application/json"}
	resp, err := discordbotsUpdater.Do(request)
	if err != nil {
		logrus.WithError(err).Warn("an error occurred while updating the discordbots.org data")
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		logrus.WithField("http-status-code", resp.StatusCode).Warn("received an unexpected http status code")
	}
}
