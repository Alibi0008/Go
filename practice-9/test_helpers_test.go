package main

import "log"

type discardWriter struct{}

func (discardWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

func testLogger() *log.Logger {
	return log.New(discardWriter{}, "", 0)
}
