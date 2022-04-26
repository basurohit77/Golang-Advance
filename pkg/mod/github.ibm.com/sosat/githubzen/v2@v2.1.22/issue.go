package githubzen

import (
	"time"

	"github.com/google/go-github/github"
)

// Issue is the githubzen abstraction of an issue. It's meant to improve some
// development aspects and add additional functionality
type Issue struct {
	Server      *Server
	GithubIssue *github.Issue
	Repo        *Repo
}

// MakeIssue will wrap the given github issue with a githubzen issue
func MakeIssue(server *Server, repo *Repo, issue *github.Issue) *Issue {
	return &Issue{Server: server, GithubIssue: issue, Repo: repo}
}

// MakeIssues will wrap the given github issue slice with a githubzen issues
func MakeIssues(server *Server, repo *Repo, issues []*github.Issue) (result []*Issue) {

	for _, i := range issues {
		result = append(result, MakeIssue(server, repo, i))
	}
	return result
}

// ID is used to return the ID as in int64 instead of a pointer
func (i *Issue) ID() int64 {
	return *i.GithubIssue.ID
}

// ExistsID will indicate if the ID value actually was set in the Issue
func (i *Issue) ExistsID() bool {
	return i.GithubIssue.ID != nil
}

// Number is used to return the int
func (i *Issue) Number() int {
	return *i.GithubIssue.Number
}

// ExistsNumber will indicate if the attribute value actually was set in the Issue
func (i *Issue) ExistsNumber() bool {
	return i.GithubIssue.Number != nil
}

// State is used to return the string
func (i *Issue) State() string {
	return *i.GithubIssue.State
}

// ExistsState will indicate if the attribute value actually was set in the Issue
func (i *Issue) ExistsState() bool {
	return i.GithubIssue.State != nil
}

// Locked is used to return the bool
func (i *Issue) Locked() bool {
	return *i.GithubIssue.Locked
}

// ExistsLocked will indicate if the attribute value actually was set in the Issue
func (i *Issue) ExistsLocked() bool {
	return i.GithubIssue.Locked != nil
}

// Title is used to return the string
func (i *Issue) Title() string {
	return *i.GithubIssue.Title
}

// ExistsTitle will indicate if the attribute value actually was set in the Issue
func (i *Issue) ExistsTitle() bool {
	return i.GithubIssue.Title != nil
}

// Body is used to return the string
func (i *Issue) Body() string {
	return *i.GithubIssue.Body
}

// ExistsBody will indicate if the attribute value actually was set in the Issue
func (i *Issue) ExistsBody() bool {
	return i.GithubIssue.Body != nil
}

// User is used to return the User
func (i *Issue) User() *github.User {
	return i.GithubIssue.User
}

// ExistsUser will indicate if the attribute value actually was set in the Issue
func (i *Issue) ExistsUser() bool {
	return i.GithubIssue.User != nil
}

// Labels is used to return the []Label
func (i *Issue) Labels() []github.Label {
	return i.GithubIssue.Labels
}

// ExistsLabels will indicate if the attribute value actually was set in the Issue
func (i *Issue) ExistsLabels() bool {
	return i.GithubIssue.Labels != nil
}

// Assignee is used to return the User
func (i *Issue) Assignee() *github.User {
	return i.GithubIssue.Assignee
}

// ExistsAssignee will indicate if the attribute value actually was set in the Issue
func (i *Issue) ExistsAssignee() bool {
	return i.GithubIssue.Assignee != nil
}

// Comments is used to return the int
func (i *Issue) Comments() int {
	return *i.GithubIssue.Comments
}

// ExistsComments will indicate if the attribute value actually was set in the Issue
func (i *Issue) ExistsComments() bool {
	return i.GithubIssue.Comments != nil
}

// ClosedAt is used to return the time.Time
func (i *Issue) ClosedAt() time.Time {
	return *i.GithubIssue.ClosedAt
}

// ExistsClosedAt will indicate if the attribute value actually was set in the Issue
func (i *Issue) ExistsClosedAt() bool {
	return i.GithubIssue.ClosedAt != nil
}

// CreatedAt is used to return the time.Time
func (i *Issue) CreatedAt() time.Time {
	return *i.GithubIssue.CreatedAt
}

// ExistsCreatedAt will indicate if the attribute value actually was set in the Issue
func (i *Issue) ExistsCreatedAt() bool {
	return i.GithubIssue.CreatedAt != nil
}

// UpdatedAt is used to return the time.Time
func (i *Issue) UpdatedAt() time.Time {
	return *i.GithubIssue.UpdatedAt
}

// ExistsUpdatedAt will indicate if the attribute value actually was set in the Issue
func (i *Issue) ExistsUpdatedAt() bool {
	return i.GithubIssue.UpdatedAt != nil
}

// ClosedBy is used to return the User
func (i *Issue) ClosedBy() *github.User {
	return i.GithubIssue.ClosedBy
}

// ExistsClosedBy will indicate if the attribute value actually was set in the Issue
func (i *Issue) ExistsClosedBy() bool {
	return i.GithubIssue.ClosedBy != nil
}

// URL is used to return the string
func (i *Issue) URL() string {
	return *i.GithubIssue.URL
}

// ExistsURL will indicate if the attribute value actually was set in the Issue
func (i *Issue) ExistsURL() bool {
	return i.GithubIssue.URL != nil
}

// HTMLURL is used to return the string
func (i *Issue) HTMLURL() string {
	return *i.GithubIssue.HTMLURL
}

// ExistsHTMLURL will indicate if the attribute value actually was set in the Issue
func (i *Issue) ExistsHTMLURL() bool {
	return i.GithubIssue.HTMLURL != nil
}

// CommentsURL is used to return the string
func (i *Issue) CommentsURL() string {
	return *i.GithubIssue.CommentsURL
}

// ExistsCommentsURL will indicate if the attribute value actually was set in the Issue
func (i *Issue) ExistsCommentsURL() bool {
	return i.GithubIssue.CommentsURL != nil
}

// EventsURL is used to return the string
func (i *Issue) EventsURL() string {
	return *i.GithubIssue.EventsURL
}

// ExistsEventsURL will indicate if the attribute value actually was set in the Issue
func (i *Issue) ExistsEventsURL() bool {
	return i.GithubIssue.EventsURL != nil
}

// LabelsURL is used to return the string
func (i *Issue) LabelsURL() string {
	return *i.GithubIssue.LabelsURL
}

// ExistsLabelsURL will indicate if the attribute value actually was set in the Issue
func (i *Issue) ExistsLabelsURL() bool {
	return i.GithubIssue.LabelsURL != nil
}

// RepositoryURL is used to return the string
func (i *Issue) RepositoryURL() string {
	return *i.GithubIssue.RepositoryURL
}

// ExistsRepositoryURL will indicate if the attribute value actually was set in the Issue
func (i *Issue) ExistsRepositoryURL() bool {
	return i.GithubIssue.RepositoryURL != nil
}

// Milestone is used to return the Milestone
func (i *Issue) Milestone() *github.Milestone {
	return i.GithubIssue.Milestone
}

// ExistsMilestone will indicate if the attribute value actually was set in the Issue
func (i *Issue) ExistsMilestone() bool {
	return i.GithubIssue.Milestone != nil
}

// PullRequestLinks is used to return the PullRequestLinks
func (i *Issue) PullRequestLinks() *github.PullRequestLinks {
	return i.GithubIssue.PullRequestLinks
}

// ExistsPullRequestLinks will indicate if the attribute value actually was set in the Issue
func (i *Issue) ExistsPullRequestLinks() bool {
	return i.GithubIssue.PullRequestLinks != nil
}

// Repository is used to return the Repository
func (i *Issue) Repository() *github.Repository {
	return i.GithubIssue.Repository
}

// ExistsRepository will indicate if the attribute value actually was set in the Issue
func (i *Issue) ExistsRepository() bool {
	return i.GithubIssue.Repository != nil
}

// Reactions is used to return the Reactions
func (i *Issue) Reactions() *github.Reactions {
	return i.GithubIssue.Reactions
}

// ExistsReactions will indicate if the attribute value actually was set in the Issue
func (i *Issue) ExistsReactions() bool {
	return i.GithubIssue.Reactions != nil
}

// Assignees is used to return the ]*User
func (i *Issue) Assignees() []*github.User {
	return i.GithubIssue.Assignees
}

// ExistsAssignees will indicate if the attribute value actually was set in the Issue
func (i *Issue) ExistsAssignees() bool {
	return i.GithubIssue.Assignees != nil
}

// NodeID is used to return the string
func (i *Issue) NodeID() string {
	return *i.GithubIssue.NodeID
}

// ExistsNodeID will indicate if the attribute value actually was set in the Issue
func (i *Issue) ExistsNodeID() bool {
	return i.GithubIssue.NodeID != nil
}

// TextMatches is used to return the ]TextMatch
func (i *Issue) TextMatches() []github.TextMatch {
	return i.GithubIssue.TextMatches
}

// ExistsTextMatches will indicate if the attribute value actually was set in the Issue
func (i *Issue) ExistsTextMatches() bool {
	return i.GithubIssue.TextMatches != nil
}

// ActiveLockReason is used to return the string
func (i *Issue) ActiveLockReason() string {
	return *i.GithubIssue.ActiveLockReason
}

// ExistsActiveLockReason will indicate if the attribute value actually was set in the Issue
func (i *Issue) ExistsActiveLockReason() bool {
	return i.GithubIssue.ActiveLockReason != nil
}
