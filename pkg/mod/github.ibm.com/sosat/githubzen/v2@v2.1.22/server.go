package githubzen

import (
	"context"
	"log"
	"time"

	"github.com/google/go-github/github"
	"github.com/gregjones/httpcache"
	"github.ibm.com/sosat/githubzen/v2/rate"
	"github.ibm.com/sosat/githubzen/v2/utils"
	"golang.org/x/oauth2"
)

// Server is a basic context for github and zen up
// It contains the basic setup information
type Server struct {
	URL     string
	Client  *github.Client
	Context context.Context
	Limiter *rate.Limiter
	Repos   map[string]*Repo
}

// SetupGit is the first function called to create the connection to the server
// Pass the URL to the github server and an access token (personal access token)
func SetupGit(url, accessToken string) (*Server, error) {

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: accessToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	// Insert a caching transport
	cachingTransport := httpcache.NewMemoryCacheTransport()
	cachingTransport.MarkCachedResponses = true // We want responses to indicate if cache hit
	cachingTransport.Transport = tc.Transport
	tc.Transport = cachingTransport

	client, err := github.NewEnterpriseClient(url, url, tc)
	if err != nil {
		log.Println("Unable to create GHE client:", err)
		return nil, err
	}

	limit := 5000
	remaining := 5000
	reset := time.Now()
	limits, resp, err := client.RateLimits(ctx)
	if err1 := utils.IsSuccessfulRequest(resp, err); err1 != nil {
		log.Println("Unable to retrieve rate limits", err1)
		log.Println("Assuming defaults")
	} else {
		limit = limits.Core.Limit
		remaining = limits.Core.Remaining
		reset = utils.GithubTimestampToTime(limits.Core.Reset)
	}

	l := rate.NewLimiter(limit, remaining, reset)

	repoMap := make(map[string]*Repo)

	return &Server{URL: url, Client: client, Context: ctx, Limiter: l, Repos: repoMap}, nil
}
