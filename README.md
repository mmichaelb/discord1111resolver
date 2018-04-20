![1.1.1.1 DNS service](https://i.imgur.com/69fFwN9.png)
# discord1111resolver (1111Resolver) [![Discord Bots](https://discordbots.org/api/widget/status/432969981366501396.svg)](https://discordbots.org/bot/432969981366501396) [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT) [![stability-beta](https://img.shields.io/badge/stability-beta-33bbff.svg)](https://github.com/mkenney/software-guides/blob/master/STABILITY-BADGES.md#beta) [![GoDoc](https://godoc.org/github.com/mmichaelb/discord1111resolver?status.svg)](https://godoc.org/github.com/mmichaelb/discord1111resolver) [![Go Report Card](https://goreportcard.com/badge/github.com/mmichaelb/discord1111resolver)](https://goreportcard.com/report/github.com/mmichaelb/discord1111resolver)
Discord bot written for fun that returns DNS queries from the 1.1.1.1 DNS service. The structure is oriented to the 
[Twitter 1111Resolver bot](https://twitter.com/1111Resolver).

## 1.1.1.1
Cloudflare and APNIC provide a fast and reliable DNS service. It is not only (one of) the fastest DNS services, but also
 attaches great importance to privacy. For more information, please visit [https://1.1.1.1/](https://1.1.1.1/).

## Server integration
If you are a Discord server owner and want to integrate the 1111Resolver you can checkout the discordbots.org site. The 
site provides an invite link as well as a status panel:

[![Discord Bots](https://discordbots.org/api/widget/432969981366501396.svg)](https://discordbots.org/bot/432969981366501396)

## Usage
As said before, this bot does not have much functionality yet but feel free to suggest new features by opening an issue 
at the [issues tab](https://github.com/mmichaelb/discord1111resolver/issues). The basic functionality can be described 
as follows:
```
@1111Resolver <A|AAAA|CNAME> <domain name>
```
An example of the usage would be:
```
@1111Resolver AAAA discordbots.org
```

*Please note that this bot is not associated with Cloudflare or APNIC.*
