package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"net"

	"golang.org/x/crypto/ssh"
)

type source int

const (
	client source = iota
	server
)

func (src source) String() string {
	switch src {
	case client:
		return "client"
	case server:
		return "server"
	default:
		return "unknown"
	}
}

func (src source) MarshalJSON() ([]byte, error) {
	return json.Marshal(src.String())
}

type channelLog struct {
	ChannelID int `json:"channel_id"`
}

type requestLog struct {
	Type      string `json:"type"`
	WantReply bool   `json:"want_reply"`
	Payload   string `json:"payload"`

	Accepted bool `json:"accepted"`
}

type logEntry interface {
	eventType() string
}

type globalRequestLog struct {
	requestLog

	Response string `json:"response"`
}

func (entry globalRequestLog) eventType() string {
	return "global_request"
}

type newChannelLog struct {
	Type      string `json:"type"`
	ExtraData string `json:"extra_data"`

	Accepted bool `json:"accepted"`
}
