package githubzen

import (
	"log"

	"github.com/google/go-github/github"
	"github.ibm.com/sosat/githubzen/v2/utils"
)

// Repo is a basic representation of a repository
type Repo struct {
	Owner  string
	Name   string
	GHRepo *github.Repository
}

// SetupRepo creates a representation of a repository. This is a respository
// that already exists within the git server
func (server *Server) SetupRepo(owner, name string) (repo *Repo, err error) {

	var resp *github.Response
	var ghrepo *github.Repository

	for {
		server.Limiter.Wait()
		ghrepo, resp, err = server.Client.Repositories.Get(server.Context, owner, name)
		if server.Limiter.HasRateError(resp, err) {
			continue
		}
		if err1 := utils.IsSuccessfulRequest(resp, err); err1 != nil {
			log.Println("SetupRepo: ERROR getting repository", name, "error=", err1)
			return nil, err
		}
		break
	}

	repo = &Repo{Owner: owner, Name: name, GHRepo: ghrepo}

	server.Repos[repoKey(owner, name)] = repo

	return repo, nil
}

// GetRepo will return a repo that was previously setup with the SetupRepo call
// nil is returned if the repo has not already been setup
func (server *Server) GetRepo(owner, name string) *Repo {
	return server.Repos[repoKey(owner, name)]
}

func repoKey(owner, name string) string {
	return owner + "/" + name
}
