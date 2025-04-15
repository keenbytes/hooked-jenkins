package main

type forward struct {
	URL     string `json:"url"`
	Headers bool   `json:"headers,omitempty"`
}
