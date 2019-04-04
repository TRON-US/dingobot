package github

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/google/go-github/github"
	"github.com/sluongng/dingbot"
)

const (
	refHeadsPre = "refs/heads/"
	refTagsPre  = "refs/tags/"
)

var (
	webhookSecret []byte
)

func init() {
	ws := os.Getenv("WEBHOOK_SECRET")
	if ws == "" {
		panic("WEBHOOK_SECRET is unset (https://developer.github.com/webhooks/securing/)")
	}
	webhookSecret = []byte(ws)
}

func Handle(w http.ResponseWriter, r *http.Request) {
	// Get access_token for dingtalk api
	accessToken := r.FormValue("access_token")
	if accessToken == "" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Validate and parse github webhook payload
	payload, err := github.ValidatePayload(r, webhookSecret)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	event, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var brief, title, url, text string
	switch event := event.(type) {
	case *github.PushEvent:
		var ref, refType string
		if strings.HasPrefix(*event.Ref, refHeadsPre) {
			refType = "branch"
			ref = (*event.Ref)[len(refHeadsPre):]
		} else if strings.HasPrefix(*event.Ref, refTagsPre) {
			refType = "tag"
			ref = (*event.Ref)[len(refTagsPre):]
		}
		brief = fmt.Sprintf("Push to %s", ref)
		action := "pushed to"
		if *event.Created {
			action = "created"
		} else if *event.Deleted {
			action = "deleted"
		} else if *event.Forced {
			action = "force pushed to"
		}
		title = fmt.Sprintf(`\[%s\] %s %s %s %s`,
			*event.Repo.Name,
			*event.Pusher.Name,
			action,
			refType,
			ref,
		)
		url = *event.Compare
		for _, commit := range event.Commits {
			text += fmt.Sprintf("[%s](%s) %s - %s\n\n", (*commit.ID)[:7], *commit.URL, *commit.Message, *commit.Committer.Name)
		}
	case *github.PullRequestEvent:
		if *event.Action == "synchronize" {
			return
		}
		brief = fmt.Sprintf("PR #%d", *event.Number)
		var info string
		if *event.Action == "review_requested" {
			info = fmt.Sprintf(": [%s]", *event.RequestedReviewer.Login)
		}
		title = fmt.Sprintf(`\[%s\] Pull request #%d **%s** by %s%s`,
			*event.Repo.Name,
			*event.Number,
			*event.Action,
			*event.Sender.Login,
			info,
		)
		url = *event.PullRequest.HTMLURL
		text = *event.PullRequest.Title + "\n\n" + *event.PullRequest.Body
	case *github.PullRequestReviewEvent:
		brief = fmt.Sprintf("PR #%d review", *event.PullRequest.Number)
		title = fmt.Sprintf(`\[%s\] Pull request #%d review %s: **%s** by %s`,
			*event.Repo.Name,
			*event.PullRequest.Number,
			*event.Action,
			*event.Review.State,
			*event.Sender.Login,
		)
		url = *event.Review.HTMLURL
		if event.Review.Body == nil {
			return
		}
		text = *event.Review.Body
	case *github.PullRequestReviewCommentEvent:
		brief = fmt.Sprintf("PR #%d comment", *event.PullRequest.Number)
		title = fmt.Sprintf(`\[%s\] Pull request #%d review comment **%s** by %s`,
			*event.Repo.Name,
			*event.PullRequest.Number,
			*event.Action,
			*event.Sender.Login,
		)
		url = *event.Comment.HTMLURL
		text = *event.Comment.Body
	case *github.IssueCommentEvent:
		brief = fmt.Sprintf("Issue #%d comment", *event.Issue.Number)
		title = fmt.Sprintf(`\[%s\] Issue/pull request #%d comment **%s** by %s`,
			*event.Repo.Name,
			*event.Issue.Number,
			*event.Action,
			*event.Sender.Login,
		)
		url = *event.Comment.HTMLURL
		text = *event.Comment.Body
	case *github.CommitCommentEvent:
		brief = fmt.Sprintf("Commit %s comment", (*event.Comment.CommitID)[:7])
		title = fmt.Sprintf(`\[%s\] Commit %s comment **%s** by %s`,
			*event.Repo.Name,
			(*event.Comment.CommitID)[:7],
			*event.Action,
			*event.Sender.Login,
		)
		url = *event.Comment.HTMLURL
		text = *event.Comment.Body
	default:
		return
	}

	// Replace all (double for markdown) linebreaks with indentation
	text = fmt.Sprintf("#### [%s](%s)\n\n%s", title, url, text)
	text = strings.ReplaceAll(text, "\n\n", "\n\n> ")
	text = strings.ReplaceAll(text, "\r\n\r\n", "\n\n> ")

	err = dingbot.NewMarkdownMessage(
		brief,
		text,
		dingbot.EmptyAtTag()).Send(accessToken)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}
