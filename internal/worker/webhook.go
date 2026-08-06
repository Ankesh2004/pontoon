package worker

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/hibiken/asynq"
	"golang.org/x/oauth2"

	gh "github.com/Ankesh2004/pontoon/internal/infrastructure/github"
	"github.com/Ankesh2004/pontoon/internal/tasks"
)

func (p *DeployProcessor) ProcessDeleteWebhookTask(ctx context.Context, t *asynq.Task) error {
	payload, err := tasks.UnmarshalDeleteWebhookPayload(t.Payload())
	if err != nil {
		return fmt.Errorf("failed to unmarshal webhook payload: %w", err)
	}

	slog.Info("processing delete webhook task", "repo_owner", payload.RepoOwner, "repo_name", payload.RepoName, "webhook_url", payload.WebhookURL)

	user, err := p.userRepo.GetByID(payload.UserID)
	if err != nil || user == nil {
		return fmt.Errorf("failed to get user for webhook deletion: %w", err)
	}

	if user.AccessToken == "" {
		slog.Warn("skipping webhook deletion: user has no access token", "user_id", user.ID)
		return nil
	}

	// Create github client with user's token
	token := &oauth2.Token{AccessToken: user.AccessToken}
	ghClient := gh.NewClient(ctx, token)

	err = ghClient.DeleteWebhook(ctx, payload.RepoOwner, payload.RepoName, payload.WebhookURL)
	if err != nil {
		return fmt.Errorf("failed to delete webhook from github: %w", err)
	}

	slog.Info("successfully deleted webhook from github", "repo", payload.RepoOwner+"/"+payload.RepoName)
	return nil
}
