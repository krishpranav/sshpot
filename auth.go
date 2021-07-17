package main

import (
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/ssh"
)

func (cfg *config) getAuthLogCallback() func(conn ssh.ConnMetadata, method string, err error) {
	return func(conn ssh.ConnMetadata, method string, err error) {
		if method == "none" {
			connContext{ConnMetadata: conn, cfg: cfg}.logEvent(noAuthLog{authLog: authLog{
				User:     conn.User(),
				Accepted: err == nil,
			}})
		}
	}
}
