package github

import (
	"fmt"
	"net/http"
	"os"

	"github.com/google/go-github/github"
	"github.com/sluongng/dingbot"
)

const (
	refsPre = "refs/heads/"
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

	var title, url, text string
	switch event := event.(type) {
	case *github.PushEvent:
		title = fmt.Sprintf(`\[%s\] %s pushed to %s`,
			*event.Repo.Name,
			*event.Pusher.Name,
			(*event.Ref)[len(refsPre):],
		)
		url = *event.Compare
		for _, commit := range event.Commits {
			text += fmt.Sprintf("> [%s](%s) %s - %s\n\n", (*commit.ID)[:7], *commit.URL, *commit.Message, *commit.Committer.Name)
		}
	case *github.PullRequestEvent:
		title = fmt.Sprintf(`\[%s\] Pull request #%d %s by %s`,
			*event.Repo.Name,
			*event.Number,
			*event.Action,
			*event.Sender.Login,
		)
		url = *event.PullRequest.HTMLURL
		text = " > " + *event.PullRequest.Title + "\n\n> " + *event.PullRequest.Body
	case *github.PullRequestReviewEvent:
		title = fmt.Sprintf(`\[%s\] Pull request #%d review %s: %s by %s`,
			*event.Repo.Name,
			*event.PullRequest.Number,
			*event.Action,
			*event.Review.State,
			*event.Sender.Login,
		)
		url = *event.Review.HTMLURL
		body := ""
		if event.Review.Body != nil {
			body = *event.Review.Body
		}
		text = "> " + body
	case *github.PullRequestReviewCommentEvent:
		title = fmt.Sprintf(`\[%s\] Pull request #%d review comment %s by %s`,
			*event.Repo.Name,
			*event.PullRequest.Number,
			*event.Action,
			*event.Sender.Login,
		)
		url = *event.Comment.HTMLURL
		text = "> " + *event.Comment.Body
	case *github.IssueCommentEvent:
		title = fmt.Sprintf(`\[%s\] Issue/pull request #%d comment %s by %s`,
			*event.Repo.Name,
			*event.Issue.Number,
			*event.Action,
			*event.Sender.Login,
		)
		url = *event.Issue.HTMLURL
		text = "> " + *event.Issue.Body
	default:
		return
	}

	err = dingbot.NewMarkdownMessage(
		title,
		fmt.Sprintf("#### [%s](%s)\n%s", title, url, text),
		dingbot.EmptyAtTag()).Send(accessToken)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}
