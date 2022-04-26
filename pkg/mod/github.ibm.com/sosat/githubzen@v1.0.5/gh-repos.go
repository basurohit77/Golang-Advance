package githubzen

import (
	"context"
	"log"
	"sync"

	"github.com/google/go-github/github"
	"github.ibm.com/sosat/githubzen/slice"
)

var reposCache = newReposCache()

type ReposCache struct {
	m    map[string][]*github.Repository
	lock sync.RWMutex
}

func newReposCache() ReposCache {
	return ReposCache{m: map[string][]*github.Repository{}}
}

func (c *ReposCache) read(key string) (value []*github.Repository, ok bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	value, ok = c.m[key]
	return value, ok
}

func (c *ReposCache) write(key string, value []*github.Repository) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.m[key] = value
}

func (c *ReposCache) invalidate() {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.m = map[string][]*github.Repository{}
}

func GetRepositoriesByOrg(ctx context.Context, client *github.Client, org string) (allRepos []*github.Repository, err error) {
	log.Printf("getRepositoriesByOrg: org=%s\n", org)

	allRepos, found := reposCache.read(org)
	if found {
		// log.Printf("getRepositoriesByOrg: returning from cache")
		return allRepos, nil
	}

	opt := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: 500},
	}
	for {
		repos, resp, err := listByOrg(ctx, client, org, opt) // rate limits handled by listByOrg()
		if err != nil {
			log.Printf("Unable to get repos: %v", err)
			return nil, err
		}
		allRepos = append(allRepos, repos...)

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
		log.Printf("getRepositoriesByOrg: page=%d\n", opt.Page)
	}

	log.Printf("getRepositoriesByOrg: org=%q, numRepos=%d", org, len(allRepos))

	reposCache.write(org, allRepos)

	return allRepos, nil
}

func listByOrg(ctx context.Context, client *github.Client, org string, opt *github.RepositoryListByOrgOptions) ([]*github.Repository, *github.Response, error) {
	for {
		ghLimiter.Wait()
		repos, resp, err := client.Repositories.ListByOrg(ctx, org, opt)
		// log.Printf("getRepositoriesByOrg: numRepos=%d, resp=%+v, err=%v\n", len(repos), resp, err)
		log.Printf("getRepositoriesByOrg: numRepos=%d, resp=%+v, err=%+v\n", len(repos), resp, err)
		log.Printf("getRepositoriesByOrg: numRepos=%d, resp=%#v, err=%#v\n", len(repos), resp, err)
		// 2019/06/24 17:12:25.182218 gh-repos.go:29: getRepositoriesByOrg: numRepos=100, resp=github.Rate{Limit:5000, Remaining:492, Reset:github.Timestamp{2019-06-24 17:46:36 -0400 EDT}}, err=<nil>
		// 2019/06/24 17:12:25.182236 gh-repos.go:30: getRepositoriesByOrg: numRepos=100, resp=github.Rate{Limit:5000, Remaining:492, Reset:github.Timestamp{2019-06-24 17:46:36 -0400 EDT}}, err=<nil>
		// 2019/06/24 17:12:25.182251 gh-repos.go:31: getRepositoriesByOrg: numRepos=100, resp=&github.Response{Response:(*http.Response)(0xc00012c900), NextPage:2, PrevPage:0, FirstPage:0, LastPage:3, Rate:github.Rate{Limit:5000, Remaining:492, Reset:github.Timestamp{Time:time.Time{wall:0x0, ext:63697009596, loc:(*time.Location)(0x16a5de0)}}}}, err=<nil>

		ghLimiter.Set(resp.Rate.Limit, resp.Rate.Remaining, GithubTimestampToTime(resp.Rate.Reset))

		if isRatelimitError(resp, err) {
			log.Printf("listByOrg: ratelimiterr=%#v\n", err)
			continue
		}

		return repos, resp, err
	}
}

// TODO: Working to get old repos archived
// TODO: Ignore archived repos
var reposToIgnore = []string{"compass_vuln_scan_output", "OSS-one-cloud-doctor", "OSS-one-cloud-tip"}

// GetRepoListForOrg returns a list of repositories for the given organization.
func GetRepoListForOrg(ctx context.Context, client *github.Client, org string) []string {
	repos := []string{}

	r, err := GetRepositoriesByOrg(ctx, client, org)
	if err != nil {
		log.Printf("getRepoListForOrg: Unable to get repositories for org=%s, err=%v", org, err)
		return repos
	}
	for _, v := range r {
		// Filter out archived repos
		if v.GetArchived() {
			continue
		}

		name := v.GetName()

		// Filter out specific repos
		_, found := slice.Find(reposToIgnore, name)
		if found {
			continue
		}

		repos = append(repos, name)
	}

	return repos
}

// GetRepoName returns the name of a repository given the repository ID.
func GetRepoName(ctx context.Context, client *github.Client, org string, repoID int64) string {
	repos, err := GetRepositoriesByOrg(ctx, client, org)
	if err != nil {
		log.Printf("GetRepoName: err=%v\n", err)
		return ""
	}

	for _, repo := range repos {
		if repo.GetID() == repoID {
			log.Printf("GetRepoName: repoID=%d, repoName=%q\n", repoID, repo.GetName())
			return repo.GetName()
		}
	}

	log.Printf("GetRepoName: unknownRepoID=%d\n", repoID)
	return ""
}

func GetRepository(ctx context.Context, client *github.Client, org string, repoName string) (*github.Repository, error) {
	for {
		ghLimiter.Wait()
		r, resp, err := client.Repositories.Get(ctx, org, repoName)
		ghLimiter.Set(resp.Rate.Limit, resp.Rate.Remaining, GithubTimestampToTime(resp.Rate.Reset))

		if isRatelimitError(resp, err) {
			log.Printf("getRepository: statusCode=%d, err: %#v\n", resp.StatusCode, err)
			continue
		}

		if err != nil {
			log.Printf("getRepository: err: %#v\n", err)
			log.Println(err)
			return r, err
		}
		log.Printf("getRepository: err: %#v\n", err)

		return r, nil
	}
}
