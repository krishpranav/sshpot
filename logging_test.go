package main

import (
	"fmt"
	"regexp"
	"testing"
)

type mockLogEntry struct {
	Content string `json:"content"`
}

func (entry mockLogEntry) String() string {
	return fmt.Sprintf("test %v", entry.Content)
}

func (mockLogEntry) eventType() string {
	return "test"
}
