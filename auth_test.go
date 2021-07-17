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

func TestAuthLogUninteresting(t *testing.T) {
	cfg := &config{}
	cfg.Auth.NoAuth = false
	callback := cfg.getAuthLogCallback()
	logBuffer := setupLogBuffer(t, cfg)
	callback(mockConnContext{}, "password", nil)
	logs := logBuffer.String()
	expectedLogs := ``
	if logs != expectedLogs {
		t.Errorf("logs=%v, want %v", string(logs), expectedLogs)
	}
}

func TestNoAuthFail(t *testing.T) {
	cfg := &config{}
	cfg.Auth.NoAuth = false
	callback := cfg.getAuthLogCallback()
	logBuffer := setupLogBuffer(t, cfg)
	callback(mockConnContext{}, "none", errors.New(""))
	logs := logBuffer.String()
	expectedLogs := `[127.0.0.1:1234] authentication for user "root" without credentials rejected
`
	if logs != expectedLogs {
		t.Errorf("logs=%v, want %v", string(logs), expectedLogs)
	}
}
