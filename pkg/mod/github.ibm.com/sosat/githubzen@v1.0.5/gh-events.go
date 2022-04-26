package githubzen

/*
// This way of getting story points is not accurate as this only gets you credit if you close the issue.
// If someone else closes the issue, it does not show in your events.
// Instead, lets try getting the closed issues for a milestone and going thru that list.
// See GetClosedEstimatesForUser()
func GetEvents(ctx context.Context, client *github.Client, zenAuth string, user string) *github.Issue {
	// var resp *github.Response
	// var err error

	for {

		// issue, found := issuesCache.read(issueNum)
		// if found {
		// 	log.Printf("getIssue: cacheHit=true, issue=%+v, resp=%+v, err=%v\n", issue.GetTitle(), resp, err)
		// 	return issue
		// }
		opt := &github.ListOptions{
			// Milestone
			// State:       "all", // open | closed | all
			// Labels:      []string{label},
			// ListOptions: github.ListOptions{PerPage: 1000},
			PerPage: 1000,
		}

		ghLimiter.Wait()
		// issue, resp, err = client.Issues.Get(ctx, owner, repo, issueNum)
		events, resp, err := client.Activity.ListEventsPerformedByUser(ctx, user, false, opt)
		ghLimiter.Set(resp.Rate.Limit, resp.Rate.Remaining, GithubTimestampToTime(resp.Rate.Reset))

		// if err, ok := err.(*github.RateLimitError); ok {
		if isRatelimitError(resp, err) {
			log.Printf("GetIssue: RateLimitError: %#v\n", err)
			continue
		}

		// issuesCache.write(issueNum, issue)

		// log.Printf("getIssue: cacheHit=false, issue=%+v, resp=%+v, err=%v\n", issue.GetTitle(), resp, err)

		// return issue

		// log.Printf("GetEvents: %+v\n", events)
		// i := 0
		for i := range events {
			// log.Printf("GetEvents: type=%+v, org=%+v, raw=%s\n", events[i].GetType(), events[i].GetOrg(), events[i].GetRawPayload())

			type Issue struct {
				URL           string `json:"url"`
				RepositoryURL string `json:"repository_url"`
				Number        int    `json:"number"`
			}

			// type Repository struct {
			// 	ID   int    `json:"id"`
			// 	Name string `json:"name"`
			// }

			type Event struct {
				Action string `json:"action"`
				Issue  Issue  `json:"issue"`
				// Repository Repository `json:"repository"`
			}

			var e Event
			err := json.Unmarshal(events[i].GetRawPayload(), &e)
			if err != nil {
				log.Println(err)
			}

			log.Printf("GetEvents: type=%+v, org=%+v, action=%s, issueNumber=%d, repoURL=%s\n", events[i].GetType(), events[i].GetOrg(), e.Action, e.Issue.Number, e.Issue.RepositoryURL)

			if "IssuesEvent" == events[i].GetType() && "closed" == e.Action {
				// When?
				// GetEstimates(auth, repoID, issueNum)
				// issue, err := getZenIssue(zenAuth, repoID, issueNum)

				// Get story points for closed issues
				org, repoName, err := parseRepo(e.Issue.RepositoryURL)
				if err != nil {
					log.Println(err)
				} else {
					log.Printf("org=%s, repoName=%s\n", org, repoName)
					repoID := GetRepoIDFromZen(ctx, client, org, repoName)
					issue, err := getZenIssue(zenAuth, repoID, e.Issue.Number)
					if err != nil {
						log.Println(err)
					}
					log.Printf("zenIssue=%+v\n", issue)

					// If not an epic, count the story points
					// TODO
				}
			}

		}
		return nil

	}
}

func parseRepo(url string) (org string, repoName string, err error) {
	// input: https://github.ibm.com/api/v3/repos/sosat/general
	// split: [https:  github.ibm.com api v3 repos cloud-sre osscatalog] // careful, one is blank

	s := strings.Split(url, "/")
	log.Printf("%+v, len=%d\n", s, len(s))

	if len(s) != 8 {
		return org, repoName, fmt.Errorf("Invalid URL %s", url)
	}

	org = s[6]
	repoName = s[7]

	return org, repoName, nil
}
*/
