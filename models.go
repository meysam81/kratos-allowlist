package main

import (
	"strings"

	"github.com/meysam81/x/logging"
)

type Config struct {
	Port           string   `koanf:"port"`
	AllowedDomains []string `koanf:"allowed-domains"`
}

type WebhookRequest struct {
	Email string `json:"email"`
}

type ValidationMessage struct {
	ID      int                    `json:"id"`
	Text    string                 `json:"text"`
	Type    string                 `json:"type"`
	Context map[string]interface{} `json:"context,omitempty"`
}

type FieldMessage struct {
	InstancePtr string              `json:"instance_ptr"`
	Messages    []ValidationMessage `json:"messages"`
}

type ValidationResponse struct {
	Messages []FieldMessage `json:"messages"`
}

type Identity struct {
	MetadataAdmin map[string]string `json:"metadata_admin"`
}

type SuccessfulResponse struct {
	Identity *Identity `json:"identity"`
}

type AppState struct {
	logger         *logging.Logger
	allowedDomains map[string]bool
}

func NewApp(c *Config) *AppState {
	allowedDomains := make(map[string]bool)
	for _, domain := range c.AllowedDomains {
		allowedDomains[strings.ToLower(domain)] = true
	}

	logger := logging.NewLogger()

	return &AppState{
		logger:         &logger,
		allowedDomains: allowedDomains,
	}
}
