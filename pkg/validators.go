package discorddnsbot

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/miekg/dns"
	"github.com/sirupsen/logrus"
	"strconv"
)

const (
	maxValueLength = 16
	maxValueSuffix = "[...]"
)

func (handler *DNSRequestHandler) validateDNSMessageType(session *discordgo.Session, channelID, dnsType string) (dnsMessageType uint16, ok bool) {
	switch dnsType {
	case "A":
		dnsMessageType = dns.TypeA
		break
	case "AAAA":
		dnsMessageType = dns.TypeAAAA
		break
	default:
		var truncated bool
		dnsType, truncated = validateValueLength(dnsType)
		dnsType = strconv.Quote(dnsType)
		if truncated {
			dnsType += maxValueSuffix
		}
		if _, err := session.ChannelMessageSendEmbed(channelID, &discordgo.MessageEmbed{
			Title: BotName,
			Color: errorEmbedColor,
			Footer: &discordgo.MessageEmbedFooter{
				Text: handler.getSyntax(),
			},
			Fields: []*discordgo.MessageEmbedField{{
				Name:   "Invalid DNS question type",
				Value:  dnsType,
				Inline: true,
			}},
		}); err != nil {
			logrus.WithError(err).WithField("channel-id", channelID).Warn("could not send discord invalid dns record type message")
		}
		ok = false
		return
	}
	ok = true
	return
}

func (handler *DNSRequestHandler) validateDomainName(session *discordgo.Session, channelID, domainName string) (ok bool) {
	if _, ok = dns.IsDomainName(domainName); ok {
		return
	}
	var truncated bool
	domainName, truncated = validateValueLength(domainName)
	domainName = strconv.Quote(domainName)
	if truncated {
		domainName += maxValueSuffix
	}
	if _, err := session.ChannelMessageSendEmbed(channelID, &discordgo.MessageEmbed{
		Title: BotName,
		Color: errorEmbedColor,
		Footer: &discordgo.MessageEmbedFooter{
			Text: handler.getSyntax(),
		},
		Fields: []*discordgo.MessageEmbedField{{
			Name:  "Invalid domain name",
			Value: domainName,
		}},
	}); err != nil {
		logrus.WithError(err).WithField("channel-id", channelID).Warn("could not send discord invalid domain name message")
	}
	return
}

func (handler *DNSRequestHandler) validateDNSResponseCode(session *discordgo.Session, channelID string, response *dns.Msg, domainName string) (ok bool) {
	var truncated bool
	domainName, truncated = validateValueLength(domainName)
	domainName = strconv.Quote(domainName)
	if truncated {
		domainName += maxValueSuffix
	}
	var errorMessage string
	switch response.Rcode {
	case dns.RcodeSuccess:
		return true
	case dns.RcodeFormatError:
		errorMessage = "Format error"
		break
	case dns.RcodeServerFailure:
		errorMessage = "Server failure"
		break
	case dns.RcodeNameError:
		errorMessage = "Non-Existent domain"
		break
	default:
		logrus.WithField("domain-name", domainName).WithField("response-code", response.Rcode).Warn("non-common response code from DNS server")
		errorMessage = fmt.Sprintf("non-common response code: %d", response.Rcode)
		break
	}
	if _, err := session.ChannelMessageSendEmbed(channelID, &discordgo.MessageEmbed{
		Title: BotName,
		Color: errorEmbedColor,
		Footer: &discordgo.MessageEmbedFooter{
			Text: handler.getSyntax(),
		},
		Fields: []*discordgo.MessageEmbedField{{
			Name:  "DNS server returned an error response code",
			Value: errorMessage,
		}},
	}); err != nil {
		logrus.WithError(err).WithField("channel-id", channelID).Warn("could not send discord invalid dns response code message")
	}
	return
}

func validateValueLength(value string) (string, bool) {
	if len(value) > maxValueLength {
		value = value[:maxValueLength-len(maxValueSuffix)]
		return value, true
	}
	return value, false
}
