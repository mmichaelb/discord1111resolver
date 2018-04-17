package discord1111resolver

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/miekg/dns"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/idna"
	"strconv"
	"strings"
)

const (
	dNSServer                 = "1.1.1.1:853"
	unknownResponseCodeFormat = "unknown response code (%d)"
	dNSDurationFormat         = "Got answer in %v."
	dNSAnswerValueFormat      = "%s - `%s`"
)

// dNSResponseCodeMessages contains DNS response codes and fitting error messages
var dNSResponseCodeMessages = map[int]string{
	dns.RcodeFormatError:   "Format error",
	dns.RcodeServerFailure: "Server failure",
	dns.RcodeNameError:     "Non-Existent domain",
}

var profile = idna.New() //PunyCode resolver profile

func (resolveHandler *ResolveHandler) executeDNSRequest(messageEmbed *discordgo.MessageEmbed, dNSMessageType uint16, dNSMessageTypeString string, domain string) (ok bool) {
	// encode punycode
	punycodeDomain, err := profile.ToASCII(domain)
	if err != nil {
		logrus.WithError(err).Warn("could not encode unicode to punycode")
		messageEmbed.Fields = []*discordgo.MessageEmbedField{{
			Name:   "An error occurred while decoding a punycode domain:",
			Value:  strconv.Quote(err.Error()),
			Inline: true,
		}}
		return false
	}
	// create new message instance from the parameter data
	message := &dns.Msg{
		Question: []dns.Question{{
			Name:   dns.Fqdn(punycodeDomain),
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
			Name:   "Unknown error while executing the DNS request:",
			Value:  strconv.Quote(err.Error()),
			Inline: true,
		}}
		return false
	}
	if errorMessage, dNSResponseCodeOk := validateDNSResponseCode(response.Rcode); !dNSResponseCodeOk {
		messageEmbed.Fields = []*discordgo.MessageEmbedField{{
			Name:   "The DNS server returned an non-successful response code:",
			Value:  errorMessage,
			Inline: true,
		}}
		return false
	}
	if len(response.Answer) > 0 {
		messageEmbed.Fields = make([]*discordgo.MessageEmbedField, len(response.Answer))
		for index, answer := range response.Answer {
			messageEmbed.Fields[index] = &discordgo.MessageEmbedField{
				Name:  domain,
				Value: parseDNSAnswer(answer),
			}
		}
	} else {
		messageEmbed.Fields = []*discordgo.MessageEmbedField{{
			Name:   "Could not find DNS entry for question type:",
			Value:  strconv.Quote(strings.ToUpper(dNSMessageTypeString)),
			Inline: true,
		}}
		return false
	}
	messageEmbed.Footer = &discordgo.MessageEmbedFooter{Text: fmt.Sprintf(dNSDurationFormat, duration)}
	return true
}

func parseDNSAnswer(answer dns.RR) string {
	switch answerType := interface{}(answer).(type) {
	case *dns.A:
		return fmt.Sprintf(dNSAnswerValueFormat, "A", answerType.A.String())
	case *dns.AAAA:
		return fmt.Sprintf(dNSAnswerValueFormat, "AAAA", answerType.AAAA.String())
	case *dns.CNAME:
		return fmt.Sprintf(dNSAnswerValueFormat, "CNAME", answerType.Target)
	default:
		logrus.WithField("answer-type", fmt.Sprintf("%T", answerType)).Warn("could not parse answer type")
		return "invalid answer type"
	}
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
