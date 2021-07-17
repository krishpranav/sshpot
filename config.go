package main

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"

	"golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v2"
)

type serverConfig struct {
	ListenAddress string   `yaml:"listen_address"`
	HostKeys      []string `yaml:"host_keys"`
}

type loggingConfig struct {
	File       string `yaml:"file"`
	JSON       bool   `yaml:"json"`
	Timestamps bool   `yaml:"timestamps"`
	Debug      bool   `yaml:"debug"`
}

type commonAuthConfig struct {
	Enabled  bool `yaml:"enabled"`
	Accepted bool `yaml:"accepted"`
}

type keyboardInteractiveAuthQuestion struct {
	Text string `yaml:"text"`
	Echo bool   `yaml:"echo"`
}

type keyboardInteractiveAuthConfig struct {
	commonAuthConfig `yaml:",inline"`
	Instruction      string                            `yaml:"instruction"`
	Questions        []keyboardInteractiveAuthQuestion `yaml:"questions"`
}

type authConfig struct {
	MaxTries                int                           `yaml:"max_tries"`
	NoAuth                  bool                          `yaml:"no_auth"`
	PasswordAuth            commonAuthConfig              `yaml:"password_auth"`
	PublicKeyAuth           commonAuthConfig              `yaml:"public_key_auth"`
	KeyboardInteractiveAuth keyboardInteractiveAuthConfig `yaml:"keyboard_interactive_auth"`
}

type sshProtoConfig struct {
	Version        string   `yaml:"version"`
	Banner         string   `yaml:"banner"`
	RekeyThreshold uint64   `yaml:"rekey_threshold"`
	KeyExchanges   []string `yaml:"key_exchanges"`
	Ciphers        []string `yaml:"ciphers"`
	MACs           []string `yaml:"macs"`
}

type config struct {
	Server   serverConfig   `yaml:"server"`
	Logging  loggingConfig  `yaml:"logging"`
	Auth     authConfig     `yaml:"auth"`
	SSHProto sshProtoConfig `yaml:"ssh_proto"`

	parsedHostKeys []ssh.Signer
	sshConfig      *ssh.ServerConfig
	logFileHandle  io.WriteCloser
}

func getDefaultConfig() *config {
	cfg := &config{}
	cfg.Server.ListenAddress = "127.0.0.1:2022"
	cfg.Logging.Timestamps = true
	cfg.Auth.PasswordAuth.Enabled = true
	cfg.Auth.PasswordAuth.Accepted = true
	cfg.Auth.PublicKeyAuth.Enabled = true
	cfg.SSHProto.Version = "SSH-2.0-sshesame"
	cfg.SSHProto.Banner = "This is an SSH honeypot. Everything is logged and monitored."
	return cfg
}

type keySignature int

const (
	rsa_key keySignature = iota
	ecdsa_key
	ed25519_key
)

func (signature keySignature) String() string {
	switch signature {
	case rsa_key:
		return "rsa"
	case ecdsa_key:
		return "ecdsa"
	case ed25519_key:
		return "ed25519"
	default:
		return "unknown"
	}
}
