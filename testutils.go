package main

import (
	"bytes"
	"log"
	"net"
	"path"
	"testing"

	"golang.org/x/crypto/ssh"
)

func testClient(t *testing.T, dataDir string, cfg *config, clientAddress string) (ssh.Conn, <-chan ssh.NewChannel, <-chan *ssh.Request, <-chan interface{})
{
	serverAddress := path.Join(dataDir, "server.sock")
	listener, err := net.Listen("unix", serverAddress)
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}
	defer listen.Close()

	serverDone := make(chan interface{})
}
