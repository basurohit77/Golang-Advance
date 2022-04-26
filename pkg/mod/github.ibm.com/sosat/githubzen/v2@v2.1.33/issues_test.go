package githubzen

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/google/go-github/github"
)

func TestGetIssue(t *testing.T) {

	// This particular test is designed to run with the live github server.  Therefore, you need to
	// set up the environment to run this test.  If you don't set up the environment variables, it will
	// just end gracefully
	myTestToken := os.Getenv("GITHUBZEN_TESTTOKEN")         // The personal token to use during test
	myTestRepoOwner := os.Getenv("GITHUBZEN_TESTREPOOWNER") // The owner of the repo to run the live test against
	myTestRepoName := os.Getenv("GITHUBZEN_TESTREPONAME")   // The name of the repo to run the live test against

	if myTestToken == "" || myTestRepoOwner == "" || myTestRepoName == "" {
		fmt.Println("Not running TestGetIssue since no live repo values provided")
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

	issue, err := server.GetIssue(repo, 2)

	if err != nil {
		t.Log("Failed to get issue")
		t.Fatal(err)
	}

	log.Println(issue.Body())

	// Second request
	fmt.Println("Second Request")
	issue, err = server.GetIssue(repo, 2)

	if err != nil {
		t.Log("Failed to get issue")
		t.Fatal(err)
	}

	log.Println("BODY START")
	log.Println(issue.Body())
	log.Println("BODY END")

	log.Println("=================== Test setting and issue comment =================== ")

	_, err = issue.CreateComment("This is a comment made programmatically!!")
	if err != nil {
		t.Log("Could not create comment")
		t.Fatal(err)
	}

	_, err = issue.AddLabels([]string{"ERROR"})
	if err != nil {
		t.Log("Could not add labels")
		t.Fatal(err)
	}

	err = issue.RemoveLabel("EXPIRED")
	if err != nil {
		t.Log("Could not remove label")
		t.Fatal(err)
	}

	log.Println("=================== Test getting all the issues =================== ")

	opt := &github.IssueListByRepoOptions{Labels: []string{"publish"}}
	issues, err := server.ListIssuesByRepo(repo, opt, 0)

	if err != nil {
		t.Log("could not list by repo")
		t.Fatal(err)
	}

	for _, i := range issues {
		fmt.Println("Title: " + i.Title())
	}

}

func TestCreateIssue(t *testing.T) {

	// This particular test is designed to run with the live github server.  Therefore, you need to
	// set up the environment to run this test.  If you don't set up the environment variables, it will
	// just end gracefully
	myTestToken := os.Getenv("GITHUBZEN_TESTTOKEN")         // The personal token to use during test
	myTestRepoOwner := os.Getenv("GITHUBZEN_TESTREPOOWNER") // The owner of the repo to run the live test against
	myTestRepoName := os.Getenv("GITHUBZEN_TESTREPONAME")   // The name of the repo to run the live test against

	if myTestToken == "" || myTestRepoOwner == "" || myTestRepoName == "" {
		fmt.Println("Not running TestGetIssue since no live repo values provided")
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

	issueRequest := &github.IssueRequest{
		Title:  github.String("Sample created issue " + time.Now().String()),
		Body:   github.String("This is the body of the issue created by a test program"),
		Labels: &[]string{"announcement"},
	}

	_, err = server.CreateIssue(repo, issueRequest)

	if err != nil {
		t.Fatal(err)
	}
}
