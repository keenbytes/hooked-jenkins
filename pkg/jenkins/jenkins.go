package jenkins

import (
	"errors"
	"strconv"
)

type Jenkins struct {
	User         string            `json:"user"`
	Token        string            `json:"token"`
	BaseURL      string            `json:"base_url"`
	Endpoints    []JenkinsEndpoint `json:"endpoints"`
	EndpointsMap map[string]*JenkinsEndpoint
}

type JenkinsEndpoint struct {
	Id        string                 `json:"id"`
	Path      string                 `json:"path"`
	Retry     JenkinsEndpointRetry   `json:"retry"`
	Success   JenkinsEndpointSuccess `json:"success"`
	Condition string                 `json:"condition"`
}

type JenkinsEndpointRetry struct {
	Delay string `json:"delay"`
	Count string `json:"count"`
}

type JenkinsEndpointSuccess struct {
	HTTPStatus string `json:"http_status"`
}

func (endpoint *JenkinsEndpoint) GetRetryCount() (int, error) {
	rc := int(1)
	if endpoint.Retry.Count != "" {
		i, err := strconv.Atoi(endpoint.Retry.Count)
		if err != nil {
			return 0, errors.New("value of Retry.Count cannot be converted to int")
		}
		rc = i
	}
	return rc, nil
}

func (endpoint *JenkinsEndpoint) GetRetryDelay() (int, error) {
	rd := int(0)
	if endpoint.Retry.Delay != "" {
		i, err := strconv.Atoi(endpoint.Retry.Count)
		if err != nil {
			return 0, errors.New("value of Retry.Delay cannot be converted to int")
		}
		rd = i
	}
	return rd, nil
}

func (endpoint *JenkinsEndpoint) CheckHTTPStatus(statusCode int) bool {
	expected, err := strconv.Atoi(endpoint.Success.HTTPStatus)
	if err != nil {
		return false
	}
	if statusCode != expected {
		return false
	}
	return true
}
