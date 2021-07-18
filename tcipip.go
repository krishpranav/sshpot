package main

import (
	"bufio"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"strconv"

	"golang.org/x/crypto/ssh"
)

type tcpipServer interface {
	server(channel ssh.Channel, input chan<- string) error
}

var servers = map[uint32]tcpipServer{
	80: httpServer{},
}

type tcpipChannelData struct {
	Address           string
	Port              uint32
	OriginatorAddress string
	OrignatorPort     uint32
}

func handleDirectTCPIPChannel(newChannel ssh.NewChannel, context channelContext) error {
	channelData := &tcpipChannelData{}
	if err := ssh.Unmarshal(newChannel.ExtraData(), channelData); err != nil {
		return err
	}
	server := servers[channelData.Port]
	if server == nil {
		warningLogger.Printf("Unsupported port %v", channelData.Port)
		return newChannel.Reject(ssh.ConnectionFailed, "Connection refused")
	}
	channel, requests, err := newChannel.Accept()
	if err != nil {
		return err
	}
	context.logEvent(directTCPIPLog{
		channelLog: channelLog{
			ChannelID: context.channelID,
		},
		From: net.JoinHostPort(channelData.OriginatorAddress, strconv.Itoa(int(channelData.OriginatorPort))),
		To:   net.JoinHostPort(channelData.Address, strconv.Itoa(int(channelData.Port))),
	})
	defer context.logEvent(directTCPIPCloseLog{
		channelLog: channelLog{
			ChannelID: context.channelID,
		},
	})

	inputChan := make(chan string)
	errorChan := make(chan error)
	go func() {
		defer close(inputChan)
		defer close(errorChan)
		errorChan <- server.serve(channel, inputChan)
	}()

	for inputChan != nil || errorChan != nil || requests != nil {
		select {
		case input, ok := <-inputChan:
			if !ok {
				inputChan = nil
				continue
			}
			context.logEvent(directTCPIPInputLog{
				channelLog: channelLog{
					ChannelID: context.channelID,
				},
				Input: input,
			})
		case err, ok := <-errorChan:
			if !ok {
				errorChan = nil
				continue
			}
			if err != nil {
				return err
			}
		case request, ok := <-requests:
			if !ok {
				requests = nil
				continue
			}
			context.logEvent(debugChannelRequestLog{
				channelLog:  channelLog{ChannelID: context.channelID},
				RequestType: request.Type,
				WantReply:   request.WantReply,
				Payload:     string(request.Payload),
			})
			warningLogger.Printf("Unsupported direct-tcpip request type %v", request.Type)
			if request.WantReply {
				if err := request.Reply(false, nil); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
