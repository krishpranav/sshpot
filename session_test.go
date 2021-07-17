package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"reflect"
	"testing"
	"time"

	"golang.org/x/crypto/ssh"
)

func testSession(t *testing.T, dataDir string, cfg *config, clientAddress string) string {
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

	// Raw exec
	channel, channelRequests, err := conn.OpenChannel("session", nil)
	if err != nil {
		t.Fatalf("Failed to open channel: %v", err)
	}
	channelRequestTypes := []string{}
	channelRequestsDone := make(chan interface{})
	go func() {
		for request := range channelRequests {
			channelRequestTypes = append(channelRequestTypes, request.Type)
		}
		channelRequestsDone <- nil
	}()
	accepted, err := channel.SendRequest("x11-req", true, ssh.Marshal(struct {
		SingleConnection         bool
		AuthProtocol, AuthCookie string
		ScreenNumber             uint32
	}{false, "MIT-MAGIC-COOKIE-1", "e16b9dbcaa8678ae85572677d847a3a5", 0}))
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	if !accepted {
		t.Errorf("accepted=false, want true")
	}
	_, err = channel.SendRequest("env", false, ssh.Marshal(struct {
		Name, Value string
	}{"LANG", "en_IE.UTF-8"}))
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	accepted, err = channel.SendRequest("exec", true, ssh.Marshal(struct {
		Command string
	}{"sh"}))
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	if !accepted {
		t.Errorf("accepted=false, want true")
	}
	if _, err := channel.Write([]byte("false\ntrue\n")); err != nil {
		t.Fatalf("Faield to write to channel: %v", err)
	}
	if err := channel.CloseWrite(); err != nil {
		t.Fatalf("Failed to close channel: %v", err)
	}
	channelResponse, err := ioutil.ReadAll(channel)
	if err != nil {
		t.Fatalf("Failed to read channel: %v", err)
	}
	expectedChannelResponse := ""
	if string(channelResponse) != expectedChannelResponse {
		t.Errorf("channelResponse=%v, want %v", string(channelResponse), expectedChannelResponse)
	}
	<-channelRequestsDone
	expectedChannelRequestTypes := []string{"exit-status"}
	if !reflect.DeepEqual(channelRequestTypes, expectedChannelRequestTypes) {
		t.Errorf("channelRequestTypes=%v, want %v", channelRequestTypes, expectedChannelRequestTypes)
	}
	channelRequestTypes = []string{}
	time.Sleep(10 * time.Millisecond)

	// Raw shell
	channel, channelRequests, err = conn.OpenChannel("session", nil)
	if err != nil {
		t.Fatalf("Failed to open channel: %v", err)
	}
	go func() {
		for request := range channelRequests {
			channelRequestTypes = append(channelRequestTypes, request.Type)
		}
		channelRequestsDone <- nil
	}()
	accepted, err = channel.SendRequest("x11-req", true, ssh.Marshal(struct {
		SingleConnection         bool
		AuthProtocol, AuthCookie string
		ScreenNumber             uint32
	}{false, "MIT-MAGIC-COOKIE-1", "e16b9dbcaa8678ae85572677d847a3a5", 0}))
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	if !accepted {
		t.Errorf("accepted=false, want true")
	}
	_, err = channel.SendRequest("env", false, ssh.Marshal(struct {
		Name, Value string
	}{"LANG", "en_IE.UTF-8"}))
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	accepted, err = channel.SendRequest("shell", true, nil)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	if !accepted {
		t.Errorf("accepted=false, want true")
	}
	if _, err := channel.Write([]byte("false\ntrue\n")); err != nil {
		t.Fatalf("Faield to write to channel: %v", err)
	}
	if err := channel.CloseWrite(); err != nil {
		t.Fatalf("Failed to close channel: %v", err)
	}
	channelResponse, err = ioutil.ReadAll(channel)
	if err != nil {
		t.Fatalf("Failed to read channel: %v", err)
	}
	expectedChannelResponse = ""
	if string(channelResponse) != expectedChannelResponse {
		t.Errorf("channelResponse=%v, want %v", string(channelResponse), expectedChannelResponse)
	}
	<-channelRequestsDone
	expectedChannelRequestTypes = []string{"exit-status"}
	if !reflect.DeepEqual(channelRequestTypes, expectedChannelRequestTypes) {
		t.Errorf("channelRequestTypes=%v, want %v", channelRequestTypes, expectedChannelRequestTypes)
	}
	channelRequestTypes = []string{}
	time.Sleep(10 * time.Millisecond)

	terminalModes, err := base64.RawStdEncoding.DecodeString("gQAAJYCAAAAlgAEAAAADAgAAABwDAAAAfwQAAAAVBQAAAAQGAAAA/wcAAAD/CAAAABEJAAAAEwoAAAAaCwAAABkMAAAAEg0AAAAXDgAAABYRAAAAFBIAAAAPHgAAAAAfAAAAACAAAAAAIQAAAAAiAAAAACMAAAAAJAAAAAEmAAAAACcAAAABKAAAAAApAAAAASoAAAABMgAAAAEzAAAAATUAAAABNgAAAAE3AAAAADgAAAAAOQAAAAA6AAAAADsAAAABPAAAAAE9AAAAAT4AAAAARgAAAAFIAAAAAUkAAAAASgAAAABLAAAAAFoAAAABWwAAAAFcAAAAAF0AAAAAAA")
	if err != nil {
		t.Errorf("Faield to encode terminal modes: %v", err)
	}

	// PTY exec
	channel, channelRequests, err = conn.OpenChannel("session", nil)
	if err != nil {
		t.Fatalf("Failed to open channel: %v", err)
	}
	go func() {
		for request := range channelRequests {
			channelRequestTypes = append(channelRequestTypes, request.Type)
		}
		channelRequestsDone <- nil
	}()
	accepted, err = channel.SendRequest("x11-req", true, ssh.Marshal(struct {
		SingleConnection         bool
		AuthProtocol, AuthCookie string
		ScreenNumber             uint32
	}{false, "MIT-MAGIC-COOKIE-1", "e16b9dbcaa8678ae85572677d847a3a5", 0}))
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	if !accepted {
		t.Errorf("accepted=false, want true")
	}
	accepted, err = channel.SendRequest("pty-req", true, ssh.Marshal(struct {
		Term                                   string
		Width, Height, PixelWidth, PixelHeight uint32
		Modes                                  string
	}{"xterm-256color", 80, 24, 123, 456, string(terminalModes)}))
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	if !accepted {
		t.Errorf("accepted=false, want true")
	}
	_, err = channel.SendRequest("env", false, ssh.Marshal(struct {
		Name, Value string
	}{"LANG", "en_IE.UTF-8"}))
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	accepted, err = channel.SendRequest("exec", true, ssh.Marshal(struct {
		Command string
	}{"sh"}))
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	if !accepted {
		t.Errorf("accepted=false, want true")
	}
	if _, err := channel.Write([]byte("false\rtrue\r\x04")); err != nil {
		t.Fatalf("Faield to write to channel: %v", err)
	}
	channelResponse, err = ioutil.ReadAll(channel)
	if err != nil {
		t.Fatalf("Failed to read channel: %v", err)
	}
	expectedChannelResponse = "$ false\r\n$ true\r\n$ \r\n"
	if string(channelResponse) != expectedChannelResponse {
		t.Errorf("channelResponse=%v, want %v", string(channelResponse), expectedChannelResponse)
	}
	<-channelRequestsDone
	expectedChannelRequestTypes = []string{"exit-status", "eow@openssh.com"}
	if !reflect.DeepEqual(channelRequestTypes, expectedChannelRequestTypes) {
		t.Errorf("channelRequestTypes=%v, want %v", channelRequestTypes, expectedChannelRequestTypes)
	}
	channelRequestTypes = []string{}
	time.Sleep(10 * time.Millisecond)

	// PTY shell
	channel, channelRequests, err = conn.OpenChannel("session", nil)
	if err != nil {
		t.Fatalf("Failed to open channel: %v", err)
	}
	go func() {
		for request := range channelRequests {
			channelRequestTypes = append(channelRequestTypes, request.Type)
		}
		channelRequestsDone <- nil
	}()
	accepted, err = channel.SendRequest("x11-req", true, ssh.Marshal(struct {
		SingleConnection         bool
		AuthProtocol, AuthCookie string
		ScreenNumber             uint32
	}{false, "MIT-MAGIC-COOKIE-1", "e16b9dbcaa8678ae85572677d847a3a5", 0}))
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	if !accepted {
		t.Errorf("accepted=false, want true")
	}
	accepted, err = channel.SendRequest("pty-req", true, ssh.Marshal(struct {
		Term                                   string
		Width, Height, PixelWidth, PixelHeight uint32
		Modes                                  string
	}{"xterm-256color", 80, 24, 123, 456, string(terminalModes)}))
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	if !accepted {
		t.Errorf("accepted=false, want true")
	}
	_, err = channel.SendRequest("env", false, ssh.Marshal(struct {
		Name, Value string
	}{"LANG", "en_IE.UTF-8"}))
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	accepted, err = channel.SendRequest("shell", true, nil)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	if !accepted {
		t.Errorf("accepted=false, want true")
	}
	if _, err := channel.Write([]byte("false\rtrue\r\x04")); err != nil {
		t.Fatalf("Faield to write to channel: %v", err)
	}
	channelResponse, err = ioutil.ReadAll(channel)
	if err != nil {
		t.Fatalf("Failed to read channel: %v", err)
	}
	expectedChannelResponse = "$ false\r\n$ true\r\n$ \r\n"
	if string(channelResponse) != expectedChannelResponse {
		t.Errorf("channelResponse=%v, want %v", string(channelResponse), expectedChannelResponse)
	}
	<-channelRequestsDone
	expectedChannelRequestTypes = []string{"exit-status", "eow@openssh.com"}
	if !reflect.DeepEqual(channelRequestTypes, expectedChannelRequestTypes) {
		t.Errorf("channelRequestTypes=%v, want %v", channelRequestTypes, expectedChannelRequestTypes)
	}
	channelRequestTypes = []string{}
	time.Sleep(10 * time.Millisecond)

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
