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

func generateKey(dataDir string, signature keySignature) (string, error) {
	keyFile := path.Join(dataDir, fmt.Sprintf("host_%v_key", signature))
	if _, err := os.Stat(keyFile); err == nil {
		return keyFile, nil
	} else if !os.IsNotExist(err) {
		return "", err
	}
	infoLogger.Printf("Host key %q not found, generating it", keyFile)
	if _, err := os.Stat(path.Dir(keyFile)); os.IsNotExist(err) {
		if err := os.MkdirAll(path.Dir(keyFile), 0755); err != nil {
			return "", err
		}
	}
	var key interface{}
	err := errors.New("unsupported key type")
	switch signature {
	case rsa_key:
		key, err = rsa.GenerateKey(rand.Reader, 3072)
	case ecdsa_key:
		key, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	case ed25519_key:
		_, key, err = ed25519.GenerateKey(rand.Reader)
	}
	if err != nil {
		return "", err
	}
	keyBytes, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return "", err
	}
	if err := ioutil.WriteFile(keyFile, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyBytes}), 0600); err != nil {
		return "", err
	}
	return keyFile, nil
}

func loadKey(keyFile string) (ssh.Signer, error) {
	keyBytes, err := ioutil.ReadFile(keyFile)
	if err != nil {
		return nil, err
	}
	return ssh.ParsePrivateKey(keyBytes)
}
