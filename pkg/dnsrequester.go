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

func (resolveHandler *ResolveHandler) executeDNSRequest(messageEmbed *discordgo.MessageEmbed, dNSMessageType uint16, dNSMessageTypeString string, domain string) (ok bool) {
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
		messageEmbed.Fields = []*discordgo.MessageEmbedField{{
			Name:   "Unknown error while executing the DNS request",
			Value:  strconv.Quote(err.Error()),
			Inline: true,
		}}
		return false
	}
	if errorMessage, dNSResponseCodeOk := validateDNSResponseCode(response.Rcode); !dNSResponseCodeOk {
		messageEmbed.Fields = []*discordgo.MessageEmbedField{{
			Name:   "The DNS server returned an non-successful response code",
			Value:  errorMessage,
			Inline: true,
		}}
		return false
	}
	if len(response.Answer) > 0 {
		messageEmbed.Fields = make([]*discordgo.MessageEmbedField, len(response.Answer))
		for index, answer := range response.Answer {
			messageEmbed.Fields[index] = &discordgo.MessageEmbedField{
				Name:  answer.Header().Name,
				Value: answer.String(),
			}
		}
	} else {
		messageEmbed.Fields = []*discordgo.MessageEmbedField{{
			Name:   "Could not find DNS entry for question type",
			Value:  strconv.Quote(strings.ToUpper(dNSMessageTypeString)),
			Inline: true,
		}}
		return false
	}
	messageEmbed.Footer = &discordgo.MessageEmbedFooter{Text: fmt.Sprintf(dNSDurationFormat, duration)}
	return true
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
