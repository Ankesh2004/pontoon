package usecase

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/container"

	"github.com/Ankesh2004/pontoon/internal/domain"
	"github.com/Ankesh2004/pontoon/internal/infrastructure/docker"
)

type CapacityUseCase struct {
	dockerClient          *docker.Client
	maxContainerMemoryMB  int
	totalContainerMemoryMB int
}

func NewCapacityUseCase(
	dockerClient *docker.Client,
	maxContainerMemoryMB int,
	totalContainerMemoryMB int,
) *CapacityUseCase {
	return &CapacityUseCase{
		dockerClient:          dockerClient,
		maxContainerMemoryMB:  maxContainerMemoryMB,
		totalContainerMemoryMB: totalContainerMemoryMB,
	}
}

func (uc *CapacityUseCase) CheckCapacity(ctx context.Context, requestedMemoryMB int) error {
	if requestedMemoryMB > uc.maxContainerMemoryMB {
		return fmt.Errorf("%w: requested %dMB exceeds max %dMB per container",
			domain.ErrCapacityExceeded, requestedMemoryMB, uc.maxContainerMemoryMB)
	}

	containers, err := uc.dockerClient.ContainerList(ctx, container.ListOptions{
		All: false,
	})
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	var totalUsedMB int
	for _, ctr := range containers {
		if _, ok := ctr.Labels["pontoon.managed"]; !ok {
			continue
		}

		inspect, err := uc.dockerClient.ContainerInspect(ctx, ctr.ID)
		if err != nil {
			continue
		}

		memoryMB := int(inspect.HostConfig.Memory / (1024 * 1024))
		totalUsedMB += memoryMB
	}

	availableMB := uc.totalContainerMemoryMB - totalUsedMB
	if requestedMemoryMB > availableMB {
		return fmt.Errorf("%w: cluster full, available %dMB, requested %dMB",
			domain.ErrCapacityExceeded, availableMB, requestedMemoryMB)
	}

	return nil
}

func (uc *CapacityUseCase) GetAvailableMemoryMB(ctx context.Context) (int, error) {
	containers, err := uc.dockerClient.ContainerList(ctx, container.ListOptions{
		All: false,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to list containers: %w", err)
	}

	var totalUsedMB int
	for _, ctr := range containers {
		if _, ok := ctr.Labels["pontoon.managed"]; !ok {
			continue
		}

		inspect, err := uc.dockerClient.ContainerInspect(ctx, ctr.ID)
		if err != nil {
			continue
		}

		memoryMB := int(inspect.HostConfig.Memory / (1024 * 1024))
		totalUsedMB += memoryMB
	}

	return uc.totalContainerMemoryMB - totalUsedMB, nil
}
