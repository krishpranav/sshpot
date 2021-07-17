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
