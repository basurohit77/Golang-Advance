package githubzen

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/google/go-github/github"
)

func TestGetPulls(t *testing.T) {

	// This particular test is designed to run with the live github server.  Therefore, you need to
	// set up the environment to run this test.  If you don't set up the environment variables, it will
	// just end gracefully
	myTestToken := os.Getenv("GITHUBZEN_TESTTOKEN")         // The personal token to use during test
	myTestRepoOwner := os.Getenv("GITHUBZEN_TESTREPOOWNER") // The owner of the repo to run the live test against
	myTestRepoName := os.Getenv("GITHUBZEN_TESTREPONAME")   // The name of the repo to run the live test against

	if myTestToken == "" || myTestRepoOwner == "" || myTestRepoName == "" {
		fmt.Println("Not running TestGetPulls since no live repo values provided")
		return
	}

	server, err := SetupGit("https://github.ibm.com/api/v3/", myTestToken)

	if err != nil {
		t.Log("Failed to get server")
		t.Fatal(err)
	}

	repo, err := server.SetupRepo(myTestRepoOwner, myTestRepoName)
	if err != nil {
		t.Fatal(err)
	}

	opt := &github.PullRequestListOptions{State: "open"}
	pulls, err := server.ListPullsByRepo(repo, opt)

	if err != nil {
		t.Log("could not list pulls by repo")
		t.Fatal(err)
	}

	for _, p := range pulls {
		fmt.Println("=====Pull Reqeust=========")
		fmt.Println(strconv.Itoa(p.Number()) + " Title: " + p.Title())
		fmt.Println("=====Commits=========")
		popt := &github.ListOptions{}
		files, _, err := p.ListFiles(repo, popt)
		if err != nil {
			t.Fatal(err.Error())
		}
		fmt.Println("=====Files=========")
		for _, f := range files {
			fmt.Println("FILE ==============> ", *f.Filename)
			fmt.Println(f.String())
		}
	}
}
