# dingobot
Dingtalk bot servers written in Golang.

## Build and run Github bot

Golang 1.11+
```bash
$ GO111MODULE=on go build -o dingobot main.go
$ WEBHOOK_SECRET='xxx' HOST=':8080' ./dingobot
```

`WEBHOOK_SECRET` is from the specific Github webhook's settings page.

The webhook URL on the settings page should be set to `http[s]://<server_ip>:8080/github?access_token=xxx` where `access_token` is taken from the Dingtalk group bot's settings page.

## Supported Github webhook events
* Commit Comments
* Issue Comments
* Pull requests
* Pull request reviews
* Pull request review comments
* Pushes

Pull requests are welcome on supporting other events.
