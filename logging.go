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
