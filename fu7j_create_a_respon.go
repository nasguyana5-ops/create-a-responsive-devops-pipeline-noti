package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/slack-go/slack"
)

type Notifier struct {
	slackToken string
	slackChannel string
}

func (n *Notifier) SendNotification(ctx context.Context, buildResult string) error {
/slack notification logic
	slackClient := slack.New(n.slackToken)
	params := &slack.PostWebhookParameters{
		URLs: []slack.WebhookURL{
			{
				URL:  "https://" + n.slackChannel,
				Icon: "https://example.com/icon.png",
			},
		},
	}
	err := slackClient.PostWebhookContext(ctx, params, slack.NewWebhookMessage(buildResult))
	return err
}

func main() {
	notifier := &Notifier{
		slackToken: os.Getenv("SLACK_TOKEN"),
		slackChannel: os.Getenv("SLACK_CHANNEL"),
	}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(10 * time.Second))

	r.Post("/notify", func(w http.ResponseWriter, r *http.Request) {
		buildResult := r.URL.Query().Get("build_result")
		if buildResult == "" {
			http.Error(w, "build_result is required", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err := notifier.SendNotification(ctx, buildResult)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusAccepted)
	})

	fmt.Println("Notifier is running on port 8080")
	http.ListenAndServe(":8080", r)
}