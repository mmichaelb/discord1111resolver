package discorddnsbot

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/miekg/dns"
	"github.com/sirupsen/logrus"
	"strconv"
	"strings"
	"time"
)

const (
	BotName           = "DNS-Bot"
	DNSServer         = "1.1.1.1:53"
	errorEmbedColor   = 16007990
	errorSuccessColor = 5025616
)

type DNSRequestHandler struct {
	CommandPrefix string
	client        *dns.Client
}

// NewDNSRequestHandler creates a new instance of the DNSRequestHandler and uses a default DNS client.
func NewDNSRequestHandler(commandPrefix string) *DNSRequestHandler {
	return &DNSRequestHandler{
		CommandPrefix: commandPrefix,
		client:        &dns.Client{},
	}
}

// Handle should be bound to a discordgo instance in order to react to DNS request messages.
func (handler *DNSRequestHandler) Handle(session *discordgo.Session, messageCreate *discordgo.MessageCreate) {
	// check if message is of a right length
	if len(messageCreate.Content) < len(handler.CommandPrefix) {
		return
	}
	// check if message has the right prefix
	if messageCreate.Content[:len(handler.CommandPrefix)] != handler.CommandPrefix {
		return
	}
	parameters := strings.Split(messageCreate.Content, " ")[1:]
	if len(parameters) != 2 {
		handler.sendSyntax(session, messageCreate.ChannelID)
		return
	}
	// check if the provided dns message type is valid (A or AAA)
	dnsMessageType, ok := handler.validateDNSMessageType(session, messageCreate.ChannelID, parameters[0])
	if !ok {
		return
	}
	// check if the provided domain name is valid
	ok = handler.validateDomainName(session, messageCreate.ChannelID, parameters[1])
	if !ok {
		return
	}
	// create new message instance
	message := &dns.Msg{
		Question: []dns.Question{{
			Name:   dns.Fqdn(parameters[1]),
			Qtype:  dnsMessageType,
			Qclass: dns.ClassINET,
		}},
	}
	message.RecursionDesired = true
	// execute DNS request
	response, duration, ok := executeDNSRequest(handler, message, parameters[1], session, messageCreate)
	if !ok {
		return
	}
	// validate the response code if it is a valid one/successful one
	ok = handler.validateDomainName(session, messageCreate.ChannelID, parameters[1])
	if !ok {
		return
	}
	sendDNSResults(response, session, messageCreate, duration)
}

func sendDNSResults(response *dns.Msg, session *discordgo.Session, messageCreate *discordgo.MessageCreate, duration time.Duration) {
	responseFields := make([]*discordgo.MessageEmbedField, len(response.Answer))
	for index, answer := range response.Answer {
		responseFields[index] = &discordgo.MessageEmbedField{
			Name:  answer.Header().Name,
			Value: answer.String(),
		}
	}
	if _, err := session.ChannelMessageSendEmbed(messageCreate.ChannelID, &discordgo.MessageEmbed{
		Title:  BotName,
		Color:  errorSuccessColor,
		Fields: responseFields,
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Duration of query: %v", duration),
		},
	}); err != nil {
		logrus.WithError(err).WithField("channel-id", messageCreate.ChannelID).Warn("could not reply to message")
	}
}

func executeDNSRequest(handler *DNSRequestHandler, message *dns.Msg, domainName string, session *discordgo.Session, messageCreate *discordgo.MessageCreate) (*dns.Msg, time.Duration, bool) {
	response, duration, err := handler.client.Exchange(message, DNSServer)
	var truncated bool
	domainName, truncated = validateValueLength(domainName)
	domainName = strconv.Quote(domainName)
	if truncated {
		domainName += maxValueSuffix
	}
	if message == nil && err != nil {
		logrus.WithError(err).WithField("domain-name", domainName).Warn("could not execute DNS request")
		if _, err := session.ChannelMessageSendEmbed(messageCreate.ChannelID, &discordgo.MessageEmbed{
			Title: BotName,
			Fields: []*discordgo.MessageEmbedField{{
				Name:  "an error occurred while querying the DNS request for domain name",
				Value: strconv.Quote(domainName),
			}},
		}); err != nil {
			logrus.WithError(err).WithField("channel-id", messageCreate.ChannelID).Warn("could not send discord dns query error message")
		}
		return response, duration, false
	}
	return response, duration, true
}

func (handler *DNSRequestHandler) getSyntax() string {
	return fmt.Sprintf("Syntax: %v <A|AAAA> <domain>", handler.CommandPrefix)
}

func (handler *DNSRequestHandler) sendSyntax(session *discordgo.Session, channelID string) {
	if _, err := session.ChannelMessageSendEmbed(channelID, &discordgo.MessageEmbed{
		Title: BotName,
		Color: errorEmbedColor,
		Footer: &discordgo.MessageEmbedFooter{
			Text: handler.getSyntax(),
		},
	}); err != nil {
		logrus.WithError(err).WithField("channel-id", channelID).Warn("could not send discord syntax message")
	}
}
