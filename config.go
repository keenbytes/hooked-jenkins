package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
	"github.com/keenbytes/hooked-jenkins/pkg/jenkins"
	"github.com/keenbytes/hooked-jenkins/pkg/trigger"
)

type config struct {
	Version  string          `json:"version"`
	Port     string          `json:"port"`
	Jenkins  jenkins.Jenkins `json:"jenkins"`
	Triggers trigger.Trigger `json:"triggers"`
	Forward  *([]forward)    `json:"forward"`
	Secret   string          `json:"secret,omitempty"`
	logLevel int
}

func (cfg *config) readFile(p string) error {
	b, err := os.ReadFile(p)
	if err != nil {
		return fmt.Errorf("error reading file %s: %w", p, err)
	}

	err = yaml.Unmarshal(b, &cfg)
	if err != nil {
		return fmt.Errorf("error unmarshalling: %w", err)
	}

	cfg.Jenkins.EndpointsMap = make(map[string]*jenkins.JenkinsEndpoint, len(cfg.Jenkins.Endpoints))
	for i, e := range cfg.Jenkins.Endpoints {
		cfg.Jenkins.EndpointsMap[e.Id] = &(cfg.Jenkins.Endpoints[i])
	}

	return nil
}
