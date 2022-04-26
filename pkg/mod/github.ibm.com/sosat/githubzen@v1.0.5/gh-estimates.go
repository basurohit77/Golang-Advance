package githubzen

import (
	"context"
	"fmt"
	"log"

	"github.com/google/go-github/github"
)

// GetClosedEstimatesForUsers returns a map of user to number of story points closed for the given milestone.
func GetClosedEstimatesForUsers(ctx context.Context, client *github.Client, zenAuth string, milestone string) (float64, map[string]float64) {
	log.Println("GetClosedEstimatesForUsers")

	var estimates float64
	m := map[string]float64{}

	type repository struct {
		org  string
		repo string
	}

	// repositories := []repository{{"cloud-sre", "ToolsPlatform"}, {"sosat", "general"}}
	repositories := []repository{}

	sosat := GetRepoListForOrg(ctx, client, "sosat")
	for _, r := range sosat {
		repositories = append(repositories, repository{"sosat", r})
	}

	cloudSRE := GetRepoListForOrg(ctx, client, "cloud-sre")
	for _, r := range cloudSRE {
		repositories = append(repositories, repository{"cloud-sre", r})
	}

	log.Printf("repositories=%+v\n", repositories)

	for _, r := range repositories {
		milestoneNumber, err := GetMilestoneNumber(ctx, client, r.org, r.repo, milestone)
		log.Printf("milestoneNumber=%d, err=%v\n", milestoneNumber, err)
		if err != nil {
			continue
		}

		opt := &github.IssueListByRepoOptions{
			Milestone:   fmt.Sprintf("%d", milestoneNumber), // the milestone number | none | *
			State:       "closed",                           // open | closed | all
			ListOptions: github.ListOptions{PerPage: 1000},
		}

		issues := GetIssues(ctx, client, r.org, r.repo, opt, 0)

		for _, issue := range issues {
			repoID := GetRepoIDFromZen(ctx, client, r.org, r.repo)
			log.Printf("org=%s, repoName=%s, repoID=%d\n", r.org, r.repo, repoID)

			zenIssue, err := GetZenIssue(zenAuth, repoID, issue.GetNumber())
			if err != nil {
				log.Println(err)
			}

			// Skip epics
			if zenIssue.IsEpic == true {
				continue
			}

			// TODO: It is possible that no one is assigned
			// e.g. https://github.ibm.com/cloud-sre/ToolsPlatform/issues/7770
			// Get who closed the issue or leave unclaimed?

			log.Printf("zenIssue=%+v, estimate=%f, user=%s, repo=%s, title=%s\n", zenIssue, zenIssue.Estimate.Value, issue.GetAssignee().GetLogin(), r.repo, issue.GetTitle())
			estimates += zenIssue.Estimate.Value
			m[issue.GetAssignee().GetLogin()] += zenIssue.Estimate.Value
		}

	}
	log.Printf("totalEstimates=%f, map=%+v\n", estimates, m)
	return estimates, m
}
