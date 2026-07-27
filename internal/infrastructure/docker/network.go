package docker

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
)

const (
	IngressNetwork = "pontoon-ingress"
)

func (c *Client) EnsureIngressNetwork(ctx context.Context) error {
	return c.ensureNetwork(ctx, IngressNetwork, false)
}

func (c *Client) EnsureTenantNetwork(ctx context.Context, tenantID string) (string, error) {
	networkName := fmt.Sprintf("pontoon-%s", tenantID)
	if err := c.ensureNetwork(ctx, networkName, true); err != nil {
		return "", err
	}
	return networkName, nil
}

func (c *Client) ensureNetwork(ctx context.Context, name string, internal bool) error {
	filterArgs := filters.NewArgs()
	filterArgs.Add("name", name)

	networks, err := c.NetworkList(ctx, types.NetworkListOptions{
		Filters: filterArgs,
	})
	if err != nil {
		return fmt.Errorf("failed to list networks: %w", err)
	}

	for _, n := range networks {
		if n.Name == name {
			return nil
		}
	}

	_, err = c.NetworkCreate(ctx, name, types.NetworkCreate{
		Driver:   "bridge",
		Internal: internal,
		Labels: map[string]string{
			"pontoon.managed": "true",
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create network %s: %w", name, err)
	}

	return nil
}

func (c *Client) ConnectToNetwork(ctx context.Context, containerID, networkName string) error {
	return c.NetworkConnect(ctx, networkName, containerID, nil)
}

func (c *Client) RemoveTenantNetwork(ctx context.Context, tenantID string) error {
	networkName := fmt.Sprintf("pontoon-%s", tenantID)
	return c.NetworkRemove(ctx, networkName)
}
