package main

import (
	"bufio"
	"errors"
	"io"
	"strings"

	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

type ptyRequest struct {
	Term                                   string
	Width, Height, PixelWidth, PixelHeight uint32
	Modes                                  string
}

func (request ptyRequest) reply() []byte {
	return nil
}
func (request ptyRequest) logEntry(channelID int) logEntry {
	return ptyLog{
		channelLog: channelLog{
			ChannelID: channelID,
		},
		Terminal: request.Term,
		Width:    request.Width,
		Height:   request.Height,
	}
}

type shellRequest struct{}

func (request shellRequest) reply() []byte {
	return nil
}
func (request shellRequest) logEntry(channelID int) logEntry {
	return shellLog{
		channelLog: channelLog{
			ChannelID: channelID,
		},
	}
}

type x11RequestPayload struct {
	SingleConnection         bool
	AuthProtocol, AuthCookie string
	ScreenNumber             uint32
}

func (request x11RequestPayload) reply() []byte {
	return nil
}
func (request x11RequestPayload) logEntry(channelID int) logEntry {
	return x11Log{
		channelLog: channelLog{
			ChannelID: channelID,
		},
		Screen: request.ScreenNumber,
	}
}

type envRequestPayload struct {
	Name, Value string
}

func (request envRequestPayload) reply() []byte {
	return nil
}
func (request envRequestPayload) logEntry(channelID int) logEntry {
	return envLog{
		channelLog: channelLog{
			ChannelID: channelID,
		},
		Name:  request.Name,
		Value: request.Value,
	}
}

type execRequestPayload struct {
	Command string
}

func (request execRequestPayload) reply() []byte {
	return nil
}
func (request execRequestPayload) logEntry(channelID int) logEntry {
	return execLog{
		channelLog: channelLog{
			ChannelID: channelID,
		},
		Command: request.Command,
	}
}

type subsystemRequestPayload struct {
	Subsystem string
}

func (request subsystemRequestPayload) reply() []byte {
	return nil
}
func (request subsystemRequestPayload) logEntry(channelID int) logEntry {
	return subsystemLog{
		channelLog: channelLog{
			ChannelID: channelID,
		},
		Subsystem: request.Subsystem,
	}
}

type windowChangeRequestPayload struct {
	Width, Height, PixelWidth, PixelHeight uint32
}

func (request windowChangeRequestPayload) reply() []byte {
	return nil
}
func (request windowChangeRequestPayload) logEntry(channelID int) logEntry {
	return windowChangeLog{
		channelLog: channelLog{
			ChannelID: channelID,
		},
		Width:  request.Width,
		Height: request.Height,
	}
}
