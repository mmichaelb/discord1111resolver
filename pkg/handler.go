package discord1111resolver

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/miekg/dns"
	"github.com/sirupsen/logrus"
	"regexp"
	"strconv"
	"strings"
)

// multipleSpaceRegex is used to trim a bot mention from Discord users.
var multipleSpaceRegex = regexp.MustCompile(`\s+`)

// allowedDNSMessageTypes contains all allowed DNS query message types (e.g. A or AAAA).
var allowedDNSMessageTypes = map[string]uint16{
	"A":     dns.TypeA,
	"AAAA":  dns.TypeAAAA,
	"CNAME": dns.TypeCNAME,
}

const (
	// maximumValueLength is the maximum length of a discordgo Field value.
	maximumValueLength = 1024
	// mentionFormat is used to check if it is a valid mention.
	mentionFormat = "<@%s>"
	// syntaxFormat is used to hand out a valid syntax to the Discord users.
	syntaxFormat = "@%s <%s> <domain>"
	// embedErrorColor is the color used for embeds which display errors/invalid formats.
	embedErrorColor = 16007990
	// embedSuccessColor is the color used for embeds which display a successful DNS response.
	embedSuccessColor = 5025616
	// baseColor is the main color used on the 1.1.1.1 website.
	baseColor = 14385742
	// baseURL is the url where more information about the DNS service can be found.
	baseURL = "https://1.1.1.1/"
	// embedTitle is the title which is used for every sent message embed.
	embedTitle = "1.1.1.1 DNS service"
	// botDescription is the description which is sent if the bot gets tagged.
	botDescription = "Cloudflare and APNIC offer a fast and secure DNS service which also cares about your privacy.\n" +
		"This bot allows you to interact with it and execute simple requests."
)

// ResolveHandler is used to handle DNS query requests by Discord users. Its Handle method should be bound to a
// discordgo session instance.
type ResolveHandler struct {
	// DiscordBotUser contains information about the Discord bot user resolved via:
	//
	//  var session *discordgo.Session
	//	// assign session
	//	user, err := session.User("@me")
	//
	// This allows the handler to correctly react to tags in order to fulfill its function as a DNS resolver.
	DiscordBotUser *discordgo.User
	// DNSClient is an instance of the miekg dns client.
	DNSClient *dns.Client
	// mentionString contains a string with the format <@DISCORD-ID> to detect request messages.
	mentionString string
	// syntax contains a string which represents the syntax used to execute DNS queries.
	syntax string
}

// Initialize sets basic internal values of the ResolveHandler instance and has to be called before binding the Handle
// function.
func (resolveHandler *ResolveHandler) Initialize() {
	resolveHandler.mentionString = fmt.Sprintf(mentionFormat, resolveHandler.DiscordBotUser.ID)
	// wrap allowedDNSMessageTypes to a string slice
	availableDNSMessageTypes := make([]string, len(allowedDNSMessageTypes))
	count := 0
	for typeName := range allowedDNSMessageTypes {
		availableDNSMessageTypes[count] = typeName
		count++
	}
	resolveHandler.syntax = fmt.Sprintf(syntaxFormat, resolveHandler.DiscordBotUser.Username, strings.Join(availableDNSMessageTypes, "|"))
}

func (resolveHandler *ResolveHandler) Handle(session *discordgo.Session, messageCreate *discordgo.MessageCreate) {
	// check if the message is not from the bot itself
	if messageCreate.Author.ID == resolveHandler.DiscordBotUser.ID {
		return
	}
	// check if the message begins with a mention
	if !strings.HasPrefix(messageCreate.Content, resolveHandler.mentionString) {
		return
	}
	// replace multiple spaces with one
	trimmedContent := multipleSpaceRegex.ReplaceAllString(messageCreate.Content, " ")
	// split whole command on every space
	commandSplit := strings.Split(trimmedContent, " ")
	// pre-declare all fields to allow a goto statement
	var fields []*discordgo.MessageEmbedField
	var ok bool
	var params []string
	messageEmbed := &discordgo.MessageEmbed{
		Title: embedTitle,
		URL:   baseURL,
		Color: baseColor,
	}
	if len(commandSplit) != 3 {
		goto syntaxCheck
	}
	// initiate params (everything after "<@DISCORD-ID> "
	params = commandSplit[1:]
	// handle bot mention
	ok = resolveHandler.handleMention(messageEmbed, params)
	// check result
	if ok {
		messageEmbed.Color = embedSuccessColor
	} else {
		messageEmbed.Color = embedErrorColor
	}
syntaxCheck:
	var fieldsNotSet = fields == nil || len(fields) == 0
	if fieldsNotSet {
		messageEmbed.Description = botDescription
	}
	if !ok || fieldsNotSet {
		messageEmbed.Footer = &discordgo.MessageEmbedFooter{Text: resolveHandler.syntax}
	}
	_, err := session.ChannelMessageSendEmbed(messageCreate.ChannelID, messageEmbed)
	if err != nil {
		logrus.WithError(err).WithField("channel-id", messageCreate.ChannelID).Warn("could not send discord message")
	}
}

// handleMention is an internal function which is called if the message starts with "<@DISCORD-ID> ". It returns whether
// the execution was a success and if not, which fields should be printed withing the error message.
func (resolveHandler *ResolveHandler) handleMention(messageEmbed *discordgo.MessageEmbed, params []string) (ok bool) {
	// check params length
	if len(params) != 2 {
		return false
	}
	// validate the DNS message type parameter
	var messageType uint16
	messageTypeString := params[0]
	messageType, ok = validateDNSMessageType(messageTypeString)
	if !ok {
		trimDiscordFieldValue(&messageTypeString)
		// the user specified an invalid DNS message type
		messageEmbed.Fields = []*discordgo.MessageEmbedField{{
			Name:   "Invalid DNS message type",
			Value:  strconv.Quote(messageTypeString),
			Inline: true,
		}}
		return false
	}
	// validate the domain name
	domainName := params[1]
	_, ok = dns.IsDomainName(domainName)
	if !ok {
		trimDiscordFieldValue(&domainName)
		// the user specified an invalid domain name
		messageEmbed.Fields = []*discordgo.MessageEmbedField{{
			Name:   "Invalid domain name",
			Value:  strconv.Quote(domainName),
			Inline: true,
		}}
		return false
	}
	return resolveHandler.executeDNSRequest(messageEmbed, messageType, messageTypeString, domainName)
}

func validateDNSMessageType(messageTypeString string) (messageType uint16, ok bool) {
	messageTypeString = strings.ToUpper(messageTypeString)
	messageType, ok = allowedDNSMessageTypes[messageTypeString]
	return
}

func trimDiscordFieldValue(value *string) {
	if len(*value) > maximumValueLength {
		*value = (*value)[:maximumValueLength-3] + "..."
	}
}
