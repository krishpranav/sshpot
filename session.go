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
