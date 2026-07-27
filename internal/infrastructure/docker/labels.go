package docker

import (
	"fmt"
	"strings"
)

func GenerateTraefikLabels(projectName, domain, containerPort string) map[string]string {
	sanitized := sanitizeName(projectName)
	
	return map[string]string{
		"traefik.enable": "true",
		fmt.Sprintf("traefik.http.routers.%s.rule", sanitized): fmt.Sprintf("Host(`%s`)", domain),
		fmt.Sprintf("traefik.http.routers.%s.entrypoints", sanitized): "websecure",
		fmt.Sprintf("traefik.http.services.%s.loadbalancer.server.port", sanitized): containerPort,
		"pontoon.project": projectName,
		"pontoon.managed": "true",
	}
}

func GenerateContainerLabels(tenantID, projectID, projectName, domain, containerPort string) map[string]string {
	labels := GenerateTraefikLabels(projectName, domain, containerPort)
	
	labels["pontoon.tenant"] = tenantID
	labels["pontoon.project-id"] = projectID
	
	return labels
}

func sanitizeName(name string) string {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, ".", "-")
	name = strings.ReplaceAll(name, "_", "-")
	return name
}
