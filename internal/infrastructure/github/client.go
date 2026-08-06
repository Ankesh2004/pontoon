package github

import (
	"context"
	"fmt"

	"github.com/google/go-github/v62/github"
	"golang.org/x/oauth2"

	"github.com/Ankesh2004/pontoon/internal/domain"
)

type Client struct {
	client *github.Client
}

func NewClient(ctx context.Context, token *oauth2.Token) *Client {
	ts := oauth2.StaticTokenSource(token)
	tc := oauth2.NewClient(ctx, ts)
	return &Client{
		client: github.NewClient(tc),
	}
}

func (c *Client) GetUser(ctx context.Context) (*domain.User, error) {
	ghUser, _, err := c.client.Users.Get(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get github user: %w", err)
	}

	return c.toDomainUser(ghUser), nil
}

func (c *Client) GetUserByID(ctx context.Context, id int64) (*domain.User, error) {
	ghUser, _, err := c.client.Users.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get github user by id: %w", err)
	}

	return c.toDomainUser(ghUser), nil
}

func (c *Client) toDomainUser(ghUser *github.User) *domain.User {
	email := ""
	if ghUser.Email != nil {
		email = *ghUser.Email
	}

	return &domain.User{
		GitHubID:       int(ghUser.GetID()),
		GitHubUsername: ghUser.GetLogin(),
		Email:          email,
		AvatarURL:      ghUser.GetAvatarURL(),
	}
}

func (c *Client) CreateWebhook(ctx context.Context, owner, repo, webhookURL, webhookSecret string) error {
	hook := &github.Hook{
		Name:   github.String("web"),
		Active: github.Bool(true),
		Events: []string{"push"},
		Config: &github.HookConfig{
			URL:          github.String(webhookURL),
			ContentType:  github.String("json"),
			Secret:       github.String(webhookSecret),
			InsecureSSL:  github.String("0"),
		},
	}

	_, _, err := c.client.Repositories.CreateHook(ctx, owner, repo, hook)
	if err != nil {
		return fmt.Errorf("failed to create webhook: %w", err)
	}

	return nil
}

func (c *Client) DeleteWebhook(ctx context.Context, owner, repo, webhookURL string) error {
	opts := &github.ListOptions{PerPage: 100}
	
	for {
		hooks, resp, err := c.client.Repositories.ListHooks(ctx, owner, repo, opts)
		if err != nil {
			return fmt.Errorf("failed to list webhooks: %w", err)
		}

		for _, hook := range hooks {
			if hook.Config != nil && hook.Config.URL != nil && *hook.Config.URL == webhookURL {
				_, err := c.client.Repositories.DeleteHook(ctx, owner, repo, hook.GetID())
				if err != nil {
					return fmt.Errorf("failed to delete webhook %d: %w", hook.GetID(), err)
				}
				return nil
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	// Not found, which is fine (already deleted or never created)
	return nil
}
