package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path"

	"github.com/adrg/xdg"
)

var (
	infoLogger    *log.Logger
	warningLogger *log.Logger
	errorLogger   *log.Logger
)

func init() {
	infoLogger = log.New(os.Stderr, "INFO ", log.LstdFlags)
	warningLogger = log.New(os.Stderr, "WARNING ", log.LstdFlags)
	errorLogger = log.New(os.Stderr, "ERROR ", log.LstdFlags)
}
