package githubzen

import (
	"time"

	"github.com/google/go-github/github"
)

// Pull is the githubzen abstraction of a pull request. It's meant to improve some
// development aspects and add additional functionality
type Pull struct {
	Server     *Server
	GithubPull *github.PullRequest
	Repo       *Repo
}

// MakePull will wrap the given github pull request with a githubzen pull
func MakePull(server *Server, repo *Repo, pull *github.PullRequest) *Pull {
	return &Pull{Server: server, GithubPull: pull, Repo: repo}
}

// MakePulls will wrap the given github pull slice with a githubzen pulls
func MakePulls(server *Server, repo *Repo, pulls []*github.PullRequest) (result []*Pull) {

	for _, p := range pulls {
		result = append(result, MakePull(server, repo, p))
	}
	return result
}

// ListFiles will get files for a particular PR
func (p *Pull) ListFiles(repo *Repo, opt *github.ListOptions) ([]*github.CommitFile, *github.Response, error) {
	return p.Server.listFilesByPull(repo, p, opt)
}

// ID is used to return the ID as in int64 instead of a pointer
func (p *Pull) ID() int64 {
	return *p.GithubPull.ID
}

// ExistsID will indicate if the ID value actually was set in the Pull
func (p *Pull) ExistsID() bool {
	return p.GithubPull.ID != nil
}

// Number is used to return the int
func (p *Pull) Number() int {
	return *p.GithubPull.Number
}

// ExistsNumber will indicate if the attribute value actually was set in the Pull
func (p *Pull) ExistsNumber() bool {
	return p.GithubPull.Number != nil
}

// State is used to return the string
func (p *Pull) State() string {
	return *p.GithubPull.State
}

// ExistsState will indicate if the attribute value actually was set in the Pull
func (p *Pull) ExistsState() bool {
	return p.GithubPull.State != nil
}

// Title is used to return the string
func (p *Pull) Title() string {
	return *p.GithubPull.Title
}

// ExistsTitle will indicate if the attribute value actually was set in the Pull
func (p *Pull) ExistsTitle() bool {
	return p.GithubPull.Title != nil
}

// Body is used to return the string
func (p *Pull) Body() string {
	return *p.GithubPull.Body
}

// ExistsBody will indicate if the attribute value actually was set in the Pull
func (p *Pull) ExistsBody() bool {
	return p.GithubPull.Body != nil
}

// User is used to return the User
func (p *Pull) User() *github.User {
	return p.GithubPull.User
}

// ExistsUser will indicate if the attribute value actually was set in the Pull
func (p *Pull) ExistsUser() bool {
	return p.GithubPull.User != nil
}

// Labels is used to return the []Label
func (p *Pull) Labels() []*github.Label {
	return p.GithubPull.Labels
}

// ExistsLabels will indicate if the attribute value actually was set in the Pull
func (p *Pull) ExistsLabels() bool {
	return p.GithubPull.Labels != nil
}

// Assignee is used to return the User
func (p *Pull) Assignee() *github.User {
	return p.GithubPull.Assignee
}

// ExistsAssignee will indicate if the attribute value actually was set in the Pull
func (p *Pull) ExistsAssignee() bool {
	return p.GithubPull.Assignee != nil
}

// Comments is used to return the int
func (p *Pull) Comments() int {
	return *p.GithubPull.Comments
}

// ExistsComments will indicate if the attribute value actually was set in the Pull
func (p *Pull) ExistsComments() bool {
	return p.GithubPull.Comments != nil
}

// Commits is used to return the int
func (p *Pull) Commits() int {
	return *p.GithubPull.Commits
}

// ExistsCommits will indicate if the attribute value actually was set in the Pull
func (p *Pull) ExistsCommits() bool {
	return p.GithubPull.Commits != nil
}

// Additions is used to return the int
func (p *Pull) Additions() int {
	return *p.GithubPull.Additions
}

// ExistsAdditions will indicate if the attribute value actually was set in the Pull
func (p *Pull) ExistsAdditions() bool {
	return p.GithubPull.Additions != nil
}

// Deletions is used to return the int
func (p *Pull) Deletions() int {
	return *p.GithubPull.Deletions
}

// ExistsDeletions will indicate if the attribute value actually was set in the Pull
func (p *Pull) ExistsDeletions() bool {
	return p.GithubPull.Deletions != nil
}

// ChangedFiles is used to return the int
func (p *Pull) ChangedFiles() int {
	return *p.GithubPull.ChangedFiles
}

// ExistsChangedFiles will indicate if the attribute value actually was set in the Pull
func (p *Pull) ExistsChangedFiles() bool {
	return p.GithubPull.ChangedFiles != nil
}

// ClosedAt is used to return the time.Time
func (p *Pull) ClosedAt() time.Time {
	return *p.GithubPull.ClosedAt
}

// ExistsClosedAt will indicate if the attribute value actually was set in the Pull
func (p *Pull) ExistsClosedAt() bool {
	return p.GithubPull.ClosedAt != nil
}

// MergedAt is used to return the time.Time
func (p *Pull) MergedAt() time.Time {
	return *p.GithubPull.MergedAt
}

// ExistsMergedAt will indicate if the attribute value actually was set in the Pull
func (p *Pull) ExistsMergedAt() bool {
	return p.GithubPull.MergedAt != nil
}

// Mergeable is used to return the bool
func (p *Pull) Mergeable() bool {
	return *p.GithubPull.Mergeable
}

// ExistsMergeable will indicate if the attribute value actually was set in the Pull
func (p *Pull) ExistsMergeable() bool {
	return p.GithubPull.Mergeable != nil
}

// MergeableState is used to return the string
func (p *Pull) MergeableState() string {
	return *p.GithubPull.MergeableState
}

// ExistsMergeableState will indicate if the attribute value actually was set in the Pull
func (p *Pull) ExistsMergeableState() bool {
	return p.GithubPull.MergeableState != nil
}

// MergeCommitSHA is used to return the string
func (p *Pull) MergeCommitSHA() string {
	return *p.GithubPull.MergeCommitSHA
}

// ExistsMergeCommitSHA will indicate if the attribute value actually was set in the Pull
func (p *Pull) ExistsMergeCommitSHA() bool {
	return p.GithubPull.MergeCommitSHA != nil
}

// Merged is used to return the bool
func (p *Pull) Merged() bool {
	return *p.GithubPull.Merged
}

// ExistsMerged will indicate if the attribute value actually was set in the Pull
func (p *Pull) ExistsMerged() bool {
	return p.GithubPull.Merged != nil
}

// CreatedAt is used to return the time.Time
func (p *Pull) CreatedAt() time.Time {
	return *p.GithubPull.CreatedAt
}

// ExistsCreatedAt will indicate if the attribute value actually was set in the Pull
func (p *Pull) ExistsCreatedAt() bool {
	return p.GithubPull.CreatedAt != nil
}

// UpdatedAt is used to return the time.Time
func (p *Pull) UpdatedAt() time.Time {
	return *p.GithubPull.UpdatedAt
}

// ExistsUpdatedAt will indicate if the attribute value actually was set in the Pull
func (p *Pull) ExistsUpdatedAt() bool {
	return p.GithubPull.UpdatedAt != nil
}

// URL is used to return the string
func (p *Pull) URL() string {
	return *p.GithubPull.URL
}

// ExistsURL will indicate if the attribute value actually was set in the Pull
func (p *Pull) ExistsURL() bool {
	return p.GithubPull.URL != nil
}

// HTMLURL is used to return the string
func (p *Pull) HTMLURL() string {
	return *p.GithubPull.HTMLURL
}

// ExistsHTMLURL will indicate if the attribute value actually was set in the Pull
func (p *Pull) ExistsHTMLURL() bool {
	return p.GithubPull.HTMLURL != nil
}

// IssueURL is used to return the string
func (p *Pull) IssueURL() string {
	return *p.GithubPull.IssueURL
}

// ExistsIssueURL will indicate if the attribute value actually was set in the Pull
func (p *Pull) ExistsIssueURL() bool {
	return p.GithubPull.IssueURL != nil
}

// StatusesURL is used to return the string
func (p *Pull) StatusesURL() string {
	return *p.GithubPull.StatusesURL
}

// ExistsStatusesURL will indicate if the attribute value actually was set in the Pull
func (p *Pull) ExistsStatusesURL() bool {
	return p.GithubPull.StatusesURL != nil
}

// DiffURL is used to return the string
func (p *Pull) DiffURL() string {
	return *p.GithubPull.DiffURL
}

// ExistsDiffURL will indicate if the attribute value actually was set in the Pull
func (p *Pull) ExistsDiffURL() bool {
	return p.GithubPull.DiffURL != nil
}

// PatchURL is used to return the string
func (p *Pull) PatchURL() string {
	return *p.GithubPull.PatchURL
}

// ExistsPatchURL will indicate if the attribute value actually was set in the Pull
func (p *Pull) ExistsPatchURL() bool {
	return p.GithubPull.PatchURL != nil
}

// CommitsURL is used to return the string
func (p *Pull) CommitsURL() string {
	return *p.GithubPull.CommitsURL
}

// ExistsCommitsURL will indicate if the attribute value actually was set in the Pull
func (p *Pull) ExistsCommitsURL() bool {
	return p.GithubPull.CommitsURL != nil
}

// CommentsURL is used to return the string
func (p *Pull) CommentsURL() string {
	return *p.GithubPull.CommentsURL
}

// ExistsCommentsURL will indicate if the attribute value actually was set in the Pull
func (p *Pull) ExistsCommentsURL() bool {
	return p.GithubPull.CommentsURL != nil
}

// ReviewCommentsURL is used to return the string
func (p *Pull) ReviewCommentsURL() string {
	return *p.GithubPull.ReviewCommentsURL
}

// ExistsReviewCommentsURL will indicate if the attribute value actually was set in the Pull
func (p *Pull) ExistsReviewCommentsURL() bool {
	return p.GithubPull.ReviewCommentsURL != nil
}

// ReviewCommentURL is used to return the string
func (p *Pull) ReviewCommentURL() string {
	return *p.GithubPull.ReviewCommentURL
}

// ExistsReviewCommentURL will indicate if the attribute value actually was set in the Pull
func (p *Pull) ExistsReviewCommentURL() bool {
	return p.GithubPull.ReviewCommentURL != nil
}

// Milestone is used to return the Milestone
func (p *Pull) Milestone() *github.Milestone {
	return p.GithubPull.Milestone
}

// ExistsMilestone will indicate if the attribute value actually was set in the Pull
func (p *Pull) ExistsMilestone() bool {
	return p.GithubPull.Milestone != nil
}

// Assignees is used to return the ]*User
func (p *Pull) Assignees() []*github.User {
	return p.GithubPull.Assignees
}

// ExistsAssignees will indicate if the attribute value actually was set in the Pull
func (p *Pull) ExistsAssignees() bool {
	return p.GithubPull.Assignees != nil
}

// NodeID is used to return the string
func (p *Pull) NodeID() string {
	return *p.GithubPull.NodeID
}

// ExistsNodeID will indicate if the attribute value actually was set in the Pull
func (p *Pull) ExistsNodeID() bool {
	return p.GithubPull.NodeID != nil
}

// ActiveLockReason is used to return the string
func (p *Pull) ActiveLockReason() string {
	return *p.GithubPull.ActiveLockReason
}

// ExistsActiveLockReason will indicate if the attribute value actually was set in the Pull
func (p *Pull) ExistsActiveLockReason() bool {
	return p.GithubPull.ActiveLockReason != nil
}

// MaintainerCanModify is used to return the string
func (p *Pull) MaintainerCanModify() bool {
	return *p.GithubPull.MaintainerCanModify
}

// ExistsMaintainerCanModify will indicate if the attribute value actually was set in the Pull
func (p *Pull) ExistsMaintainerCanModify() bool {
	return p.GithubPull.MaintainerCanModify != nil
}

// AuthorAssociation is used to return the string
func (p *Pull) AuthorAssociation() string {
	return *p.GithubPull.AuthorAssociation
}

// ExistsAuthorAssociation will indicate if the attribute value actually was set in the Pull
func (p *Pull) ExistsAuthorAssociation() bool {
	return p.GithubPull.AuthorAssociation != nil
}

// RequestedReviewers is used to return the string
func (p *Pull) RequestedReviewers() []*github.User {
	return p.GithubPull.RequestedReviewers
}

// ExistsRequestedReviewers will indicate if the attribute value actually was set in the Pull
func (p *Pull) ExistsRequestedReviewers() bool {
	return p.GithubPull.RequestedReviewers != nil
}

// Head is used to return the string
func (p *Pull) Head() github.PullRequestBranch {
	return *p.GithubPull.Head
}

// ExistsHead will indicate if the attribute value actually was set in the Pull
func (p *Pull) ExistsHead() bool {
	return p.GithubPull.Head != nil
}

// Base is used to return the string
func (p *Pull) Base() github.PullRequestBranch {
	return *p.GithubPull.Base
}

// ExistsBase will indicate if the attribute value actually was set in the Pull
func (p *Pull) ExistsBase() bool {
	return p.GithubPull.Base != nil
}
