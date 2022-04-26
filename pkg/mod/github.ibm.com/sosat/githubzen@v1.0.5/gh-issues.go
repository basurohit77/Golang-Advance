package githubzen

import (
	"context"
	"log"
	"sync"

	"github.com/google/go-github/github"
)

var issuesCache = newIssuesCache()

func GetIssue(ctx context.Context, client *github.Client, owner string, repo string, issueNum int) *github.Issue {
	var resp *github.Response
	var err error

	for {
		issue, found := issuesCache.read(issueNum)
		if found {
			log.Printf("getIssue: cacheHit=true, issue=%+v, resp=%+v, err=%v\n", issue.GetTitle(), resp, err)
			return issue
		}

		ghLimiter.Wait()
		issue, resp, err = client.Issues.Get(ctx, owner, repo, issueNum)
		if resp != nil {
			ghLimiter.Set(resp.Rate.Limit, resp.Rate.Remaining, GithubTimestampToTime(resp.Rate.Reset))
		}

		// if err, ok := err.(*github.RateLimitError); ok {
		if isRatelimitError(resp, err) {
			log.Printf("GetIssue: RateLimitError: %#v\n", err)
			continue
		}

		issuesCache.write(issueNum, issue)

		log.Printf("getIssue: cacheHit=false, issue=%+v, resp=%+v, err=%v\n", issue.GetTitle(), resp, err)

		return issue
	}
}

func GetIssues(ctx context.Context, client *github.Client, org string, repo string, opt *github.IssueListByRepoOptions, limit int) (allIssues []*github.Issue) {
	for {
		// ghLimiter.Wait()
		// issues, resp, err := client.Issues.ListByRepo(ctx, org, repo, opt)
		issues, resp, err := listByRepo(ctx, client, org, repo, opt)
		log.Printf("getIssues: resp=%+v, err=%v\n", resp, err)
		// ghLimiter.Set(resp.Rate.Limit, resp.Rate.Remaining, GithubTimestampToTime(resp.Rate.Reset))

		if err != nil {
			log.Println(err)
			return
		}

		allIssues = append(allIssues, issues...)

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
		log.Printf("getIssues: page=%d\n", opt.Page)

		if limit != 0 && len(allIssues) >= limit {
			log.Printf("getIssues: reached limit numIssues=%d, limit=%d\n", len(allIssues), limit)
			break
		}
	}

	log.Printf("getIssues: org=%q, repo=%q numIssues=%d", org, repo, len(allIssues))

	return allIssues
}

func listByRepo(ctx context.Context, client *github.Client, org string, repo string, opt *github.IssueListByRepoOptions) ([]*github.Issue, *github.Response, error) {
	var issues []*github.Issue
	var resp *github.Response
	var err error

	for {
		ghLimiter.Wait()

		issues, resp, err = client.Issues.ListByRepo(ctx, org, repo, opt)
		log.Printf("listByRepo: resp=%+v, err=%v\n", resp, err)

		if resp != nil {
			ghLimiter.Set(resp.Rate.Limit, resp.Rate.Remaining, GithubTimestampToTime(resp.Rate.Reset))
		}

		// if err, ok := err.(*github.RateLimitError); ok {
		if isRatelimitError(resp, err) {
			log.Printf("listByRepo: RateLimitError: %#v\n", err)
			continue
		}

		break
	}

	return issues, resp, err
}

type IssuesCache struct {
	m    map[int]*github.Issue
	lock sync.RWMutex
}

func newIssuesCache() IssuesCache {
	return IssuesCache{m: map[int]*github.Issue{}}
}

func (c *IssuesCache) read(key int) (issue *github.Issue, ok bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	issue, ok = c.m[key]
	return issue, ok
}

func (c *IssuesCache) write(key int, value *github.Issue) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.m[key] = value
}

func (c *IssuesCache) invalidate() {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.m = map[int]*github.Issue{}
}
