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
