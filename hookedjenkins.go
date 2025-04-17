package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"gopkg.pl/mikogs/hooked-jenkins/pkg/githubwebhookpayload"
	"gopkg.pl/mikogs/hooked-jenkins/pkg/jenkins"
	jenkinsapi "gopkg.pl/mikogs/hooked-jenkins/pkg/jenkinsapi"
	"gopkg.pl/mikogs/hooked-jenkins/pkg/trigger"
)

type hookedJenkins struct {
	config   *config
	logLevel int
}

func (hj *hookedJenkins) startAPI() {
	router := mux.NewRouter()
	router.HandleFunc("/", hj.apiHandler).Methods("POST")

	log.Print("Starting daemon listening on " + hj.config.Port + "...")
	log.Fatal(http.ListenAndServe(fmt.Sprint(":%s", hj.config.Port), router))
}

func (hj *hookedJenkins) apiHandler(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	event := githubwebhookpayload.GetEvent(r)
	signature := githubwebhookpayload.GetSignature(r)
	if hj.config.Secret != "" {
		if !githubwebhookpayload.VerifySignature([]byte(hj.config.Secret), signature, &b) {
			http.Error(w, "Signature verification failed", 401)
			return
		}
	}

	if event != "ping" {
		err = hj.processGitHubPayload(&b, event)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		err = hj.forwardGitHubPayload(&b, r.Header)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("content-type", "application/json")
}

func (hj *hookedJenkins) processGitHubPayload(b *([]byte), event string) error {
	j := make(map[string]interface{})
	err := json.Unmarshal(*b, &j)
	if err != nil {
		return errors.New("Got non-JSON payload")
	}

	if hj.config.Triggers.Jenkins != nil {
		for _, t := range hj.config.Triggers.Jenkins {
			err = hj.processPayloadOnJenkinsTrigger(&t, j, event)
			if err != nil {
				log.Print("Error processing endpoint " + t.Endpoint + ". Breaking.")
				break
			}
		}
	}
	return nil
}

func (hj *hookedJenkins) forwardGitHubPayload(b *([]byte), h http.Header) error {
	githubHeaders := []string{"X-GitHubPayload-Event", "X-Hub-Signature", "X-GitHubPayload-Delivery", "content-type"}
	if hj.config.Forward == nil {
		return nil
	}

	for _, f := range *(hj.config.Forward) {
		if f.URL == "" {
			continue
		}

		req, err := http.NewRequest("POST", f.URL, bytes.NewReader(*b))
		if err != nil {
			return err
		}

		if f.Headers {
			for _, k := range githubHeaders {
				if h.Get(k) != "" {
					req.Header.Add(k, h.Get(k))
				}
			}
		}

		c := &http.Client{}
		_, err = c.Do(req)
		if err != nil {
			return err
		}

		log.Print("Forwarded to endpoint " + f.URL)
	}

	return nil
}

func (hj *hookedJenkins) processPayloadOnJenkinsTrigger(jenkinstrigger *trigger.JenkinsTrigger, j map[string]interface{}, event string) error {
	repo := githubwebhookpayload.GetRepository(j, event)
	ref := githubwebhookpayload.GetRef(j, event)
	branch := githubwebhookpayload.GetBranch(j, event)
	action := ""
	if jenkinstrigger.Events.PullRequest != nil && event == "pull_request" {
		action = githubwebhookpayload.GetAction(j, event)
	}
	if repo == "" {
		return nil
	}

	if event == "push" && ref == "" {
		return nil
	}

	if event == "push" && branch == "" {
		return nil
	}

	endp := hj.config.Jenkins.EndpointsMap[jenkinstrigger.Endpoint]
	if endp == nil {
		return nil
	}

	err := jenkinstrigger.CheckEvent(repo, branch, action, event)
	if err != nil {
		return nil
	}

	rd, err := endp.GetRetryDelay()
	if err != nil {
		return nil
	}
	rc, err := endp.GetRetryCount()
	if err != nil {
		return nil
	}

	return hj.processJenkinsEndpointRetries(endp, repo, branch, rd, rc)
}

func (hj *hookedJenkins) processJenkinsEndpointRetries(endpointDef *jenkins.JenkinsEndpoint, repo string, branch string, retryDelay int, retryCount int) error {
	iterations := int(0)
	if retryCount <= 0 {
		return errors.New("Unable to post to endpoint " + endpointDef.Path)
	}

	for iterations < retryCount {
		hj.printIteration(iterations, retryCount)

		crumb, err := hj.getCrumbAndSleep(hj.config.Jenkins.User, hj.config.Jenkins.Token, retryDelay)
		if err != nil {
			iterations++
			continue
		}

		endpointPath := hj.replacePathWithRepoAndBranch(endpointDef.Path, repo, branch)

		resp, err := jenkinsapi.Post(hj.config.Jenkins.BaseURL+"/"+endpointPath, hj.config.Jenkins.User, hj.config.Jenkins.Token, crumb)
		if err != nil {
			log.Print("Error from request to " + endpointPath)
			time.Sleep(time.Second * time.Duration(retryDelay))
			iterations++
			continue
		}

		log.Print("Posted to endpoint " + endpointPath)

		if !endpointDef.CheckHTTPStatus(resp.StatusCode) {
			rs := strconv.Itoa(resp.StatusCode)
			log.Print("HTTP Status " + rs + " different than expected ")
			time.Sleep(time.Second * time.Duration(retryDelay))
			iterations++
			continue
		}

		return nil
	}
	return nil
}

func (hj *hookedJenkins) printIteration(i int, rc int) {
	log.Print("Retry: (" + strconv.Itoa(i+1) + "/" + strconv.Itoa(rc) + ")")
}

func (hj *hookedJenkins) getCrumbAndSleep(u string, t string, rd int) (string, error) {
	crumb, err := jenkinsapi.GetCrumb(hj.config.Jenkins.BaseURL, u, t)
	if err != nil {
		log.Print("Error getting crumb")
		time.Sleep(time.Second * time.Duration(rd))
		return "", errors.New("Error getting crumb")
	}
	return crumb, nil
}

func (hj *hookedJenkins) replacePathWithRepoAndBranch(p string, r string, b string) string {
	s := strings.ReplaceAll(p, "{{.repository}}", r)
	s = strings.ReplaceAll(s, "{{.branch}}", b)
	return s
}
