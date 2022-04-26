package githubzen

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/go-github/github"
	"github.ibm.com/sosat/githubzen/limiter"
)

type RepoIDCache struct {
	m    map[string]int64
	lock sync.RWMutex
}

func newRepoIDCache() RepoIDCache {
	return RepoIDCache{m: map[string]int64{}}
}

func (c *RepoIDCache) Read(key string) (value int64, ok bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	value, ok = c.m[key]
	return value, ok
}

func (c *RepoIDCache) Write(key string, value int64) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.m[key] = value
}

var RepoIDs = newRepoIDCache()

func GetRepoIDFromZen(ctx context.Context, client *github.Client, org string, repo string) int64 {
	id, found := RepoIDs.Read(repo)
	if !found {
		// Repo is not in the map, query GitHub for it
		log.Printf("getRepoIDFromZen: repo=%s\n", repo)
		// ghLimiter.Wait()
		// r, resp, err := client.Repositories.Get(ctx, org, repo)
		// ghLimiter.Set(resp.Rate.Limit, resp.Rate.Remaining, GithubTimestampToTime(resp.Rate.Reset))
		// if err != nil {
		// 	log.Println(err)
		// 	return 0
		// }
		r, err := GetRepository(ctx, client, org, repo)
		if err != nil {
			log.Printf("GetRepoIDFromZen error getting repo from GH: %v\n", err)
			return 0
		}
		id = r.GetID()

		// Add to the map
		RepoIDs.Write(repo, id)
	}

	return id
}

type ZenEpicIssues struct {
	Issues []struct {
		RepoID      int    `json:"repo_id"`
		IssueNumber int    `json:"issue_number"`
		IssueURL    string `json:"issue_url"`
	} `json:"epic_issues"`
}

func GetEpics(auth string, repoID int64) (ZenEpicIssues, error) {
	log.Printf("getEpics: repoID=%d\n", repoID)
	zenLimiter.Wait()

	var zen ZenEpicIssues

	u := fmt.Sprintf("https://zenhub.ibm.com/p1/repositories/%d/epics", repoID)

	resp, err := zenPost(u, auth)
	if err != nil {
		log.Println(err)
		return zen, err
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return zen, err
	}

	err = json.Unmarshal(b, &zen)
	if err != nil {
		log.Println(err)
		return zen, err
	}

	log.Printf("zen=%+v\n", zen)

	return zen, nil
}

type ZenEpicData struct {
	Issues []struct {
		RepoID      int `json:"repo_id"`
		IssueNumber int `json:"issue_number"`
		Estimate    struct {
			Value float64 `json:"value"`
		} `json:"estimate"`
		IsEpic   bool `json:"is_epic"`
		Pipeline struct {
			Name string `json:"name"`
		} `json:"pipeline"`
	} `json:"issues"`
	TotalEpicEstimates struct {
		Value float64 `json:"value"`
	} `json:"total_epic_estimates"`
	Estimate struct {
		Value float64 `json:"value"`
	} `json:"estimate"`
}

// GetEpicData returns:
// the total Epic Estimate value (the sum of all the Estimates of Issues contained within the Epic, as well as the Estimate of the Epic itself)
// the Estimate of the Epic
// the name of the Pipeline the Epic is in
// issues belonging to the Epic
//
// For each issue belonging to the Epic:
// issue number
// repo ID
// Estimate value
// is_epic flag (true or false)
// if the issue is from the same repository as the Epic, the ZenHub Board’s Pipeline name (from the repo the Epic is in) is attached.
func GetEpicData(auth, repoID, issue string) (ZenEpicData, error) {
	// curl -H 'X-Authentication-Token: MY_TOKEN' -H 'Content-Type: application/json' https://zenhub.ibm.com/p1/repositories/369922/epics/3

	log.Printf("getEpicData, repoID=%s, issue=%s\n", repoID, issue)

	var zen ZenEpicData

	u := fmt.Sprintf("https://zenhub.ibm.com/p1/repositories/%s/epics/%s", repoID, issue)

	resp, err := zenPost(u, auth)
	if err != nil {
		log.Printf("GetEpicData error: %v\n", err)
		return zen, err
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("GetEpicData error: %v\n", err)
		return zen, err
	}

	err = json.Unmarshal(b, &zen)
	if err != nil {
		log.Printf("GetEpicData error: %v\n", err)
		return zen, err
	}

	log.Printf("zen=%+v\n", zen)

	return zen, nil
}

func GetEstimates(auth, repoID, issueNum string) (float64, float64, float64) {
	log.Printf("GetEstimates: repoID=%s, issueNum=%s\n", repoID, issueNum)

	zen, err := GetEpicData(auth, repoID, issueNum)
	if err != nil {
		log.Println(err)
	}

	var epicEstimate, totalEstimateOpen, totalEstimateClosed float64
	epicEstimate = zen.TotalEpicEstimates.Value

	for _, v := range zen.Issues {
		log.Printf("GetEstimates: pipeline=%s\n", v.Pipeline.Name)
		if "closed" == strings.ToLower(v.Pipeline.Name) {
			totalEstimateClosed += v.Estimate.Value
		} else {
			totalEstimateOpen += v.Estimate.Value
		}
	}

	return epicEstimate, totalEstimateOpen, totalEstimateClosed
}

var zenLimiter = limiter.Limiter{
	Name:      "zen",
	Limit:     100,
	Remaining: 100,
	Threshold: 0.1, // sleep after 10% remaining
}

// When under the rate limit:
// dash 2019/06/11 15:06:51.965476 zenhub.go:84: zenHeader: X-Ratelimit-Limit, [100]
// dash 2019/06/11 15:06:51.965507 zenhub.go:84: zenHeader: X-Ratelimit-Used, [6]
// dash 2019/06/11 15:06:51.965534 zenhub.go:84: zenHeader: Date, [Tue, 11 Jun 2019 19:07:00 GMT]
// dash 2019/06/11 15:06:51.965560 zenhub.go:84: zenHeader: Content-Type, [application/json; charset=utf-8]
// dash 2019/06/11 15:06:51.965584 zenhub.go:84: zenHeader: Content-Length, [61]
// dash 2019/06/11 15:06:51.965608 zenhub.go:84: zenHeader: Access-Control-Allow-Credentials, [true]
// dash 2019/06/11 15:06:51.965640 zenhub.go:84: zenHeader: Etag, [W/"3d-1421463222"]
// dash 2019/06/11 15:06:51.965670 zenhub.go:84: zenHeader: Strict-Transport-Security, [max-age=31536000; includeSubdomains]
// dash 2019/06/11 15:06:51.965696 zenhub.go:84: zenHeader: Server, [nginx]
// dash 2019/06/11 15:06:51.965724 zenhub.go:84: zenHeader: Vary, [Origin]
// dash 2019/06/11 15:06:51.965749 zenhub.go:84: zenHeader: X-Zenhub-Request-Id, [02bc5d9d-7e70-4790-8b1f-47278d00a75a]
// dash 2019/06/11 15:06:51.965773 zenhub.go:84: zenHeader: X-Ratelimit-Reset, [1560280080]

// When exceeded the rate limit:
// dash 2019/06/11 15:28:59.413934 zenhub.go:256: zenHeader: Server, [nginx]
// dash 2019/06/11 15:28:59.413955 zenhub.go:256: zenHeader: Content-Type, [application/json; charset=utf-8]
// dash 2019/06/11 15:28:59.413970 zenhub.go:256: zenHeader: Etag, [W/"3d-1421463222"]
// dash 2019/06/11 15:28:59.413991 zenhub.go:256: zenHeader: X-Ratelimit-Limit, [100]
// dash 2019/06/11 15:28:59.414030 zenhub.go:256: zenHeader: X-Ratelimit-Used, [51]
// dash 2019/06/11 15:28:59.414065 zenhub.go:256: zenHeader: X-Ratelimit-Reset, [1560281400]
// dash 2019/06/11 15:28:59.414094 zenhub.go:256: zenHeader: Date, [Tue, 11 Jun 2019 19:29:08 GMT]
// dash 2019/06/11 15:28:59.414123 zenhub.go:256: zenHeader: Content-Length, [61]
// dash 2019/06/11 15:28:59.414155 zenhub.go:256: zenHeader: Vary, [Origin]
// dash 2019/06/11 15:28:59.414184 zenhub.go:256: zenHeader: Access-Control-Allow-Credentials, [true]
// dash 2019/06/11 15:28:59.414210 zenhub.go:256: zenHeader: X-Zenhub-Request-Id, [80d57d24-f90b-42ba-8c76-8ce3a5ce68df]
// dash 2019/06/11 15:28:59.414238 zenhub.go:256: zenHeader: Strict-Transport-Security, [max-age=31536000; includeSubdomains]

func setZenLimiterFromHeader(lim *limiter.Limiter, header http.Header) {
	var limit, remaining int
	var reset time.Time
	var err error

	// 	for k, v := range header {
	// 		log.Printf("zenHeader: %s, %v\n", k, v)
	// 	}
	// dash 2019/07/02 08:57:06.510726 zenhub.go:242: zenHeader: Date, [Tue, 02 Jul 2019 12:25:16 GMT]
	// dash 2019/07/02 08:57:06.510745 zenhub.go:242: zenHeader: Content-Length, [369]
	// dash 2019/07/02 08:57:06.510765 zenhub.go:242: zenHeader: Vary, [Origin]
	// dash 2019/07/02 08:57:06.510779 zenhub.go:242: zenHeader: Access-Control-Allow-Credentials, [true]
	// dash 2019/07/02 08:57:06.510800 zenhub.go:242: zenHeader: X-Ratelimit-Limit, [100]
	// dash 2019/07/02 08:57:06.510828 zenhub.go:242: zenHeader: X-Ratelimit-Reset, [1562070360]
	// dash 2019/07/02 08:57:06.510850 zenhub.go:242: zenHeader: Strict-Transport-Security, [max-age=31536000; includeSubdomains]
	// dash 2019/07/02 08:57:06.510892 zenhub.go:242: zenHeader: Server, [nginx]
	// dash 2019/07/02 08:57:06.510915 zenhub.go:242: zenHeader: X-Zenhub-Request-Id, [3f9d540d-3d42-4530-9ce2-24b152ebd653]
	// dash 2019/07/02 08:57:06.510934 zenhub.go:242: zenHeader: X-Ratelimit-Used, [7]
	// dash 2019/07/02 08:57:06.510948 zenhub.go:242: zenHeader: Etag, [W/"171-486145404"]
	// dash 2019/07/02 08:57:06.510960 zenhub.go:242: zenHeader: Content-Type, [application/json; charset=utf-8]

	// Limit
	h, found := header["X-Ratelimit-Limit"]
	if found {
		limit, err = strconv.Atoi(h[0])
		if err != nil {
			log.Printf("Could not parse zenhub X-Ratelimit-Limit: %q, err=%v\n", header["X-Ratelimit-Limit"], err)
		}
	} else {
		log.Printf("Could not get zenhub X-Ratelimit-Limit: %q,\n", header["X-Ratelimit-Limit"])
	}

	// Remaining
	var used int
	h, found = header["X-Ratelimit-Used"]
	if found {
		used, err = strconv.Atoi(h[0])
		if err != nil {
			log.Printf("Could not parse zenhub X-Ratelimit-Used: %v\n", header["X-Ratelimit-Used"])
		}
		remaining = limit - used
	} else {
		log.Printf("Could not get zenhub X-Ratelimit-Used: %v\n", header["X-Ratelimit-Used"])
	}

	// Reset
	// "To avoid time differences between your computer and our servers,
	// we suggest to use the Date header in the response to know exactly
	// when the limit is reset."
	var adjust time.Duration
	h, found = header["Date"] // Tue, 02 Jul 2019 12:25:16 GMT
	if found {
		d, err := time.Parse("Mon, 2 Jan 2006 15:04:05 MST", h[0])
		if err == nil {
			adjust = time.Since(d)
		}
	}

	var epoch int64
	h, found = header["X-Ratelimit-Reset"]
	if found {
		epoch, err = strconv.ParseInt(h[0], 10, 64)
		if err != nil {
			log.Printf("Could not parse zenhub X-Ratelimit-Reset: %v\n", header["X-Ratelimit-Reset"])
		}
		reset = time.Unix(epoch, 0)
	} else {
		log.Printf("Could not get zenhub X-Ratelimit-Reset: %v\n", header["X-Ratelimit-Reset"])
	}

	log.Printf("setZenLimiterFromHeader, limit=%d, used=%d, remaining=%d, reset=%s, adjust=%v, reset2=%s\n", limit, used, remaining, reset, adjust, reset.Add(adjust))
	lim.Set(limit, remaining, reset.Add(adjust))
}

func zenPost(url string, auth string) (*http.Response, error) {
	header := http.Header{}
	header.Set("X-Authentication-Token", auth)
	if auth != "" {
		header.Set("Content-Type", "application/json")
	}

	for true {
		log.Printf("zenPost\n")
		zenLimiter.Wait()

		resp, err := httpDoWithRetry("GET", url, nil, header)
		if err != nil {
			return nil, err
		}

		log.Printf("zenPost: statusCode=%d\n", resp.StatusCode)
		setZenLimiterFromHeader(&zenLimiter, resp.Header)

		if resp.StatusCode == 429 || resp.StatusCode == 403 { // ZenHub returns 403 instead of 429
			// zenLimiter.Wait()
			continue
		}

		log.Printf("zenPost return http response\n")
		return resp, nil
	}

	return nil, errors.New("Should never get here")
}

func httpDoWithRetry(method string, url string, body io.Reader, header http.Header) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header = header

	client := &http.Client{
		Timeout: 20 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return resp, err
	}

	return resp, nil
}
