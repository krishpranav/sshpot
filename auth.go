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

func (cfg *config) getPasswordCallback() func(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
	if !cfg.Auth.PasswordAuth.Enabled {
		return nil
	}
	return func(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
		connContext{ConnMetadata: conn, cfg: cfg}.logEvent(passwordAuthLog{
			authLog: authLog{
				User:     conn.User(),
				Accepted: authAccepted(cfg.Auth.PasswordAuth.Accepted),
			},
			Password: string(password),
		})
		if !cfg.Auth.PasswordAuth.Accepted {
			return nil, errors.New("")
		}
		return nil, nil
	}
}

func (cfg *config) getPublicKeyCallback() func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
	if !cfg.Auth.PublicKeyAuth.Enabled {
		return nil
	}
	return func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
		connContext{ConnMetadata: conn, cfg: cfg}.logEvent(publicKeyAuthLog{
			authLog: authLog{
				User:     conn.User(),
				Accepted: authAccepted(cfg.Auth.PublicKeyAuth.Accepted),
			},
			PublicKeyFingerprint: ssh.FingerprintSHA256(key),
		})
		if !cfg.Auth.PublicKeyAuth.Accepted {
			return nil, errors.New("")
		}
		return nil, nil
	}
}
