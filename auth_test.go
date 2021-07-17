package main

import (
	"errors"
	"net"
	"reflect"
	"testing"
)

type mockConnContext struct{}

func (context mockConnContext) User() string {
	return "root"
}

func (context mockConnContext) SessionID() []byte {
	return []byte("somesession")
}

func (context mockConnContext) ClientVersion() []byte {
	return []byte("SSH-2.0-testclient")
}

func (context mockConnContext) ServerVersion() []byte {
	return []byte("SSH-2.0-testserver")
}

func (context mockConnContext) RemoteAddr() net.Addr {
	return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 1234}
}

func (context mockConnContext) LocalAddr() net.Addr {
	return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 2022}
}
