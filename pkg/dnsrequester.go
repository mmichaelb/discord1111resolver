package discord1111resolver

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/miekg/dns"
	"github.com/sirupsen/logrus"
	"strconv"
	"strings"
)

const (
	dNSServer                 = "1.1.1.1:853"
	unknownResponseCodeFormat = "unknown response code (%d)"
	dNSDurationFormat         = "Got answer in %v."
)

// dNSResponseCodeMessages contains DNS response codes and fitting error messages
var dNSResponseCodeMessages = map[int]string{
	dns.RcodeFormatError:   "Format error",
	dns.RcodeServerFailure: "Server failure",
	dns.RcodeNameError:     "Non-Existent domain",
}

func (resolveHandler *ResolveHandler) executeDNSRequest(session *discordgo.Session, messageCreate *discordgo.MessageCreate, dNSMessageType uint16, dNSMessageTypeString string, domain string) (responseFields []*discordgo.MessageEmbedField, ok bool) {
	// create new message instance from the parameter data
	message := &dns.Msg{
		Question: []dns.Question{{
			Name:   dns.Fqdn(domain),
			Qtype:  dNSMessageType,
			Qclass: dns.ClassINET,
		}},
	}
	message.RecursionDesired = true
	// execute DNS request
	response, duration, err := resolveHandler.DNSClient.Exchange(message, dNSServer)
	if err != nil {
		logrus.WithError(err).Warn("could not execute DNS request")
		return []*discordgo.MessageEmbedField{{
			Name:   "Unknown error while executing the DNS request",
			Value:  strconv.Quote(err.Error()),
			Inline: true,
		}}, false
	}
	if errorMessage, dNSResponseCodeOk := validateDNSResponseCode(response.Rcode); !dNSResponseCodeOk {
		return []*discordgo.MessageEmbedField{{
			Name:   "The DNS server returned an non-successful response code",
			Value:  errorMessage,
			Inline: true,
		}}, false
	}
	if len(response.Answer) > 0 {
		responseFields = make([]*discordgo.MessageEmbedField, len(response.Answer))
		for index, answer := range response.Answer {
			responseFields[index] = &discordgo.MessageEmbedField{
				Name:  answer.Header().Name,
				Value: answer.String(),
			}
		}
	} else {
		return []*discordgo.MessageEmbedField{{
			Name:   "Could not find DNS entry for question type",
			Value:  strconv.Quote(strings.ToUpper(dNSMessageTypeString)),
			Inline: true,
		}}, false
	}
	_, err = session.ChannelMessageSendEmbed(messageCreate.ChannelID, &discordgo.MessageEmbed{
		Title:  resolveHandler.DiscordBotUser.Username,
		Color:  embedSuccessColor,
		Fields: responseFields,
		Footer: &discordgo.MessageEmbedFooter{Text: fmt.Sprintf(dNSDurationFormat, duration)},
	})
	if err != nil {
		logrus.WithError(err).WithField("channel-id", messageCreate.ChannelID).Warn("could not send dns response message to discord")
	}
	return nil, true
}

func validateDNSResponseCode(dNSResponseCode int) (errorMessage string, ok bool) {
	if dNSResponseCode == dns.RcodeSuccess {
		return "", true
	}
	var errorMessageFound bool
	if errorMessage, errorMessageFound = dNSResponseCodeMessages[dNSResponseCode]; !errorMessageFound {
		errorMessage = fmt.Sprintf(unknownResponseCodeFormat, dNSResponseCode)
	}
	return
}
