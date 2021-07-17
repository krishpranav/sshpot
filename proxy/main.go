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

func streamHeader(reader io.Reader) <-chan string {
	input := make(chan string)
	go func() {
		defer close(input)
		buffer := make([]byte, 256)
		for {
			n, err := reader.Read(buffer)
			if n > 0 {
				input <- string(buffer[:n])
			}
			if err != nil {
				if err != io.EOF {
					panic(err)
				}
				return
			}
		}
	}()
	return input
}

func handleChannel(channelID int, clientChannel ssh.Channel, clientRequests <-chan *ssh.Request, serverChannel ssh.Channel, serverRequests <-chan *ssh.Request) {
	clientInputStream := streamReader(clientChannel)
	serverInputStream := streamReader(serverChannel)
	serverErrorStream := streamReader(serverChannel.Stderr())

	for clientInputStream != nil || clientRequests != nil || serverInputStream != nil || serverRequests != nil {
		select {
		case clientInput, ok := <-clientInputStream:
			if !ok {
				if serverInputStream != nil {
					logEvent(channelEOFLog{
						channelLog: channelLog{
							ChannelID: channelID,
						},
					}, client)
					if err := serverChannel.CloseWrite(); err != nil {
						panic(err)
					}
				}
				clientInputStream = nil
				continue
			}
			logEvent(channelDataLog{
				channelLog: channelLog{
					ChannelID: channelID,
				},
				Data: clientInput,
			}, client)
			if _, err := serverChannel.Write([]byte(clientInput)); err != nil {
				panic(err)
			}
		case clientRequest, ok := <-clientRequests:
			if !ok {
				if serverRequests != nil {
					logEvent(channelCloseLog{
						channelLog: channelLog{
							ChannelID: channelID,
						},
					}, client)
					if err := serverChannel.Close(); err != nil {
						panic(err)
					}
				}
				clientRequests = nil
				continue
			}
			accepted, err := serverChannel.SendRequest(clientRequest.Type, clientRequest.WantReply, clientRequest.Payload)
			if err != nil {
				panic(err)
			}
			logEvent(channelRequestLog{
				channelLog: channelLog{
					ChannelID: channelID,
				},
				requestLog: requestLog{
					Type:      clientRequest.Type,
					WantReply: clientRequest.WantReply,
					Payload:   base64.RawStdEncoding.EncodeToString(clientRequest.Payload),
					Accepted:  accepted,
				},
			}, client)
			if clientRequest.WantReply {
				if err := clientRequest.Reply(accepted, nil); err != nil {
					panic(err)
				}
			}
		case serverInput, ok := <-serverInputStream:
			if !ok {
				if clientInputStream != nil {
					logEvent(channelEOFLog{
						channelLog: channelLog{
							ChannelID: channelID,
						},
					}, server)
					if err := clientChannel.CloseWrite(); err != nil {
						panic(err)
					}
				}
				serverInputStream = nil
				continue
			}
			logEvent(channelDataLog{
				channelLog: channelLog{
					ChannelID: channelID,
				},
				Data: serverInput,
			}, server)
			if _, err := clientChannel.Write([]byte(serverInput)); err != nil {
				panic(err)
			}
		case serverError, ok := <-serverErrorStream:
			if !ok {
				serverErrorStream = nil
				continue
			}
			logEvent(channelErrorLog{
				channelLog: channelLog{
					ChannelID: channelID,
				},
				Data: serverError,
			}, server)
			if _, err := clientChannel.Stderr().Write([]byte(serverError)); err != nil {
				panic(err)
			}
		case serverRequest, ok := <-serverRequests:
			if !ok {
				if clientRequests != nil {
					logEvent(channelCloseLog{
						channelLog: channelLog{
							ChannelID: channelID,
						},
					}, server)
					if err := clientChannel.Close(); err != nil {
						panic(err)
					}
				}
				serverRequests = nil
				continue
			}
			accepted, err := clientChannel.SendRequest(serverRequest.Type, serverRequest.WantReply, serverRequest.Payload)
			if err != nil {
				panic(err)
			}
			logEvent(channelRequestLog{
				channelLog: channelLog{
					ChannelID: channelID,
				},
				requestLog: requestLog{
					Type:      serverRequest.Type,
					WantReply: serverRequest.WantReply,
					Payload:   base64.RawStdEncoding.EncodeToString(serverRequest.Payload),
					Accepted:  accepted,
				},
			}, server)
			if serverRequest.WantReply {
				if err := serverRequest.Reply(accepted, nil); err != nil {
					panic(err)
				}
			}
		}
	}
}
