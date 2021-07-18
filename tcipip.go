package main

import (
	"bufio"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"strconv"

	"golang.org/x/crypto/ssh"
)

type tcpipServer interface {
	server(channel ssh.Channel, input chan<- string) error
}

var servers = map[uint32]tcpipServer{
	80: httpServer{},
}

type tcpipChannelData struct {
	Address           string
	Port              uint32
	OriginatorAddress string
	OrignatorPort     uint32
}
