package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"
)

type logEntry interface {
	fmt.Stringer
	eventType() string
}

type authAccepted bool

func (accepted authAccepted) String() string {
	if accepted {
		return "accepted"
	}
	return "rejected"
}

type authLog struct {
	User     string       `json:"user"`
	Accepted authAccepted `json:"accepted"`
}

type noAuthLog struct {
	authLog
}

func (entry noAuthLog) String() string {
	return fmt.Sprintf("authentication for user %q without credentials %v", entry.User, entry.Accepted)
}
func (entry noAuthLog) eventType() string {
	return "no_auth"
}

type passwordAuthLog struct {
	authLog
	Password string `json:"password"`
}

func (entry passwordAuthLog) String() string {
	return fmt.Sprintf("authentication for user %q with password %q %v", entry.User, entry.Password, entry.Accepted)
}
func (entry passwordAuthLog) eventType() string {
	return "password_auth"
}

type publicKeyAuthLog struct {
	authLog
	PublicKeyFingerprint string `json:"public_key"`
}

func (entry publicKeyAuthLog) String() string {
	return fmt.Sprintf("authentication for user %q with public key %q %v", entry.User, entry.PublicKeyFingerprint, entry.Accepted)
}
func (entry publicKeyAuthLog) eventType() string {
	return "public_key_auth"
}

type keyboardInteractiveAuthLog struct {
	authLog
	Answers []string `json:"answers"`
}

func (entry keyboardInteractiveAuthLog) String() string {
	return fmt.Sprintf("authentication for user %q with keyboard interactive answers %q %v", entry.User, entry.Answers, entry.Accepted)
}
func (entry keyboardInteractiveAuthLog) eventType() string {
	return "keyboard_interactive_auth"
}

type connectionLog struct {
	ClientVersion string `json:"client_version"`
}

func (entry connectionLog) String() string {
	return fmt.Sprintf("connection with client version %q established", entry.ClientVersion)
}
func (entry connectionLog) eventType() string {
	return "connection"
}

type connectionCloseLog struct {
}

func (entry connectionCloseLog) String() string {
	return "connection closed"
}
func (entry connectionCloseLog) eventType() string {
	return "connection_close"
}

type tcpipForwardLog struct {
	Address string `json:"address"`
}

func (entry tcpipForwardLog) String() string {
	return fmt.Sprintf("TCP/IP forwarding on %v requested", entry.Address)
}
func (entry tcpipForwardLog) eventType() string {
	return "tcpip_forward"
}

type cancelTCPIPForwardLog struct {
	Address string `json:"address"`
}

func (entry cancelTCPIPForwardLog) String() string {
	return fmt.Sprintf("TCP/IP forwarding on %v canceled", entry.Address)
}
func (entry cancelTCPIPForwardLog) eventType() string {
	return "cancel_tcpip_forward"
}

type cancelTCPIPForwardLog struct {
	Address string `json:"address"`
}

func (entry cancelTCPIPForwardLog) String() string {
	return fmt.Sprintf("TCP/IP forwarding on %v canceled", entry.Address)
}
func (entry cancelTCPIPForwardLog) eventType() string {
	return "cancel_tcpip_forward"
}

type noMoreSessionsLog struct {
}

func (entry noMoreSessionsLog) String() string {
	return "rejection of further session channels requested"
}
func (entry noMoreSessionsLog) eventType() string {
	return "no_more_sessions"
}

type channelLog struct {
	ChannelID int `json:"channel_id"`
}

type sessionLog struct {
	channelLog
}

func (entry sessionLog) String() string {
	return fmt.Sprintf("[channel %v] session requested", entry.ChannelID)
}
func (entry sessionLog) eventType() string {
	return "session"
}

type sessionCloseLog struct {
	channelLog
}

func (entry sessionCloseLog) String() string {
	return fmt.Sprintf("[channel %v] closed", entry.ChannelID)
}
func (entry sessionCloseLog) eventType() string {
	return "session_close"
}

type sessionInputLog struct {
	channelLog
	Input string `json:"input"`
}
