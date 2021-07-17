package main

import (
	"errors"
	"math/rand"
	"net"
	"strconv"

	"golang.org/x/crypto/ssh"
)

type globalRequestPayload interface {
	reply() []byte
	logEntry() logEntry
}

type globalRequestPayloadParser func(data []byte) (globalRequestPayload, error)

type channelRequestPayload interface {
	reply() []byte
	logEntry(channelID int) logEntry
}

type channelRequestPayloadParser func(data []byte) (channelRequestPayload, error)

type tcpipRequest struct {
	Address string
	Port    uint32
}

func (request tcpipRequest) reply() []byte {
	if request.Port != 0 {
		return nil
	}
	return ssh.Marshal(struct{ port uint32 }{uint32(rand.Intn(65536-1024) + 1024)})
}
func (request tcpipRequest) logEntry() logEntry {
	return tcpipForwardLog{
		Address: net.JoinHostPort(request.Address, strconv.Itoa(int(request.Port))),
	}
}

func (request cancelTCPIPRequest) reply() []byte {
	return nil
}
func (request cancelTCPIPRequest) logEntry() logEntry {
	return cancelTCPIPForwardLog{
		Address: net.JoinHostPort(request.Address, strconv.Itoa(int(request.Port))),
	}
}

type noMoreSessionsRequest struct {
}

func (request noMoreSessionsRequest) reply() []byte {
	return nil
}
func (request noMoreSessionsRequest) logEntry() logEntry {
	return noMoreSessionsLog{}
}

var globalRequestPayloads = map[string]globalRequestPayloadParser{
	"tcpip-forward": func(data []byte) (globalRequestPayload, error) {
		payload := &tcpipRequest{}
		if err := ssh.Unmarshal(data, payload); err != nil {
			return nil, err
		}
		return payload, nil
	},
	"cancel-tcpip-forward": func(data []byte) (globalRequestPayload, error) {
		payload := &cancelTCPIPRequest{}
		if err := ssh.Unmarshal(data, payload); err != nil {
			return nil, err
		}
		return payload, nil
	},
	"no-more-sessions@openssh.com": func(data []byte) (globalRequestPayload, error) {
		if len(data) != 0 {
			return nil, errors.New("invalid request payload")
		}
		return &noMoreSessionsRequest{}, nil
	},
}
