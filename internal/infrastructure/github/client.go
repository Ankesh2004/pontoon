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
