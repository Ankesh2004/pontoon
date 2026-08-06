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
	DeploymentID  string
	Domain        string
	MemoryLimitMB int
	CPULimit      float64
}

func (c *Client) RunContainer(ctx context.Context, cfg RunConfig) (string, error) {
	// Enforce Strict PORT Contract
	port := "8080" // Default industry standard port
	if p, ok := cfg.EnvVars["PORT"]; ok && p != "" {
		port = p
	}

	var env []string
	hasPort := false
	for k, v := range cfg.EnvVars {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
		if k == "PORT" {
			hasPort = true
		}
	}
	
	// Inject PORT if not explicitly provided in EnvVars
	if !hasPort {
		env = append(env, fmt.Sprintf("PORT=%s", port))
	}


	labels := GenerateContainerLabels(
		cfg.TenantID,
		cfg.ProjectID,
		cfg.ProjectName,
		cfg.DeploymentID,
		cfg.Domain,
		port,
	)

	hostConfig := &container.HostConfig{
		Resources: container.Resources{
			Memory:     int64(cfg.MemoryLimitMB) * 1024 * 1024,
			NanoCPUs:   int64(cfg.CPULimit * 1e9),
		},
		LogConfig: container.LogConfig{
			Type: "json-file",
			Config: map[string]string{
				"max-size": "10m",
				"max-file": "3",
			},
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
			nat.Port(port + "/tcp"): {},
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
