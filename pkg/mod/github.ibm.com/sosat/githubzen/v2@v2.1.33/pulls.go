package githubzen

import (
	"log"

	"github.com/google/go-github/github"
	"github.ibm.com/sosat/githubzen/v2/utils"
)

// ListPullsByRepo will retrieve all of the pulls in a repo.
func (server *Server) ListPullsByRepo(repo *Repo, opt *github.PullRequestListOptions) (allPulls []*Pull, err error) {

	pulls, resp, err := server.listPRsByRepo(repo, opt)
	if err = utils.IsSuccessfulRequest(resp, err); err != nil {
		log.Println("ListPullsByRepo: FAILURE:", err)
		return nil, err
	}

	log.Printf("ListPullsByRepo: owner=%s, repo=%s numPulls=%d", repo.Owner, repo.Name, len(pulls))

	return pulls, nil

}

// GetPull retrieves a single pull from a repository
func (server *Server) GetPull(repo *Repo, pullNum int) (*Pull, error) {

	for {
		server.Limiter.Wait()
		pull, resp, err := server.Client.PullRequests.Get(server.Context, repo.Owner, repo.Name, pullNum)
		if server.Limiter.HasRateError(resp, err) {
			log.Printf("GetPull: RateLimitError: %#v\n", err)
			continue
		}

		if err = utils.IsSuccessfulRequest(resp, err); err != nil {
			log.Println("GetPull: FAILURE:", err)
			return nil, err
		}

		return MakePull(server, repo, pull), nil
	}
}

func (server *Server) listPRsByRepo(repo *Repo, opt *github.PullRequestListOptions) ([]*Pull, *github.Response, error) {
	var pulls []*github.PullRequest
	var resp *github.Response
	var err error

	for {
		server.Limiter.Wait()

		pulls, resp, err = server.Client.PullRequests.List(server.Context, repo.Owner, repo.Name, opt)
		if server.Limiter.HasRateError(resp, err) {
			log.Printf("listPullsByRepo: RateLimitError: %#v\n", err)
			continue
		}

		if err = utils.IsSuccessfulRequest(resp, err); err != nil {
			log.Println("listPullsByRepo: FAILURE:", err)
			return nil, resp, err
		}

		break
	}

	return MakePulls(server, repo, pulls), resp, err
}

func (server *Server) listFilesByPull(repo *Repo, pull *Pull, opt *github.ListOptions) ([]*github.CommitFile, *github.Response, error) {

	var files []*github.CommitFile
	var resp *github.Response
	var err error

	for {
		server.Limiter.Wait()

		files, resp, err = server.Client.PullRequests.ListFiles(server.Context, repo.Owner, repo.Name, pull.Number(), opt)
		if server.Limiter.HasRateError(resp, err) {
			log.Printf("listFilesbyPull: RateLimitError: %#v\n", err)
			continue
		}

		if err = utils.IsSuccessfulRequest(resp, err); err != nil {
			log.Println("listFilesbyPull: FAILURE:", err)
			return nil, resp, err
		}

		break
	}

	return files, resp, err
}
