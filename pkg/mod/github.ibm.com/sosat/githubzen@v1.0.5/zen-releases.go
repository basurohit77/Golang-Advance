package githubzen

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"time"
)

type ZenReleases []struct {
	ReleaseID      string      `json:"release_id"`
	Title          string      `json:"title"`
	Description    string      `json:"description"`
	StartDate      time.Time   `json:"start_date"`
	DesiredEndDate time.Time   `json:"desired_end_date"`
	CreatedAt      time.Time   `json:"created_at"`
	ClosedAt       interface{} `json:"closed_at"`
	State          string      `json:"state"`
}

type ZenReleaseIssues []struct {
	RepoID      int `json:"repo_id,omitempty"`
	IssueNumber int `json:"issue_number,omitempty"`
}

// GetReleases gets the release reports for a repository
func GetReleases(auth string, repoID int64) (ZenReleases, error) {
	// # curl -H "X-Authentication-Token: $tok" -H 'Content-Type: application/json' https://zenhub.ibm.com/p1/repositories/106331/reports/releases
	// # output: [{"release_id":"5de56d365a9f1a78736747d2","title":"TIP Q1/20","description":"This release will include all the TIP work we plan on getting done in Q1 of 2020","start_date":"2020-01-01T17:00:00.000Z","desired_end_date":"2020-03-31T16:00:00.000Z","created_at":"2019-12-02T19:59:50.737Z","closed_at":null,"state":"open"},{"release_id":"5e14e1fd0f7b4c082d6f1ba8","title":"TIP Q2/20","description":"","start_date":"2020-04-01T16:00:00.000Z","desired_end_date":"2020-06-30T04:00:00.000Z","created_at":"2020-01-07T19:54:37.452Z","closed_at":null,"state":"open"},{"release_id":"5e14e24510a6ec08148ec76a","title":"TIP Q3/20","description":"","start_date":"2020-07-01T16:00:00.000Z","desired_end_date":"2020-09-30T16:00:00.000Z","created_at":"2020-01-07T19:55:49.077Z","closed_at":null,"state":"open"},{"release_id":"5e14e2660f7b4c082d6f1bae","title":"TIP Q4/20","description":"","start_date":"2020-10-01T16:00:00.000Z","desired_end_date":"2020-12-31T17:00:00.000Z","created_at":"2020-01-07T19:56:22.042Z","closed_at":null,"state":"open"}]

	log.Printf("GetReleases: repoID=%d", repoID)

	var zen ZenReleases

	zenLimiter.Wait()

	u := fmt.Sprintf("https://zenhub.ibm.com/p1/repositories/%d/reports/releases", repoID)

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
		log.Printf("GetReleases: body=%s, err=%v\n", string(b), err)
		return zen, err
	}

	// log.Printf("GetReleases=%+v, body=%s\n", zen, string(b))
	setZenLimiterFromHeader(&zenLimiter, resp.Header)

	return zen, nil
}

// GetReleaseIssues gets all the issues for a release report
func GetReleaseIssues(auth string, releaseID string) (ZenReleaseIssues, error) {
	// curl -H "X-Authentication-Token: $tok" -H 'Content-Type: application/json' https://zenhub.ibm.com/p1/reports/release/5de56d365a9f1a78736747d2/issues
	// # output: [{"repo_id":106331,"issue_number":3829},{"repo_id":106331,"issue_number":4107},{"repo_id":106331,"issue_number":5100},{"repo_id":106331,"issue_number":5316},{"repo_id":106331,"issue_number":5491},{"repo_id":106331,"issue_number":5525},{"repo_id":106331,"issue_number":6846},{"repo_id":106331,"issue_number":7380},{"repo_id":106331,"issue_number":7735},{"repo_id":106331,"issue_number":7736},{"repo_id":106331,"issue_number":7755},{"repo_id":106331,"issue_number":7823},{"repo_id":106331,"issue_number":7969},{"repo_id":106331,"issue_number":8116},{"repo_id":106331,"issue_number":8157},{"repo_id":106331,"issue_number":8320},{"repo_id":106331,"issue_number":8329},{"repo_id":106331,"issue_number":8348},{"repo_id":106331,"issue_number":8404},{"repo_id":106331,"issue_number":8486},{"repo_id":106331,"issue_number":8487},{"repo_id":106331,"issue_number":8506},{"repo_id":106331,"issue_number":8546},{"repo_id":106331,"issue_number":8557},{"repo_id":106331,"issue_number":8559},{"repo_id":106331,"issue_number":8581},{"repo_id":106331,"issue_number":8584},{"repo_id":106331,"issue_number":8585},{"repo_id":106331,"issue_number":8586},{"repo_id":106331,"issue_number":8589},{"repo_id":106331,"issue_number":8591},{"repo_id":106331,"issue_number":8598},{"repo_id":106331,"issue_number":8601},{"repo_id":106331,"issue_number":8602},{"repo_id":585230,"issue_number":38}]

	log.Printf("GetReleaseIssues: releaseID=%s", releaseID)

	var zen ZenReleaseIssues

	zenLimiter.Wait()

	u := fmt.Sprintf("https://zenhub.ibm.com/p1/reports/release/%s/issues", releaseID)

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
		log.Printf("GetReleaseIssues: body=%s, err=%v\n", string(b), err)
		return zen, err
	}

	log.Printf("GetReleaseIssues=%+v, body=%s\n", zen, string(b))
	setZenLimiterFromHeader(&zenLimiter, resp.Header)

	return zen, nil
}
