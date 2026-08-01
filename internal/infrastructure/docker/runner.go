package docker

import (
	"bytes"
	"context"
	"fmt"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
)

type RunConfig struct {
	Image         string
	ContainerName string
	EnvVars       map[string]string
	TenantID      string
	ProjectID     string
	ProjectName   string
	Domain        string
	MemoryLimitMB int
	CPULimit      float64
}

func (c *Client) RunContainer(ctx context.Context, cfg RunConfig) (string, error) {
	env := make([]string, 0, len(cfg.EnvVars))
	for k, v := range cfg.EnvVars {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	labels := GenerateContainerLabels(
		cfg.TenantID,
		cfg.ProjectID,
		cfg.ProjectName,
		cfg.Domain,
		"80", // most containers (nginx etc) expose 80, not 8080
	)

	hostConfig := &container.HostConfig{
		Resources: container.Resources{
			Memory:     int64(cfg.MemoryLimitMB) * 1024 * 1024,
			NanoCPUs:   int64(cfg.CPULimit * 1e9),
		},
		SecurityOpt: []string{"no-new-privileges:true"},
		CapDrop: []string{
			"SYS_ADMIN",
			"SYS_PTRACE",
			"SYS_RAWIO",
			"SYS_MODULE",
			"SYS_BOOT",
			"SYS_TIME",
			"NET_ADMIN",
			"NET_RAW",
			"MKNOD",
			"AUDIT_WRITE",
			"MAC_ADMIN",
			"MAC_OVERRIDE",
		},
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeTmpfs,
				Target: "/tmp",
			},
		},
	}

	networkConfig := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			IngressNetwork: {},
		},
	}

	containerConfig := &container.Config{
		Image:  cfg.Image,
		Env:    env,
		Labels: labels,
		ExposedPorts: map[nat.Port]struct{}{
			"80/tcp": {},
		},
	}

	resp, err := c.ContainerCreate(ctx, containerConfig, hostConfig, networkConfig, nil, cfg.ContainerName)
	if err != nil {
		return "", fmt.Errorf("failed to create container: %w", err)
	}

	tenantNetwork, err := c.EnsureTenantNetwork(ctx, cfg.TenantID)
	if err != nil {
		return "", fmt.Errorf("failed to ensure tenant network: %w", err)
	}

	if err := c.ConnectToNetwork(ctx, resp.ID, tenantNetwork); err != nil {
		return "", fmt.Errorf("failed to connect to tenant network: %w", err)
	}

	if err := c.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return "", fmt.Errorf("failed to start container: %w", err)
	}

	return resp.ID, nil
}

func (c *Client) StopContainer(ctx context.Context, containerID string) error {
	timeout := 30
	return c.ContainerStop(ctx, containerID, container.StopOptions{
		Timeout: &timeout,
	})
}

func (c *Client) RemoveContainer(ctx context.Context, containerID string) error {
	return c.ContainerRemove(ctx, containerID, container.RemoveOptions{
		Force: true,
	})
}

func (c *Client) GetContainerLogs(ctx context.Context, containerID string, tail int) (string, error) {
	options := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     false,
	}

	if tail > 0 {
		options.Tail = fmt.Sprintf("%d", tail)
	}

	logs, err := c.ContainerLogs(ctx, containerID, options)
	if err != nil {
		return "", fmt.Errorf("failed to get container logs: %w", err)
	}
	defer logs.Close()

	outBuf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)
	if _, err := stdcopy.StdCopy(outBuf, errBuf, logs); err != nil {
		return "", fmt.Errorf("failed to demultiplex logs: %w", err)
	}

	return outBuf.String() + errBuf.String(), nil
}
