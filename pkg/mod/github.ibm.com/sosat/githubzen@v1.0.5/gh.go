package githubzen

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/go-github/github"
	"github.ibm.com/sosat/githubzen/limiter"

	"golang.org/x/oauth2"
)

var ghLimiter = limiter.Limiter{
	Name:      "github",
	Limit:     5000,
	Remaining: 5000,
	Threshold: 0.1, // sleep after 10% remaining
}

// type GitHubClient struct {
// 	Client *github.Client,
// 	ZenAuth string,
// }

// var GHClient GitHubClient{}

func SetupGitHub(accessToken string) (context.Context, *github.Client, string) {
	const gheURL = "https://github.ibm.com/api/v3/"

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: accessToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	client, err := github.NewEnterpriseClient(gheURL, gheURL, tc)
	if err != nil {
		panic(fmt.Sprintf("Unable to create GHE client: %v", err))
	}

	zenAuth := os.Getenv("DASH_ZENHUB_ACCESS_TOKEN")

	// GHClient.Client = client
	// GHClient.ZenAuth = zenAuth

	return ctx, client, zenAuth
}

func GithubTimestampToTime(ts github.Timestamp) time.Time {
	e := ts.UTC().Unix()
	t := time.Unix(e, 0)
	return t
}

func InvalidateCaches() {
	log.Printf("InvalidateCaches")

	issuesCache.invalidate()
	reposCache.invalidate()

	zenIssuesCache.invalidate()
	// do not invalidate zen RepoIDs cache
}

func isRatelimitError(resp *github.Response, err error) bool {
	if err, ok := err.(*github.RateLimitError); ok {
		log.Printf("isRatelimitError: %#v\n", err)
		return true
	}

	if err != nil && resp != nil && resp.StatusCode == 403 {
		log.Printf("isRatelimitError: statusCode=%d, err: %#v\n", resp.StatusCode, err)

		// 2019/07/01 18:37:04.649008 gh-repos.go:153: getRepository: statusCode=403, err: &github.ErrorResponse{Response:(*http.Response)(0xc000180120), Message:"", Errors:[]github.Error(nil), Block:(*struct { Reason string "json:\"reason,omitempty\""; CreatedAt *github.Timestamp "json:\"created_at,omitempty\"" })(nil), DocumentationURL:""}

		// 2019/07/01 18:37:04.649008 gh-repos.go:153: getRepository: statusCode=403,
		// err: &github.ErrorResponse{Response:(*http.Response)(0xc000180120),
		// Message:"",
		// Errors:[]github.Error(nil),
		// Block:(*struct { Reason string "json:\"reason,omitempty\""; CreatedAt *github.Timestamp "json:\"created_at,omitempty\"" })(nil),
		// DocumentationURL:""}

		// 2019/07/01 18:30:57.487500 gh-repos.go:154: GET https://github.ibm.com/api/v3/repos/cloud-sre/pnp-abstraction: 403  []

		return true
	}

	return false
}
