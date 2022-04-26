package gitnotice

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/go-github/github"
)

var SampleText1 = `**==== START====**

2019-10-31T13:00:00Z

**==== SERVICE ====** (https://osscatviewer.stage1.mybluemix.net)

cloudantnosqldb

**==== IMPACTED LOCATIONS ====** (https://osscatviewer.stage1.mybluemix.net/environments)

crn:v1:bluemix:public::us-south::::
crn:v1:bluemix:public::us-east::::
crn:v1:bluemix:public::eu-gb::::
crn:v1:bluemix:public::eu-de::::
crn:v1:bluemix:public::au-syd::::
crn:v1:bluemix:public::jp-tok::::

**==== AUDIENCE ====**
private

**==== TITLE ====**

Type the title of the notification here

**==== DESCRIPTION ====**

This is a description of what happened.  Can have <b>formatting</b> in HTML as needed.
<ul>
<li>hello
<li>world
</ul>
end`

var Audience = `**==== AUDIENCE ====**
private
`
var AudienceContent = `
private
`

var Start = `**==== START====**

2019-10-31T13:00:00Z
`
var StartContent = `
2019-10-31T13:00:00Z
`
var Description = `**==== DESCRIPTION ====**

This is a description of what happened.  Can have <b>formatting</b> in HTML as needed.
<ul>
<li>hello
<li>world
</ul>
end`
var DescriptionContent = `
This is a description of what happened.  Can have <b>formatting</b> in HTML as needed.
<ul>
<li>hello
<li>world
</ul>
end`

var Locations = `**==== IMPACTED LOCATIONS ====** (https://osscatviewer.stage1.mybluemix.net/environments)

crn:v1:bluemix:public::us-south::::
crn:v1:bluemix:public::us-east::::
crn:v1:bluemix:public::eu-gb::::
crn:v1:bluemix:public::eu-de::::
crn:v1:bluemix:public::au-syd::::
crn:v1:bluemix:public::jp-tok::::
`
var LocationsContent = `
crn:v1:bluemix:public::us-south::::
crn:v1:bluemix:public::us-east::::
crn:v1:bluemix:public::eu-gb::::
crn:v1:bluemix:public::eu-de::::
crn:v1:bluemix:public::au-syd::::
crn:v1:bluemix:public::jp-tok::::
`
var Title = `**==== TITLE ====**

Type the title of the notification here
`
var TitleContent = `
Type the title of the notification here
`

func TestMarker(t *testing.T) {

	txt := getMarkerContent(StartMarker, SampleText1)
	if txt != Start {
		t.Log("SHOULD:" + strings.ReplaceAll(Start, "\n", "\\n"))
		t.Log("BUTWAS:" + strings.ReplaceAll(txt, "\n", "\\n"))
		t.Fatal("Did not match " + StartMarker)
	}

	txt = getMarkerContent(DescriptionMarker, SampleText1)
	if txt != Description {
		t.Log("SHOULD:" + strings.ReplaceAll(Description, "\n", "\\n"))
		t.Log("BUTWAS:" + strings.ReplaceAll(txt, "\n", "\\n"))
		t.Fatal("Did not match " + DescriptionMarker)
	}

	txt = getMarkerContent(LocationMarker, SampleText1)
	if txt != Locations {
		t.Log("SHOULD:" + strings.ReplaceAll(Locations, "\n", "\\n"))
		t.Log("BUTWAS:" + strings.ReplaceAll(txt, "\n", "\\n"))
		t.Fatal("Did not match " + LocationMarker)
	}

	txt = getMarkerContent(AudienceMarker, SampleText1)
	if txt != Audience {
		t.Log("SHOULD:" + strings.ReplaceAll(Audience, "\n", "\\n"))
		t.Log("BUTWAS:" + strings.ReplaceAll(txt, "\n", "\\n"))
		t.Fatal("Did not match " + AudienceMarker)
	}
}

func TestParser(t *testing.T) {
	pmap := ParseDescription(SampleText1)

	txt := pmap[StartMarker].Content
	if txt != Start {
		t.Log("SHOULD:" + strings.ReplaceAll(Start, "\n", "\\n"))
		t.Log("BUTWAS:" + strings.ReplaceAll(txt, "\n", "\\n"))
		t.Fatal("Did not match " + StartMarker)
	}
	txt = pmap[StartMarker].GetContent()
	if txt != StartContent {
		t.Log("SHOULD:" + strings.ReplaceAll(StartContent, "\n", "\\n"))
		t.Log("BUTWAS:" + strings.ReplaceAll(txt, "\n", "\\n"))
		t.Fatal("Did not match content for " + StartMarker)
	}

	txt = pmap[DescriptionMarker].Content
	if txt != Description {
		t.Log("SHOULD:" + strings.ReplaceAll(Description, "\n", "\\n"))
		t.Log("BUTWAS:" + strings.ReplaceAll(txt, "\n", "\\n"))
		t.Fatal("Did not match " + DescriptionMarker)
	}
	txt = pmap[DescriptionMarker].GetContent()
	if txt != DescriptionContent {
		t.Log("SHOULD:" + strings.ReplaceAll(DescriptionContent, "\n", "\\n"))
		t.Log("BUTWAS:" + strings.ReplaceAll(txt, "\n", "\\n"))
		t.Fatal("Did not match content for " + DescriptionMarker)
	}

	txt = pmap[LocationMarker].Content
	if txt != Locations {
		t.Log("SHOULD:" + strings.ReplaceAll(Locations, "\n", "\\n"))
		t.Log("BUTWAS:" + strings.ReplaceAll(txt, "\n", "\\n"))
		t.Fatal("Did not match " + LocationMarker)
	}
	txt = pmap[LocationMarker].GetContent()
	if txt != LocationsContent {
		t.Log("SHOULD:" + strings.ReplaceAll(LocationsContent, "\n", "\\n"))
		t.Log("BUTWAS:" + strings.ReplaceAll(txt, "\n", "\\n"))
		t.Fatal("Did not match content for " + LocationMarker)
	}

	txt = pmap[TitleMarker].Content
	if txt != Title {
		t.Log("SHOULD:" + strings.ReplaceAll(Title, "\n", "\\n"))
		t.Log("BUTWAS:" + strings.ReplaceAll(txt, "\n", "\\n"))
		t.Fatal("Did not match " + TitleMarker)
	}
	txt = pmap[TitleMarker].GetContent()
	if txt != TitleContent {
		t.Log("SHOULD:" + strings.ReplaceAll(TitleContent, "\n", "\\n"))
		t.Log("BUTWAS:" + strings.ReplaceAll(txt, "\n", "\\n"))
		t.Fatal("Did not match " + TitleMarker)
	}
}

func TestCreateNotice(t *testing.T) {

	notice, err := BuildNotice(SampleText1)

	if err != nil {
		t.Fatal(err.String())
	}

	if notice.Title != TitleContent {
		t.Log("SHOULD:" + strings.ReplaceAll(TitleContent, "\n", "\\n"))
		t.Log("BUTWAS:" + strings.ReplaceAll(notice.Title, "\n", "\\n"))
		t.Fatal("Did not match " + TitleMarker)
	}

	if notice.Description != DescriptionContent {
		t.Log("SHOULD:" + strings.ReplaceAll(DescriptionContent, "\n", "\\n"))
		t.Log("BUTWAS:" + strings.ReplaceAll(notice.Description, "\n", "\\n"))
		t.Fatal("Did not match content for " + DescriptionMarker)
	}

	if len(notice.Services) != 1 {
		msg := fmt.Sprintf("Wrong number of services %d", len(notice.Services))
		t.Fatal(msg)
	}

	if notice.Services[0] != "cloudantnosqldb" {
		t.Error("Wrong service value:", notice.Services[0])
	}

	if len(notice.Locations) != 6 {
		t.Error("Wrong number of environments")
	}

	startTime := time.Date(2019, 10, 31, 13, 0, 0, 0, time.UTC)

	if notice.Start != startTime {
		t.Error("Wrong start time")
	}
}

var SampleTextBadDate = `**==== START====**

2019-10-31T13:00:00

**==== SERVICE ====** (https://osscatviewer.stage1.mybluemix.net)

cloudantnosqldb

**==== IMPACTED LOCATIONS ====** (https://osscatviewer.stage1.mybluemix.net/environments)

crn:v1:bluemix:public::us-south::::
crn:v1:bluemix:public::us-east::::
crn:v1:bluemix:public::eu-gb::::
crn:v1:bluemix:public::eu-de::::
crn:v1:bluemix:public::au-syd::::
crn:v1:bluemix:public::jp-tok::::

**==== TITLE ====**

Type the title of the notification here

**==== DESCRIPTION ====**

This is a description of what happened.  Can have <b>formatting</b> in HTML as needed.
<ul>
<li>hello
<li>world
</ul>
end`

func TestBadDate(t *testing.T) {

	notice, err := BuildNotice(SampleTextBadDate)

	if err == nil {
		t.Fatal("Expected error for bad date format, but did not get an error " + notice.Start.String())
	}

}

func TestGitParse(t *testing.T) {
	eventIfc, err := github.ParseWebHook("issues", []byte(receivedDataSample))
	if err != nil {
		t.Fatal(err)
	}

	var event *github.IssuesEvent

	switch eventCvt := eventIfc.(type) {
	case *github.IssuesEvent:
		event = eventCvt
	default:
		t.Fatalf("ERROR unrecognized event type. %T", event)
		return
	}

	fmt.Println(event.GetRepo().GetOwner().GetLogin())
	fmt.Println(event.GetRepo().GetName())
	fmt.Println(event.Issue.GetUser().GetLogin())
}

var receivedDataSample = `{
  "action": "edited",
  "issue": {
    "url": "https://github.ibm.com/api/v3/repos/cloud-sre/pnp-notification-messages-test/issues/5",
    "repository_url": "https://github.ibm.com/api/v3/repos/cloud-sre/pnp-notification-messages-test",
    "labels_url": "https://github.ibm.com/api/v3/repos/cloud-sre/pnp-notification-messages-test/issues/5/labels{/name}",
    "comments_url": "https://github.ibm.com/api/v3/repos/cloud-sre/pnp-notification-messages-test/issues/5/comments",
    "events_url": "https://github.ibm.com/api/v3/repos/cloud-sre/pnp-notification-messages-test/issues/5/events",
    "html_url": "https://github.ibm.com/cloud-sre/pnp-notification-messages-test/issues/5",
    "id": 9534132,
    "node_id": "MDU6SXNzdWU5NTM0MTMy",
    "number": 5,
    "title": "An issue to announce a service price drop",
    "user": {
      "login": "kparzygn",
      "id": 5122,
      "node_id": "MDQ6VXNlcjUxMjI=",
      "avatar_url": "https://avatars.github.ibm.com/u/5122?",
      "gravatar_id": "",
      "url": "https://github.ibm.com/api/v3/users/kparzygn",
      "html_url": "https://github.ibm.com/kparzygn",
      "followers_url": "https://github.ibm.com/api/v3/users/kparzygn/followers",
      "following_url": "https://github.ibm.com/api/v3/users/kparzygn/following{/other_user}",
      "gists_url": "https://github.ibm.com/api/v3/users/kparzygn/gists{/gist_id}",
      "starred_url": "https://github.ibm.com/api/v3/users/kparzygn/starred{/owner}{/repo}",
      "subscriptions_url": "https://github.ibm.com/api/v3/users/kparzygn/subscriptions",
      "organizations_url": "https://github.ibm.com/api/v3/users/kparzygn/orgs",
      "repos_url": "https://github.ibm.com/api/v3/users/kparzygn/repos",
      "events_url": "https://github.ibm.com/api/v3/users/kparzygn/events{/privacy}",
      "received_events_url": "https://github.ibm.com/api/v3/users/kparzygn/received_events",
      "type": "User",
      "site_admin": false
    },
    "labels": [
      {
        "id": 5209050,
        "node_id": "MDU6TGFiZWw1MjA5MDUw",
        "url": "https://github.ibm.com/api/v3/repos/cloud-sre/pnp-notification-messages-test/labels/announcement",
        "name": "announcement",
        "color": "000000",
        "default": false
      },
      {
        "id": 5213389,
        "node_id": "MDU6TGFiZWw1MjEzMzg5",
        "url": "https://github.ibm.com/api/v3/repos/cloud-sre/pnp-notification-messages-test/labels/publish",
        "name": "publish",
        "color": "008000",
        "default": false
      }
    ],
    "state": "open",
    "locked": false,
    "assignee": null,
    "assignees": [],
    "milestone": null,
    "comments": 0,
    "created_at": "2019-07-31T18:43:20Z",
    "updated_at": "2019-08-02T19:07:19Z",
    "closed_at": null,
    "author_association": "CONTRIBUTOR",
    "body": "**==== START ====**\r\n\r\n2019-10-31T13:00:00\r\n\r\n**==== SERVICE ====** (https://osscatviewer.stage1.mybluemix.net)\r\n\r\ncloudantnosqldb\r\n\r\n**==== IMPACTED LOCATIONS ====** (https://osscatviewer.stage1.mybluemix.net/environments)\r\n\r\ncrn:v1:bluemix:public::us-south::::\r\ncrn:v1:bluemix:public::us-east::::\r\n\r\n**==== TITLE ====**\r\nType the title of the notification here. Test announcement notification. Please ignore.\r\n\r\n**==== DESCRIPTION ====**\r\nThis is a description of what happened.  Can have <b>formatting</b> in HTML as needed.\r\n<ul>\r\n<li>hello\r\n<li>world\r\n</ul>\r\n"
  },
  "changes": {
    "body": {
      "from": "**==== START ====**\r\n\r\n2019-10-31T13:00:00\r\n\r\n**==== SERVICE ====** (https://osscatviewer.stage1.mybluemix.net)\r\n\r\ncloudantnosqldb\r\n\r\n**==== IMPACTED LOCATIONS ====** (https://osscatviewer.stage1.mybluemix.net/environments)\r\n\r\ncrn:v1:bluemix:public::us-south::::\r\ncrn:v1:bluemix:public::us-east::::\r\n\r\n**==== TITLE ====**\r\nType the title of the notification here\r\n\r\n**==== DESCRIPTION ====**\r\nThis is a description of what happened.  Can have <b>formatting</b> in HTML as needed.\r\n<ul>\r\n<li>hello\r\n<li>world\r\n</ul>\r\n"
    }
  },
  "repository": {
    "id": 615833,
    "node_id": "MDEwOlJlcG9zaXRvcnk2MTU4MzM=",
    "name": "pnp-notification-messages-test",
    "full_name": "cloud-sre/pnp-notification-messages-test",
    "private": true,
    "owner": {
      "login": "cloud-sre",
      "id": 10240,
      "node_id": "MDEyOk9yZ2FuaXphdGlvbjEwMjQw",
      "avatar_url": "https://avatars.github.ibm.com/u/10240?",
      "gravatar_id": "",
      "url": "https://github.ibm.com/api/v3/users/cloud-sre",
      "html_url": "https://github.ibm.com/cloud-sre",
      "followers_url": "https://github.ibm.com/api/v3/users/cloud-sre/followers",
      "following_url": "https://github.ibm.com/api/v3/users/cloud-sre/following{/other_user}",
      "gists_url": "https://github.ibm.com/api/v3/users/cloud-sre/gists{/gist_id}",
      "starred_url": "https://github.ibm.com/api/v3/users/cloud-sre/starred{/owner}{/repo}",
      "subscriptions_url": "https://github.ibm.com/api/v3/users/cloud-sre/subscriptions",
      "organizations_url": "https://github.ibm.com/api/v3/users/cloud-sre/orgs",
      "repos_url": "https://github.ibm.com/api/v3/users/cloud-sre/repos",
      "events_url": "https://github.ibm.com/api/v3/users/cloud-sre/events{/privacy}",
      "received_events_url": "https://github.ibm.com/api/v3/users/cloud-sre/received_events",
      "type": "Organization",
      "site_admin": false
    },
    "html_url": "https://github.ibm.com/cloud-sre/pnp-notification-messages-test",
    "description": "Used to test the ability to pull notifications from github",
    "fork": false,
    "url": "https://github.ibm.com/api/v3/repos/cloud-sre/pnp-notification-messages-test",
    "forks_url": "https://github.ibm.com/api/v3/repos/cloud-sre/pnp-notification-messages-test/forks",
    "keys_url": "https://github.ibm.com/api/v3/repos/cloud-sre/pnp-notification-messages-test/keys{/key_id}",
    "collaborators_url": "https://github.ibm.com/api/v3/repos/cloud-sre/pnp-notification-messages-test/collaborators{/collaborator}",
    "teams_url": "https://github.ibm.com/api/v3/repos/cloud-sre/pnp-notification-messages-test/teams",
    "hooks_url": "https://github.ibm.com/api/v3/repos/cloud-sre/pnp-notification-messages-test/hooks",
    "issue_events_url": "https://github.ibm.com/api/v3/repos/cloud-sre/pnp-notification-messages-test/issues/events{/number}",
    "events_url": "https://github.ibm.com/api/v3/repos/cloud-sre/pnp-notification-messages-test/events",
    "assignees_url": "https://github.ibm.com/api/v3/repos/cloud-sre/pnp-notification-messages-test/assignees{/user}",
    "branches_url": "https://github.ibm.com/api/v3/repos/cloud-sre/pnp-notification-messages-test/branches{/branch}",
    "tags_url": "https://github.ibm.com/api/v3/repos/cloud-sre/pnp-notification-messages-test/tags",
    "blobs_url": "https://github.ibm.com/api/v3/repos/cloud-sre/pnp-notification-messages-test/git/blobs{/sha}",
    "git_tags_url": "https://github.ibm.com/api/v3/repos/cloud-sre/pnp-notification-messages-test/git/tags{/sha}",
    "git_refs_url": "https://github.ibm.com/api/v3/repos/cloud-sre/pnp-notification-messages-test/git/refs{/sha}",
    "trees_url": "https://github.ibm.com/api/v3/repos/cloud-sre/pnp-notification-messages-test/git/trees{/sha}",
    "statuses_url": "https://github.ibm.com/api/v3/repos/cloud-sre/pnp-notification-messages-test/statuses/{sha}",
    "languages_url": "https://github.ibm.com/api/v3/repos/cloud-sre/pnp-notification-messages-test/languages",
    "stargazers_url": "https://github.ibm.com/api/v3/repos/cloud-sre/pnp-notification-messages-test/stargazers",
    "contributors_url": "https://github.ibm.com/api/v3/repos/cloud-sre/pnp-notification-messages-test/contributors",
    "subscribers_url": "https://github.ibm.com/api/v3/repos/cloud-sre/pnp-notification-messages-test/subscribers",
    "subscription_url": "https://github.ibm.com/api/v3/repos/cloud-sre/pnp-notification-messages-test/subscription",
    "commits_url": "https://github.ibm.com/api/v3/repos/cloud-sre/pnp-notification-messages-test/commits{/sha}",
    "git_commits_url": "https://github.ibm.com/api/v3/repos/cloud-sre/pnp-notification-messages-test/git/commits{/sha}",
    "comments_url": "https://github.ibm.com/api/v3/repos/cloud-sre/pnp-notification-messages-test/comments{/number}",
    "issue_comment_url": "https://github.ibm.com/api/v3/repos/cloud-sre/pnp-notification-messages-test/issues/comments{/number}",
    "contents_url": "https://github.ibm.com/api/v3/repos/cloud-sre/pnp-notification-messages-test/contents/{+path}",
    "compare_url": "https://github.ibm.com/api/v3/repos/cloud-sre/pnp-notification-messages-test/compare/{base}...{head}",
    "merges_url": "https://github.ibm.com/api/v3/repos/cloud-sre/pnp-notification-messages-test/merges",
    "archive_url": "https://github.ibm.com/api/v3/repos/cloud-sre/pnp-notification-messages-test/{archive_format}{/ref}",
    "downloads_url": "https://github.ibm.com/api/v3/repos/cloud-sre/pnp-notification-messages-test/downloads",
    "issues_url": "https://github.ibm.com/api/v3/repos/cloud-sre/pnp-notification-messages-test/issues{/number}",
    "pulls_url": "https://github.ibm.com/api/v3/repos/cloud-sre/pnp-notification-messages-test/pulls{/number}",
    "milestones_url": "https://github.ibm.com/api/v3/repos/cloud-sre/pnp-notification-messages-test/milestones{/number}",
    "notifications_url": "https://github.ibm.com/api/v3/repos/cloud-sre/pnp-notification-messages-test/notifications{?since,all,participating}",
    "labels_url": "https://github.ibm.com/api/v3/repos/cloud-sre/pnp-notification-messages-test/labels{/name}",
    "releases_url": "https://github.ibm.com/api/v3/repos/cloud-sre/pnp-notification-messages-test/releases{/id}",
    "deployments_url": "https://github.ibm.com/api/v3/repos/cloud-sre/pnp-notification-messages-test/deployments",
    "created_at": "2019-07-20T20:22:05Z",
    "updated_at": "2019-07-27T13:17:49Z",
    "pushed_at": "2019-07-27T13:17:48Z",
    "git_url": "git://github.ibm.com/cloud-sre/pnp-notification-messages-test.git",
    "ssh_url": "git@github.ibm.com:cloud-sre/pnp-notification-messages-test.git",
    "clone_url": "https://github.ibm.com/cloud-sre/pnp-notification-messages-test.git",
    "svn_url": "https://github.ibm.com/cloud-sre/pnp-notification-messages-test",
    "homepage": null,
    "size": 6,
    "stargazers_count": 0,
    "watchers_count": 0,
    "language": null,
    "has_issues": true,
    "has_projects": true,
    "has_downloads": true,
    "has_wiki": true,
    "has_pages": false,
    "forks_count": 0,
    "mirror_url": null,
    "archived": false,
    "open_issues_count": 6,
    "license": null,
    "forks": 0,
    "open_issues": 6,
    "watchers": 0,
    "default_branch": "master"
  },
  "organization": {
    "login": "cloud-sre",
    "id": 10240,
    "node_id": "MDEyOk9yZ2FuaXphdGlvbjEwMjQw",
    "url": "https://github.ibm.com/api/v3/orgs/cloud-sre",
    "repos_url": "https://github.ibm.com/api/v3/orgs/cloud-sre/repos",
    "events_url": "https://github.ibm.com/api/v3/orgs/cloud-sre/events",
    "hooks_url": "https://github.ibm.com/api/v3/orgs/cloud-sre/hooks",
    "issues_url": "https://github.ibm.com/api/v3/orgs/cloud-sre/issues",
    "members_url": "https://github.ibm.com/api/v3/orgs/cloud-sre/members{/member}",
    "public_members_url": "https://github.ibm.com/api/v3/orgs/cloud-sre/public_members{/member}",
    "avatar_url": "https://avatars.github.ibm.com/u/10240?",
    "description": null
  },
  "sender": {
    "login": "kparzygn",
    "id": 5122,
    "node_id": "MDQ6VXNlcjUxMjI=",
    "avatar_url": "https://avatars.github.ibm.com/u/5122?",
    "gravatar_id": "",
    "url": "https://github.ibm.com/api/v3/users/kparzygn",
    "html_url": "https://github.ibm.com/kparzygn",
    "followers_url": "https://github.ibm.com/api/v3/users/kparzygn/followers",
    "following_url": "https://github.ibm.com/api/v3/users/kparzygn/following{/other_user}",
    "gists_url": "https://github.ibm.com/api/v3/users/kparzygn/gists{/gist_id}",
    "starred_url": "https://github.ibm.com/api/v3/users/kparzygn/starred{/owner}{/repo}",
    "subscriptions_url": "https://github.ibm.com/api/v3/users/kparzygn/subscriptions",
    "organizations_url": "https://github.ibm.com/api/v3/users/kparzygn/orgs",
    "repos_url": "https://github.ibm.com/api/v3/users/kparzygn/repos",
    "events_url": "https://github.ibm.com/api/v3/users/kparzygn/events{/privacy}",
    "received_events_url": "https://github.ibm.com/api/v3/users/kparzygn/received_events",
    "type": "User",
    "site_admin": false
  }
}`
