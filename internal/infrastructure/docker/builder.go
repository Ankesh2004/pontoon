package docker

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/archive"
)

type BuildConfig struct {
	RepoURL   string
	Branch    string
	ImageTag  string
	WorkDir   string
}

func (c *Client) BuildImage(ctx context.Context, cfg BuildConfig) (string, error) {
	buildContext, err := createBuildContext(cfg.WorkDir)
	if err != nil {
		return "", fmt.Errorf("failed to create build context: %w", err)
	}
	defer buildContext.Close()

	buildOptions := types.ImageBuildOptions{
		Tags:       []string{cfg.ImageTag},
		Dockerfile: "Dockerfile",
		Remove:     true,
		NoCache:    false,
	}

	resp, err := c.ImageBuild(ctx, buildContext, buildOptions)
	if err != nil {
		return "", fmt.Errorf("failed to build image: %w", err)
	}
	defer resp.Body.Close()

	if _, err := io.Copy(io.Discard, resp.Body); err != nil {
		return "", fmt.Errorf("failed to read build response: %w", err)
	}

	return cfg.ImageTag, nil
}

func createBuildContext(contextDir string) (io.ReadCloser, error) {
	if _, err := os.Stat(contextDir); err != nil {
		return nil, fmt.Errorf("context directory does not exist: %w", err)
	}

	return archive.Tar(contextDir, archive.Uncompressed)
}

func (c *Client) RemoveImage(ctx context.Context, imageID string) error {
	_, err := c.ImageRemove(ctx, imageID, types.ImageRemoveOptions{
		Force:         true,
		PruneChildren: true,
	})
	return err
}

func createTarFromDir(srcDir string) (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)
	defer tw.Close()

	err := filepath.Walk(srcDir, func(file string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(fi, fi.Name())
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(srcDir, file)
		if err != nil {
			return err
		}
		header.Name = relPath

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if fi.IsDir() {
			return nil
		}

		f, err := os.Open(file)
		if err != nil {
			return err
		}
		defer f.Close()

		if _, err := io.Copy(tw, f); err != nil {
			return err
		}

		return nil
	})

	return buf, err
}
