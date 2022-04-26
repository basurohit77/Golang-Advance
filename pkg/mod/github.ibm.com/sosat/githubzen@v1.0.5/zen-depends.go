package githubzen

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io/ioutil"
	"log"
	"strconv"
	"strings"

	"github.com/google/go-github/github"
)

type Deps struct {
	Dependencies []struct {
		Blocking struct {
			IssueNumber int `json:"issue_number"`
			RepoID      int `json:"repo_id"`
		} `json:"blocking"`
		Blocked struct {
			IssueNumber int `json:"issue_number"`
			RepoID      int `json:"repo_id"`
		} `json:"blocked"`
	} `json:"dependencies"`
}

func GetDependencies(ctx context.Context, client *github.Client, org string, repos []string, auth string) (Deps, error) {
	var deps Deps

	for _, repo := range repos {
		repoID := GetRepoIDFromZen(ctx, client, org, repo)
		log.Printf("GetDependencies, repo=%q, repoID=%d\n", repo, repoID)

		u := fmt.Sprintf("https://zenhub.ibm.com/p1/repositories/%d/dependencies", repoID)

		resp, err := zenPost(u, auth)
		if err != nil {
			return deps, err
		}
		defer resp.Body.Close()

		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return deps, err
		}
		var d Deps
		err = json.Unmarshal(b, &d)
		if err != nil {
			return deps, err
		}
		log.Printf("repo=%q, deps=%+v\n", repo, d)
		deps.Dependencies = append(deps.Dependencies, d.Dependencies...)
	}

	log.Printf("all deps=%+v\n", deps)
	return deps, nil
}

func GetBlocked(ctx context.Context, client *github.Client, zenAuth string, org string, deps Deps, targetLabel string) string {
	// digraph graphname {
	// 	rankdir=LR
	// 	issue_1077 [label="1077\nBuild correlations database - Phase 2\nEstimate: 5\nMilestone: 2017-08"];
	// 	838 -> issue_1077;
	// 	839 -> issue_1077;
	// 	1077 -> 840;

	buf := strings.Builder{}

	buf.WriteString("digraph graphname {\nrankdir=RL\n")

	labeled := map[string]bool{} // The key is "repo_number:issue_number".  Existential

	for _, d := range deps.Dependencies {
		// Check for target label (like GaaS)
		foundLabel := false
		if targetLabel != "" && hasLabel(ctx, client, org, d.Blocking.RepoID, d.Blocking.IssueNumber, targetLabel) {
			foundLabel = true
		}

		if targetLabel != "" && hasLabel(ctx, client, org, d.Blocked.RepoID, d.Blocked.IssueNumber, targetLabel) {
			foundLabel = true
		}

		if !foundLabel {
			continue
		}

		// Write dependencies
		blockingIssueNum := strconv.Itoa(d.Blocking.IssueNumber)
		blockingRepoID := strconv.Itoa(d.Blocking.RepoID)

		blockedIssueNum := strconv.Itoa(d.Blocked.IssueNumber)
		blockedRepoID := strconv.Itoa(d.Blocked.RepoID)

		buf.WriteString("repo_" + blockedRepoID + "_issue_" + blockedIssueNum + " -> " + "repo_" + blockingRepoID + "_issue_" + blockingIssueNum + ";\n")

		// Write labels
		s := getDotInfo(ctx, client, zenAuth, org, d.Blocked.RepoID, d.Blocked.IssueNumber)
		key := fmt.Sprintf("%d:%d", d.Blocked.RepoID, d.Blocked.IssueNumber)
		if _, ok := labeled[key]; !ok {
			buf.WriteString(s)
			labeled[key] = true
		}

		s = getDotInfo(ctx, client, zenAuth, org, d.Blocking.RepoID, d.Blocking.IssueNumber)
		key = fmt.Sprintf("%d:%d", d.Blocking.RepoID, d.Blocking.IssueNumber)
		if _, ok := labeled[key]; !ok {
			buf.WriteString(s)
			labeled[key] = true
		}
	}

	buf.WriteString("}\n")

	return buf.String()
}

func getDotInfo(ctx context.Context, client *github.Client, zenAuth string, org string, repoID int, issueNum int) string {
	repoName := GetRepoName(ctx, client, org, int64(repoID))
	issue := GetIssue(ctx, client, org, repoName, int(issueNum)) // get issue from github
	log.Printf("issueNum=%d, state=%q, assignee=%q, title=%q\n", issueNum, issue.GetState(), issue.GetAssignee().GetLogin(), issue.GetTitle())

	title := issue.GetTitle() //+ " (" + issue.GetState() + ")"
	title = html.EscapeString(title)

	state := issue.GetState()
	state = strings.ToLower(state)

	// Debug: check for unexpected states
	if state != "open" && state != "closed" {
		log.Printf("unexpectedState=%s\n", state)
	}

	milestone := issue.GetMilestone().GetTitle()

	var aa []string
	for _, v := range issue.Assignees {
		aa = append(aa, v.GetLogin())
	}
	assignees := strings.Join(aa, ", ")
	if assignees == "" {
		assignees = "<none>"
	}

	blocked := false
	for _, l := range issue.Labels {
		if strings.ToLower(l.GetName()) == "blocked" {
			blocked = true
		}
	}

	var pipeline string
	zenIssue, err := GetZenIssue(zenAuth, int64(repoID), issueNum)
	if err != nil {
		log.Printf("error getting zen issue for pipeline: %v\n", err)
		pipeline = ""
	}
	pipeline = zenIssue.Pipeline.Name

	var attr string
	if state == "closed" {
		attr = ", style=filled"
	} else {
		// If open, then color based on the milestone
		// if milestone != "" {
		// 	c := getColor(milestone)
		// 	attr = attr + ", style=filled,color=\"" + c + "\""
		// }
		// If open, then color based on the pipeline
		if pipeline != "" {
			c := getColor(pipeline)
			if blocked {
				attr = attr + ", style=filled,fillcolor=\"" + c + "\",color=red,penwidth=5"
			} else {
				attr = attr + ", style=filled,color=\"" + c + "\""
			}
		} else {
			log.Printf("empty pipeline for issue %d\n", issueNum)
		}
	}

	if zenIssue.IsEpic {
		attr = attr + ", shape=box"
	}

	// Add the URL to the issue
	attr = fmt.Sprintf("%s, URL=\"https://github.ibm.com/%s/%s/issues/%d\", target=\"_blank\"", attr, org, repoName, issueNum)

	if milestone == "" {
		milestone = "<none>"
	}

	s := fmt.Sprintf("repo_%d_issue_%d [label=\"%s %d\\n%s\\nAssignees: %s\\nMilestone: %s\\nPipeline: %s\"%s];\n",
		repoID, issueNum, repoName, issueNum, title, assignees, milestone, pipeline, attr)

	return s
}

// var colorMap = map[string]string{}
var colorMap = map[string]string{
	"Closed":      "grey",
	"Done":        "grey",
	"Icebox":      "#00b7df", // blue
	"Review/QA":   "#e9c400", // yellow
	"New Issues":  "#fc7d53", // orange
	"In Progress": "#8edf4a", // green
	"Backlog":     "#cf90de", // purple
}
var colorIndex = 0

func getColor(key string) string {
	// var data = google.visualization.arrayToDataTable([
	//   ['City', '2010 Population', { role: 'style' },  { role: 'annotation' }],
	//   ['BLUEMIX_IDS', 8175000, 'color: #00b7df', 888],
	//   ['CONNECT', 3792000, 'color: #e9c400', 333],
	//   ['CSI', 2695000, 'color: #fc7d53', 222],
	//   ['EMPTORIS', 2099000, 'color: #8edf4a', 222],
	//   ['NOC', 1526000, 'color: #cf90de', 111]
	// ]);

	// colors := []string{
	// 	"#00b7df", // blue
	// 	"#e9c400", // yellow
	// 	"#fc7d53", // orange
	// 	"#8edf4a", // green
	// 	"#cf90de", // purple
	// }

	c, ok := colorMap[key]
	if ok {
		log.Printf("getColor key=%q, color=%q\n", key, c)
		log.Printf("getColor mapLen=%d\n", len(colorMap))
		return c
	}

	// A new key, select the next available color
	// i := len(colorMap) % len(colors)
	// c = colors[i]
	//
	// Hard code for unknown keys
	i := 0
	c = "red"

	//
	colorMap[key] = c
	log.Printf("getColor i=%d, key=%q, color=%q\n", i, key, c)

	log.Printf("getColor mapLen=%d\n", len(colorMap))
	return c
}
