package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path"
	"reflect"
	"testing"

	"golang.org/x/crypto/ssh"
)

type mockPublicKey struct {
	signature keySignature
}

func (publicKey mockPublicKey) Type() string {
	return publicKey.signature.String()
}

func (publicKey mockPublicKey) Marshal() []byte {
	return []byte(publicKey.signature.String())
}

func (publicKey mockPublicKey) Verify(data []byte, sig *ssh.Signature) error {
	return nil
}

type mockFile struct {
	closed bool
}

func (file *mockFile) Write(p []byte) (n int, err error) {
	return 0, errors.New("")
}

func (file *mocFile) Close() error {
	if file.closed {
		return errors.New("")
	}
	file.closed = true
	return nil
}

func verifyConfig(t *testing.T, cfg *config, expected *config) {
	if !reflect.DeepEqual(cfg.Server, expected.Server) {
		t.Errorf("Server=%v, want %v", cfg.Server, expected.Server)
	}
	if !reflect.DeepEqual(cfg.Logging, expected.Logging) {
		t.Errorf("Logging=%v, want %v", cfg.Logging, expected.Logging)
	}
	if !reflect.DeepEqual(cfg.Auth, expected.Auth) {
		t.Errorf("Auth=%v, want %v", cfg.Auth, expected.Auth)
	}
	if !reflect.DeepEqual(cfg.SSHProto, expected.SSHProto) {
		t.Errorf("SSHProto=%v, want %v", cfg.SSHProto, expected.SSHProto)
	}

	if cfg.sshConfig.RekeyThreshold != expected.SSHProto.RekeyThreshold {
		t.Errorf("sshConfig.RekeyThreshold=%v, want %v", cfg.sshConfig.RekeyThreshold, expected.SSHProto.RekeyThreshold)
	}
	if !reflect.DeepEqual(cfg.sshConfig.KeyExchanges, expected.SSHProto.KeyExchanges) {
		t.Errorf("sshConfig.KeyExchanges=%v, want %v", cfg.sshConfig.KeyExchanges, expected.SSHProto.KeyExchanges)
	}
	if !reflect.DeepEqual(cfg.sshConfig.Ciphers, expected.SSHProto.Ciphers) {
		t.Errorf("sshConfig.Ciphers=%v, want %v", cfg.sshConfig.Ciphers, expected.SSHProto.Ciphers)
	}
	if !reflect.DeepEqual(cfg.sshConfig.MACs, expected.SSHProto.MACs) {
		t.Errorf("sshConfig.MACs=%v, want %v", cfg.sshConfig.MACs, expected.SSHProto.MACs)
	}
	if cfg.sshConfig.NoClientAuth != expected.Auth.NoAuth {
		t.Errorf("sshConfig.NoClientAuth=%v, want %v", cfg.sshConfig.NoClientAuth, expected.Auth.NoAuth)
	}
	if cfg.sshConfig.MaxAuthTries != expected.Auth.MaxTries {
		t.Errorf("sshConfig.MaxAuthTries=%v, want %v", cfg.sshConfig.MaxAuthTries, expected.Auth.MaxTries)
	}
	if (cfg.sshConfig.PasswordCallback != nil) != expected.Auth.PasswordAuth.Enabled {
		t.Errorf("sshConfig.PasswordCallback=%v, want %v", cfg.sshConfig.PasswordCallback != nil, expected.Auth.PasswordAuth.Enabled)
	}
	if (cfg.sshConfig.PublicKeyCallback != nil) != expected.Auth.PublicKeyAuth.Enabled {
		t.Errorf("sshConfig.PasswordCallback=%v, want %v", cfg.sshConfig.PublicKeyCallback != nil, expected.Auth.PublicKeyAuth.Enabled)
	}
	if (cfg.sshConfig.KeyboardInteractiveCallback != nil) != expected.Auth.KeyboardInteractiveAuth.Enabled {
		t.Errorf("sshConfig.KeyboardInteractiveCallback=%v, want %v", cfg.sshConfig.KeyboardInteractiveCallback != nil, expected.Auth.KeyboardInteractiveAuth.Enabled)
	}
	if cfg.sshConfig.AuthLogCallback == nil {
		t.Errorf("sshConfig.AuthLogCallback=nil, want a callback")
	}
	if cfg.sshConfig.ServerVersion != expected.SSHProto.Version {
		t.Errorf("sshConfig.ServerVersion=%v, want %v", cfg.sshConfig.ServerVersion, expected.SSHProto.Version)
	}
	if (cfg.sshConfig.BannerCallback != nil) != (expected.SSHProto.Banner != "") {
		t.Errorf("sshConfig.BannerCallback=%v, want %v", cfg.sshConfig.BannerCallback != nil, expected.SSHProto.Banner != "")
	}
	if cfg.sshConfig.GSSAPIWithMICConfig != nil {
		t.Errorf("sshConfig.GSSAPIWithMICConfig=%v, want nil", cfg.sshConfig.GSSAPIWithMICConfig)
	}
	if len(cfg.parsedHostKeys) != len(expected.Server.HostKeys) {
		t.Errorf("len(parsedHostKeys)=%v, want %v", len(cfg.parsedHostKeys), len(expected.Server.HostKeys))
	}

	if expected.Logging.File == "" {
		if cfg.logFileHandle != nil {
			t.Errorf("logFileHandle=%v, want nil", cfg.logFileHandle)
		}
	} else {
		if cfg.logFileHandle == nil {
			t.Errorf("logFileHandle=nil, want a file")
		}
	}
}

func verifyDefaultKeys(t *testing.T, dataDir string) {
	files, err := ioutil.ReadDir(dataDir)
	if err != nil {
		t.Fatalf("Faield to list directory: %v", err)
	}
	expectedKeys := map[string]string{
		"host_rsa_key":     "ssh-rsa",
		"host_ecdsa_key":   "ecdsa-sha2-nistp256",
		"host_ed25519_key": "ssh-ed25519",
	}
	keys := map[string]string{}
	for _, file := range files {
		keyBytes, err := ioutil.ReadFile(path.Join(dataDir, file.Name()))
		if err != nil {
			t.Fatalf("Failed to read key: %v", err)
		}
		signer, err := ssh.ParsePrivateKey(keyBytes)
		if err != nil {
			t.Fatalf("Failed to parse private key: %v", err)
		}
		keys[file.Name()] = signer.PublicKey().Type()
	}
	if !reflect.DeepEqual(keys, expectedKeys) {
		t.Errorf("keys=%v, want %v", keys, expectedKeys)
	}
}

func TestDefaultConfig(t *testing.T) {
	dataDir := t.TempDir()
	cfg, err := getConfig("", dataDir)
	if err != nil {
		t.Fatalf("Failed to get config: %v", err)
	}
	expectedConfig := &config{}
	expectedConfig.Server.ListenAddress = "127.0.0.1:2022"
	expectedConfig.Server.HostKeys = []string{
		path.Join(dataDir, "host_rsa_key"),
		path.Join(dataDir, "host_ecdsa_key"),
		path.Join(dataDir, "host_ed25519_key"),
	}
	expectedConfig.Logging.Timestamps = true
	expectedConfig.Auth.PasswordAuth.Enabled = true
	expectedConfig.Auth.PasswordAuth.Accepted = true
	expectedConfig.Auth.PublicKeyAuth.Enabled = true
	expectedConfig.SSHProto.Version = "SSH-2.0-sshesame"
	expectedConfig.SSHProto.Banner = "This is an SSH honeypot. Everything is logged and monitored."
	verifyConfig(t, cfg, expectedConfig)
	verifyDefaultKeys(t, dataDir)
}
