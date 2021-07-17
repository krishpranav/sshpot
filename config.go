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
