package main

import (
	"net"
	"sync"

	"golang.org/x/crypto/ssh"
)

type connContext struct {
	ssh.ConnMetadata
	cfg            *config
	noMoreSessions bool
}

type channelContext struct {
	connContext
	channelID int
}
