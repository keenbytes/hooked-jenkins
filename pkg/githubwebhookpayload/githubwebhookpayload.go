package githubwebhookpayload

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"net/http"
	"strings"
)

func GetEvent(r *http.Request) string {
	return r.Header.Get("X-GitHub-Event")
}

func GetSignature(r *http.Request) string {
	return r.Header.Get("X-Hub-Signature")
}

func signBody(secret []byte, body []byte) []byte {
	computed := hmac.New(sha1.New, secret)
	computed.Write(body)
	return []byte(computed.Sum(nil))
}

func VerifySignature(secret []byte, signature string, body *([]byte)) bool {
	actual := make([]byte, 20)
	hex.Decode(actual, []byte(signature[5:]))
	return hmac.Equal(signBody(secret, *body), actual)
}

func GetRef(j map[string]interface{}, event string) string {
	if j["ref"] != nil {
		return j["ref"].(string)
	}

	return ""
}
func GetRefType(j map[string]interface{}, event string) string {
	if j["ref_type"] != nil {
		return j["ref_type"].(string)
	}

	return ""
}
func GetBranch(j map[string]interface{}, event string) string {
	if event == "push" {
		ref := strings.Split(j["ref"].(string), "/")
		if ref[1] == "tag" {
			return ""
		}
		branch := ref[2]
		return branch
	}
	if event == "create" || event == "delete" {
		ref := j["ref"].(string)
		refType := j["ref_type"].(string)
		if refType != "branch" {
			return ""
		}
		return ref
	}
	return ""
}
func GetAction(j map[string]interface{}, event string) string {
	if event == "pull_request" && j["action"] != nil {
		return j["action"].(string)
	}
	return ""
}
func GetRepository(j map[string]interface{}, event string) string {
	switch event {
	case "push", "create", "delete":
		v, ok := j["repository"].(map[string]interface{})
		if !ok {
			return ""
		}
		v2, ok := v["name"]
		if ok {
			return v2.(string)
		}
		return ""
	case "pull_request":
		v, ok := j["pull_request"].(map[string]interface{})
		if !ok {
			return ""
		}
		v2, ok := v["head"]
		if !ok {
			return ""
		}
		v3, ok := v2.(map[string]interface{})
		if !ok {
			return ""
		}
		v4, ok := v3["repo"]
		if !ok {
			return ""
		}
		v5, ok := v4.(map[string]interface{})
		if !ok {
			return ""
		}
		v6, ok := v5["name"]
		if ok {
			return v6.(string)
		}
		return ""
	default:
		return ""
	}
}
