package githubzen

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/google/go-github/github"
)

// CreateLabel creates a label on the given repo
func CreateLabel(ctx context.Context, client *github.Client, org string, repo string, labelName string, labelColor string, labelDescription string) error {
	if org == "" {
		log.Printf("GetLabels Error: org is empty\n")
		return fmt.Errorf("org is empty")
	}

	if repo == "" {
		log.Printf("GetLabels Error: repo is empty\n")
		return fmt.Errorf("repo is empty")
	}

	for {
		ghLimiter.Wait()

		label := github.Label{Name: &labelName, Color: &labelColor, Description: &labelDescription}
		l, resp, err := client.Issues.CreateLabel(ctx, org, repo, &label)
		log.Printf("label=%v, resp=%v, err=%v\n", l, resp, err)

		ghLimiter.Set(resp.Rate.Limit, resp.Rate.Remaining, GithubTimestampToTime(resp.Rate.Reset))

		if isRatelimitError(resp, err) {
			log.Printf("listLabels: RateLimitError: %#v\n", err)
			continue
		}

		return err
	}
}

// EditLabel updates a label on the given repo
func EditLabel(ctx context.Context, client *github.Client, org string, repo string, labelName string, labelColor string, labelDescription string) error {
	if org == "" {
		log.Printf("GetLabels Error: org is empty\n")
		return fmt.Errorf("org is empty")
	}

	if repo == "" {
		log.Printf("GetLabels Error: repo is empty\n")
		return fmt.Errorf("repo is empty")
	}

	for {
		ghLimiter.Wait()

		label := github.Label{Name: &labelName, Color: &labelColor, Description: &labelDescription}
		l, resp, err := client.Issues.EditLabel(ctx, org, repo, labelName, &label)
		log.Printf("label=%v, resp=%v, err=%v\n", l, resp, err)

		ghLimiter.Set(resp.Rate.Limit, resp.Rate.Remaining, GithubTimestampToTime(resp.Rate.Reset))

		if isRatelimitError(resp, err) {
			log.Printf("listLabels: RateLimitError: %#v\n", err)
			continue
		}

		return err
	}
}

// DeleteLabel removes a label on the given repo
func DeleteLabel(ctx context.Context, client *github.Client, org string, repo string, labelName string) error {
	if org == "" {
		log.Printf("GetLabels Error: org is empty\n")
		return fmt.Errorf("org is empty")
	}

	if repo == "" {
		log.Printf("GetLabels Error: repo is empty\n")
		return fmt.Errorf("repo is empty")
	}

	for {
		ghLimiter.Wait()

		resp, err := client.Issues.DeleteLabel(ctx, org, repo, labelName)
		log.Printf("label=%v, resp=%v, err=%v\n", labelName, resp, err)

		ghLimiter.Set(resp.Rate.Limit, resp.Rate.Remaining, GithubTimestampToTime(resp.Rate.Reset))

		if isRatelimitError(resp, err) {
			log.Printf("listLabels: RateLimitError: %#v\n", err)
			continue
		}

		return err
	}
}

// GetLabels gets all labels for a given repository.
func GetLabels(ctx context.Context, client *github.Client, org string, repo string) []string {
	allLabels := []string{}

	if org == "" {
		log.Printf("GetLabels Error: org is empty\n")
		return allLabels
	}

	if repo == "" {
		log.Printf("GetLabels Error: repo is empty\n")
		return allLabels
	}

	opt := &github.ListOptions{
		PerPage: 1000,
	}

	for {
		// ghLimiter.Wait()
		// labels, resp, err := client.Issues.ListLabels(ctx, org, repo, opt)
		// log.Printf("getLabels: org=%s, repo=%s, resp=%+v, err=%v\n", org, repo, resp, err)
		// ghLimiter.Set(resp.Rate.Limit, resp.Rate.Remaining, GithubTimestampToTime(resp.Rate.Reset))
		labels, resp, err := listLabels(ctx, client, org, repo, opt)

		if err != nil {
			// log.Fatal(err)
			log.Printf("Error: %v\n", err)
			return allLabels
		}

		for _, label := range labels {
			allLabels = append(allLabels, label.GetName())
		}

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
		log.Printf("getLabels: page=%d\n", opt.Page)
	}

	log.Printf("getLabels: org=%q, repo=%q numLabels=%d", org, repo, len(allLabels))

	return allLabels
}

func listLabels(ctx context.Context, client *github.Client, org string, repo string, opt *github.ListOptions) ([]*github.Label, *github.Response, error) {
	var labels []*github.Label
	var resp *github.Response
	var err error

	for {
		ghLimiter.Wait()
		labels, resp, err = client.Issues.ListLabels(ctx, org, repo, opt)
		log.Printf("listLabels: org=%s, repo=%s, resp=%+v, err=%v\n", org, repo, resp, err)
		ghLimiter.Set(resp.Rate.Limit, resp.Rate.Remaining, GithubTimestampToTime(resp.Rate.Reset))

		// if err, ok := err.(*github.RateLimitError); ok {
		if isRatelimitError(resp, err) {
			log.Printf("listLabels: RateLimitError: %#v\n", err)
			continue
		}

		return labels, resp, err
	}
}

func hasLabel(ctx context.Context, client *github.Client, org string, repoID int, issueNum int, target string) bool {
	repoName := GetRepoName(ctx, client, org, int64(repoID))
	issue := GetIssue(ctx, client, org, repoName, issueNum)

	for _, label := range issue.Labels {
		if strings.ToLower(target) == strings.ToLower(label.GetName()) {
			log.Printf("hasLabel: label=%s, found=true\n", target)
			return true
		}
	}

	log.Printf("hasLabel: label=%s, found=false\n", target)
	return false
}
