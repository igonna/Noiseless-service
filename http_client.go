package main

import (
	"net/http"
	"time"
)

const (
	defaultHTTPTimeout = 20 * time.Second
	openAIHTTPTimeout  = 75 * time.Second
)

var (
	httpClient       = &http.Client{Timeout: defaultHTTPTimeout}
	openaiHTTPClient = &http.Client{Timeout: openAIHTTPTimeout}
)
