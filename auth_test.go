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
	return []byte("session-1")
}

func (context mockConnContext) ClientVersion() []byte {
	return []byte("SSH-2.0-testclient")
}

func (context mockConnContext) ServerVersion() []byte {
	return []byte("SSH-2.0-testserver")
}
