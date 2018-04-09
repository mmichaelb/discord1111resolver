package discorddnsbot

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/miekg/dns"
	"github.com/sirupsen/logrus"
	"strconv"
	"strings"
)

const DNSServer = "8.8.8.8:53"

type DNSRequestHandler struct {
	CommandPrefix string
	client        *dns.Client
}

func NewDNSRequestHandler(commandPrefix string) *DNSRequestHandler {
	if commandPrefix[len(commandPrefix)-1:] != " " {
		commandPrefix += " "
	}
	return &DNSRequestHandler{
		CommandPrefix:commandPrefix,
		client:&dns.Client{},
	}
}

// Handle should be bound to a discordgo instance in order to react to DNS request messages.
func (handler *DNSRequestHandler) Handle(session *discordgo.Session, messageCreate *discordgo.MessageCreate) {
	// check if message is of a right length
	if len(messageCreate.Content) <= len(handler.CommandPrefix) {
		return
	}
	// check if message has the right prefix
	if messageCreate.Content[:len(handler.CommandPrefix)] != handler.CommandPrefix {
		return
	}
	parameters := strings.Split(messageCreate.Content[len(handler.CommandPrefix):], " ")
	if len(parameters) != 2 {
		handler.sendSyntax(session, messageCreate.ChannelID)
		return
	}
	ok, dnsMessageType := parseDNSMessageType(parameters[0])
	if !ok {
		if _, err := session.ChannelMessageSendEmbed(messageCreate.ChannelID, &discordgo.MessageEmbed{
			Title: "DNS query bot",
			Fields: []*discordgo.MessageEmbedField{{
				Name:  "Invalid DNS message type",
				Value: strconv.Quote(parameters[0]),
			}},
			Description: handler.getSyntax(),
		}); err != nil {
			logrus.WithError(err).WithField("channel-id", messageCreate.ChannelID).Warn("could not send discord invalid dns record type message")
		}
		return
	}
	if _, ok := dns.IsDomainName(parameters[1]); !ok {
		if _, err := session.ChannelMessageSendEmbed(messageCreate.ChannelID, &discordgo.MessageEmbed{
			Title: "DNS query bot",
			Fields: []*discordgo.MessageEmbedField{{
				Name:  "Invalid domain name",
				Value: strconv.Quote(parameters[1]),
			}},
			Description: handler.getSyntax(),
		}); err != nil {
			logrus.WithError(err).WithField("channel-id", messageCreate.ChannelID).Warn("could not send discord invalid domain name message")
		}
		return
	}
	message := &dns.Msg{
		Question: []dns.Question{{
				Name:   dns.Fqdn(parameters[1]),
				Qtype:  dnsMessageType,
				Qclass: dns.ClassINET,
		}},
	}
	message.RecursionDesired = true
	response, duration, err := handler.client.Exchange(message, DNSServer)
	if err != nil {
		logrus.WithError(err).WithField("domain-name", parameters[1]).Warn("could not execute DNS request")
		if _, err := session.ChannelMessageSendEmbed(messageCreate.ChannelID, &discordgo.MessageEmbed{
			Title: "DNS query bot",
			Fields: []*discordgo.MessageEmbedField{{
				Name:  "DNS query error with domain name",
				Value: strconv.Quote(parameters[1]),
			}},
		}); err != nil {
			logrus.WithError(err).WithField("channel-id", messageCreate.ChannelID).Warn("could not send discord dns query error message")
		}
		return
	}
	if response.Rcode != dns.RcodeSuccess {
		logrus.WithField("domain-name", parameters[1]).WithField("response-code", response.Rcode).Warn("invalid response code from DNS server")
		if _, err := session.ChannelMessageSendEmbed(messageCreate.ChannelID, &discordgo.MessageEmbed{
			Title: "DNS query bot",
			Fields: []*discordgo.MessageEmbedField{{
				Name:  "DNS query error with domain name",
				Value: strconv.Quote(parameters[1]),
			}},
		}); err != nil {
			logrus.WithError(err).WithField("channel-id", messageCreate.ChannelID).Warn("could not send discord invalid dns response code message")
		}
		return
	}
	responseFields := make([]*discordgo.MessageEmbedField, len(response.Answer))
	for index, answer := range response.Answer {
		responseFields[index] = &discordgo.MessageEmbedField{
			Name:answer.Header().Name,
			Value:answer.String(),
		}
	}
	if _, err := session.ChannelMessageSendEmbed(messageCreate.ChannelID, &discordgo.MessageEmbed{
		Title: "DNS query bot",
		Fields: responseFields,
		Description:fmt.Sprintf("Got response in %v.", duration),
	}); err != nil {
		logrus.WithError(err).WithField("channel-id", messageCreate.ChannelID).Warn("could not reply to message")
	}

}

func (handler *DNSRequestHandler) getSyntax() string {
	return fmt.Sprintf("Please use this syntax: %v <A|AAA> <domain>", handler.CommandPrefix)
}

func (handler *DNSRequestHandler) sendSyntax(session *discordgo.Session, channelID string) {
	if _, err := session.ChannelMessageSend(channelID, handler.getSyntax()); err != nil {
		logrus.WithField("channel-id", channelID).WithError(err).Warn("could not send syntax message")
	}
}

func parseDNSMessageType(parameter string) (ok bool, dnsMessageType uint16) {
	switch parameter {
	case "A":
		return true, dns.TypeA
	case "AAA":
		return true, dns.TypeAAAA
	default:
		return false, 0
	}
}
