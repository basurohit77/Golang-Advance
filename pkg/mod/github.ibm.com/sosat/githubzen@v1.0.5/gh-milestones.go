package githubzen

import (
	"context"
	"fmt"
	"log"

	"github.com/google/go-github/github"
)

// GetMilestones gets all the milestone objects from a repo
func GetMilestones(ctx context.Context, client *github.Client, owner string, repo string) []*github.Milestone {
	for {
		// issue, found := issuesCache.read(issueNum)
		// if found {
		// 	log.Printf("getIssue: cacheHit=true, issue=%+v, resp=%+v, err=%v\n", issue.GetTitle(), resp, err)
		// 	return issue
		// }

		opt := &github.MilestoneListOptions{
			ListOptions: github.ListOptions{PerPage: 1000},
		}

		ghLimiter.Wait()
		milestones, resp, err := client.Issues.ListMilestones(ctx, owner, repo, opt)
		// log.Printf("GetMilestones: milestons=%v, resp=%v, err=%v\n", err)

		ghLimiter.Set(resp.Rate.Limit, resp.Rate.Remaining, GithubTimestampToTime(resp.Rate.Reset))

		// if err, ok := err.(*github.RateLimitError); ok {
		if isRatelimitError(resp, err) {
			log.Printf("GetMilestones: RateLimitError: %#v\n", err)
			continue
		}

		// issuesCache.write(issueNum, issue)

		// log.Printf("getIssue: cacheHit=false, issue=%+v, resp=%+v, err=%v\n", issue.GetTitle(), resp, err)
		// log.Printf("GetMilestones: %+v\n", milestones)

		return milestones
	}
}

// GetMilestone gets the milestone object given a milestone title and a repo
func GetMilestone(ctx context.Context, client *github.Client, org string, repo string, milestoneTitle string) *github.Milestone {
	ms := GetMilestones(ctx, client, org, repo)
	// log.Printf("milestones=%+v\n", ms)
	for _, m := range ms {
		if m.GetTitle() == milestoneTitle {
			return m
		}
	}
	log.Printf("milestoneNotFound=%s\n", milestoneTitle)
	return nil
}

// GetMilestoneNumber gets the milestone nubmer for a specific milestone
func GetMilestoneNumber(ctx context.Context, client *github.Client, org string, repo string, milestoneTitle string) (int, error) {
	ms := GetMilestones(ctx, client, org, repo)
	for _, m := range ms {
		if m.GetTitle() == milestoneTitle {
			return *m.Number, nil
		}
	}
	log.Printf("milestoneNotFound=%s\n", milestoneTitle)
	return 0, fmt.Errorf("Milestone not found")
}
