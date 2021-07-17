package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"reflect"
	"testing"

	"golang.org/x/crypto/ssh"
)

func testTCP(t *testing.T, dataDir string, cfg *config, clientAddress string) string {
	logBuffer := setupLogBuffer(t, cfg)

	conn, newChannels, requests, done := testClient(t, dataDir, cfg, clientAddress)
	defer conn.Close()

	channelTypes := []string{}
	channelsDone := make(chan interface{})
	go func() {
		for newChannel := range newChannels {
			channelTypes = append(channelTypes, newChannel.ChannelType())
		}
		channelsDone <- nil
	}()

	requestTypes := []string{}
	requestsDone := make(chan interface{})
	go func() {
		for request := range requests {
			requestTypes = append(requestTypes, request.Type)
		}
		requestsDone <- nil
	}()

	channel, channelRequests, err := conn.OpenChannel("direct-tcpip", ssh.Marshal(struct {
		Address           string
		Port              uint32
		OriginatorAddress string
		OriginatorPort    uint32
	}{"example.org", 80, "localhost", 8080}))
	if err != nil {
		t.Fatalf("Failed to open channel: %v", err)
	}
	if _, err := channel.Write([]byte("GET / HTTP/1.1\r\n\r\n")); err != nil {
		t.Fatalf("Faield to write to channel: %v", err)
	}
	if err := channel.CloseWrite(); err != nil {
		t.Fatalf("Failed to close channel: %v", err)
	}
	channelRequestTypes := []string{}
	channelRequestsDone := make(chan interface{})
	go func() {
		for request := range channelRequests {
			channelRequestTypes = append(channelRequestTypes, request.Type)
		}
		channelRequestsDone <- nil
	}()

	channelResponse, err := ioutil.ReadAll(channel)
	if err != nil {
		t.Fatalf("Failed to read channel: %v", err)
	}
	expectedChannelResponse := "HTTP/1.1 404 Not Found\r\nContent-Length: 0\r\n\r\n"
	if string(channelResponse) != expectedChannelResponse {
		t.Errorf("channelResponse=%v, want %v", string(channelResponse), expectedChannelResponse)
	}

	<-channelRequestsDone
	if len(channelRequestTypes) != 0 {
		t.Errorf("channelRequestTypes=%v, want []", channelRequestTypes)
	}

	conn.Close()

	<-channelsDone
	<-requestsDone
	<-done

	expectedRequestTypes := []string{"hostkeys-00@openssh.com"}
	if !reflect.DeepEqual(requestTypes, expectedRequestTypes) {
		t.Errorf("requestTypes=%v, want %v", requestTypes, expectedRequestTypes)
	}

	if len(channelTypes) != 0 {
		t.Errorf("channelTypes=%v, want []", channelTypes)
	}

	return logBuffer.String()
}
