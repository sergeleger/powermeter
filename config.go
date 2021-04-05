package main

import (
	"encoding/json"
	"io"
)

// Config provides external parameters.
type Config struct {
	Token  string `json:"token"`
	Bucket string `json:"bucket"`
	Org    string `json:"org"`
	Host   string `json:"host"`
}

func (c *Config) ReadFrom(r io.Reader) error {
	return json.NewDecoder(r).Decode(c)
}

func (c *Config) WriteTo(w io.Writer) error {
	return json.NewEncoder(w).Encode(c)
}
