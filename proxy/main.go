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

func (entry newChannelLog) eventType() string {
	return "new_channel"
}

type channelRequestLog struct {
	channelLog
	requestLog
}

func (entry channelRequestLog) eventType() string {
	return "channel_request"
}

type channelDataLog struct {
	channelLog
	Data string `json:"data"`
}

func (entry channelDataLog) eventType() string {
	return "channel_data"
}

type channelErrorLog struct {
	channelLog
	Data string `json:"data"`
}

func (entry channelErrorLog) eventType() string {
	return "channel_error"
}

type channelEOFLog struct {
	channelLog
}

func (entry channelEOFLog) eventType() string {
	return "channel_eof"
}

type channelCloseLog struct {
	channelLog
}

func (entry channelCloseLog) eventType() string {
	return "channel_close"
}

type connectionCloseLog struct{}

func (entry connectionCloseLog) eventType() string {
	return "connection_close"
}

func logEvent(entry logEntry, src source) {
	jsonBytes, err := json.Marshal(struct {
		Source    string   `json:"source"`
		EventType string   `json:"event_type"`
		Event     logEntry `json:"event"`
	}{
		Source:    src.String(),
		EventType: entry.eventType(),
		Event:     entry,
	})
	if err != nil {
		panic(err)
	}
	log.Printf("%s", jsonBytes)
}
