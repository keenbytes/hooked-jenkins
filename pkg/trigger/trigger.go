package trigger

import (
	"errors"
	"log"
)

type Trigger struct {
	Jenkins []JenkinsTrigger `json:"jenkins"`
}

type JenkinsTrigger struct {
	Endpoint string `json:"endpoint"`
	Events   Events `json:"events"`
}

type Events struct {
	Push        *EndpointConditions `json:"push,omitempty"`
	PullRequest *EndpointConditions `json:"pull_request,omitempty"`
	Create      *EndpointConditions `json:"create,omitempty"`
	Delete      *EndpointConditions `json:"delete,omitempty"`
}

type EndpointConditions struct {
	Repositories        *([]EndpointConditionRepository) `json:"repositories,omitempty"`
	Branches            *([]EndpointConditionBranch)     `json:"branches,omitempty"`
	ExcludeRepositories *([]EndpointConditionRepository) `json:"exclude_repositories,omitempty"`
	ExcludeBranches     *([]EndpointConditionBranch)     `json:"exclude_branches,omitempty"`
	Actions             *([]string)                      `json:"actions,omitempty"`
}

type EndpointConditionRepository struct {
	Name     string      `json:"name"`
	Branches *([]string) `json:"branches,omitempty"`
}

type EndpointConditionBranch struct {
	Name         string      `json:"name"`
	Repositories *([]string) `json:"repositories,omitempty"`
}

func (jenkinstrigger *JenkinsTrigger) CheckEvent(repo string, branch string, action string, event string) error {
	if jenkinstrigger.Events.PullRequest != nil && event == "pull_request" {
		if action == "" {
			return errors.New("action is empty")
		}
		inActions := jenkinstrigger.Events.PullRequest.CheckActions(action)
		if !inActions {
			return errors.New("Event " + event + "not supported")
		}
	}

	var c *EndpointConditions
	if event == "push" && jenkinstrigger.Events.Push != nil {
		c = jenkinstrigger.Events.Push
	} else if event == "pull_request" && jenkinstrigger.Events.PullRequest != nil {
		c = jenkinstrigger.Events.PullRequest
	} else if event == "create" && jenkinstrigger.Events.Create != nil {
		c = jenkinstrigger.Events.Create
	} else if event == "delete" && jenkinstrigger.Events.Delete != nil {
		c = jenkinstrigger.Events.Delete
	} else {
		return errors.New("Event " + event + "not supported")
	}

	inRepos := false
	if c.Repositories != nil {
		inRepos = c.CheckRepositories(repo, branch, false)
	}
	inBranches := false
	if c.Branches != nil && event == "push" {
		inBranches = c.CheckBranches(branch, repo, false)
	}
	inExcludeRepos := false
	if c.ExcludeRepositories != nil {
		inExcludeRepos = c.CheckRepositories(repo, branch, true)
	}
	inExcludeBranches := false
	if c.ExcludeBranches != nil && event == "push" {
		inExcludeBranches = c.CheckBranches(branch, repo, true)
	}
	if (inRepos || inBranches) && !inExcludeRepos && !inExcludeBranches {
		return nil
	}

	return errors.New("Event " + event + "not supported")
}

func (cond *EndpointConditions) CheckActions(action string) bool {
	for _, a := range *cond.Actions {
		if a == action || a == "*" {
			return true
		}
	}
	return false
}

func (cond *EndpointConditions) CheckRepositories(repo string, branch string, exclude bool) bool {
	repos := cond.Repositories
	if exclude {
		repos = cond.ExcludeRepositories
	}
	for _, r := range *repos {
		if r.Name == repo || r.Name == "*" {
			if r.Branches == nil || len(*(r.Branches)) == 0 {
				log.Print("Found " + r.Name + " repo")
				return true
			}
			for _, b := range *(r.Branches) {
				if b != branch {
					continue
				}
				log.Print("Found " + b + " branch in " + r.Name + " repo")
				return true
			}
		}
	}
	return false
}
func (cond *EndpointConditions) CheckBranches(branch string, repo string, exclude bool) bool {
	branches := cond.Branches
	if exclude {
		branches = cond.ExcludeBranches
	}
	for _, b := range *branches {
		if b.Name != branch && b.Name != "*" {
			continue
		}

		if b.Repositories == nil || len(*(b.Repositories)) == 0 {
			log.Print("Found " + b.Name + " branch")
			return true
		}

		for _, r := range *(b.Repositories) {
			if r == repo {
				log.Print("Found " + r + " repository in " + b.Name + " branch")
				return true
			}
		}
	}
	return false
}
