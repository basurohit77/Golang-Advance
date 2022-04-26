package githubzen

import (
	"errors"
	"log"

	"github.com/google/go-github/github"
	"github.ibm.com/sosat/githubzen/v2/utils"
)

// GetIssue retrieves a single issue from a repository
func (server *Server) GetIssue(repo *Repo, issueNum int) (*Issue, error) {

	for {
		server.Limiter.Wait()
		issue, resp, err := server.Client.Issues.Get(server.Context, repo.Owner, repo.Name, issueNum)
		if server.Limiter.HasRateError(resp, err) {
			log.Printf("GetIssue: RateLimitError: %#v\n", err)
			continue
		}

		if err = utils.IsSuccessfulRequest(resp, err); err != nil {
			log.Println("GetIssue: FAILURE:", err)
			return nil, err
		}

		return MakeIssue(server, repo, issue), nil
	}
}

// CreateIssue will create a new issue in the git repo and return the issue object.
func (server *Server) CreateIssue(repo *Repo, issueRequest *github.IssueRequest) (*Issue, error) {

	for {
		server.Limiter.Wait()
		issue, resp, err := server.Client.Issues.Create(server.Context, repo.Owner, repo.Name, issueRequest)
		if server.Limiter.HasRateError(resp, err) {
			log.Printf("CreateIssue: RateLimitError: %#v\n", err)
			continue
		}

		if err = utils.IsSuccessfulRequest(resp, err); err != nil {
			log.Println("CreateIssue: FAILURE:", err)
			return nil, err
		}

		return &Issue{Repo: repo, Server: server, GithubIssue: issue}, err

	}
}

// ListIssuesByRepo will retrieve all of the issues in a repo.  You can limit the number by setting the limit.  A limit of 0
// will retrieve all issues.
func (server *Server) ListIssuesByRepo(repo *Repo, opt *github.IssueListByRepoOptions, limit int) (allIssues []*Issue, err error) {

	for {
		issues, resp, err := server.listByRepo(repo, opt, limit)
		if err = utils.IsSuccessfulRequest(resp, err); err != nil {
			log.Println("GetIssues: FAILURE:", err)
			return nil, err
		}

		allIssues = append(allIssues, issues...)

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
		log.Printf("GetIssues: page=%d\n", opt.Page)

		if limit != 0 && len(allIssues) >= limit {
			log.Printf("GetIssues: reached limit numIssues=%d, limit=%d\n", len(allIssues), limit)
			break
		}
	}

	log.Printf("GetIssues: owner=%s, repo=%s numIssues=%d", repo.Owner, repo.Name, len(allIssues))

	return allIssues, nil

}

func (server *Server) listByRepo(repo *Repo, opt *github.IssueListByRepoOptions, limit int) ([]*Issue, *github.Response, error) {
	var issues []*github.Issue
	var resp *github.Response
	var err error

	for {
		server.Limiter.Wait()

		issues, resp, err = server.Client.Issues.ListByRepo(server.Context, repo.Owner, repo.Name, opt)
		if server.Limiter.HasRateError(resp, err) {
			log.Printf("listByRepo: RateLimitError: %#v\n", err)
			continue
		}

		if err = utils.IsSuccessfulRequest(resp, err); err != nil {
			log.Println("listByRepo: FAILURE:", err)
			return nil, resp, err
		}

		break
	}

	return MakeIssues(server, repo, issues), resp, err
}

// CreateComment creates a comment in an Issue.
func (i *Issue) CreateComment(comment string) (result *github.IssueComment, err error) {

	var resp *github.Response

	ic := &github.IssueComment{Body: github.String(comment)}

	for {
		i.Server.Limiter.Wait()

		result, resp, err = i.Server.Client.Issues.CreateComment(i.Server.Context, i.Repo.Owner, i.Repo.Name, i.Number(), ic)
		if i.Server.Limiter.HasRateError(resp, err) {
			log.Printf("CreateComment: RateLimitError: %#v\n", err)
			continue
		}

		if err = utils.IsSuccessfulRequest(resp, err); err != nil {
			log.Println("CreateComment: FAILURE:", err)
			return nil, err
		}

		break
	}
	return result, nil
}

// AddLabels will add labels to the given issue
func (i *Issue) AddLabels(labels []string) (labelList []*github.Label, err error) {

	var resp *github.Response

	if labels == nil || len(labels) == 0 {
		return nil, errors.New("No labels passed")
	}

	for {
		i.Server.Limiter.Wait()

		labelList, resp, err = i.Server.Client.Issues.AddLabelsToIssue(i.Server.Context, i.Repo.Owner, i.Repo.Name, i.Number(), labels)
		if i.Server.Limiter.HasRateError(resp, err) {
			log.Printf("AddLabels: RateLimitError: %#v\n", err)
			continue
		}

		if err = utils.IsSuccessfulRequest(resp, err); err != nil {
			log.Println("AddLabels: FAILURE:", err)
			return nil, err
		}

		break
	}
	return labelList, nil
}

// RemoveLabel will remove a label from the given issue
func (i *Issue) RemoveLabel(label string) (err error) {

	var resp *github.Response

	if len(label) == 0 {
		return nil
	}

	for {
		i.Server.Limiter.Wait()

		resp, err = i.Server.Client.Issues.RemoveLabelForIssue(i.Server.Context, i.Repo.Owner, i.Repo.Name, i.Number(), label)
		if i.Server.Limiter.HasRateError(resp, err) {
			log.Printf("RemoveLabel: RateLimitError: %#v\n", err)
			continue
		}

		if err = utils.IsSuccessfulRequest(resp, err); err != nil {
			if resp.StatusCode != 404 { // If the label is not found, then not an error to remove it
				log.Println("RemoveLabel: FAILURE:", err)
				return err

			}
		}

		break
	}
	return nil
}
