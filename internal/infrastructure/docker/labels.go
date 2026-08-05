package docker

import (
	"fmt"
	"strings"
)

// figures out if we're dealing with a local domain (no TLS available)
func isLocalDomain(domain string) bool {
	return strings.Contains(domain, "localhost") || strings.Contains(domain, "127.0.0.1")
}

func GenerateTraefikLabels(projectName, projectID, domain, containerPort string) map[string]string {
	sanitized := sanitizeName(projectName) + "-" + projectID[:8]

	// localhost can't do TLS, so we use the plain HTTP entrypoint
	entrypoint := "websecure"
	if isLocalDomain(domain) {
		entrypoint = "web"
	}

	labels := map[string]string{
		"traefik.enable": "true",
		fmt.Sprintf("traefik.http.routers.%s.rule", sanitized):                      fmt.Sprintf("Host(`%s`)", domain),
		fmt.Sprintf("traefik.http.routers.%s.entrypoints", sanitized):               entrypoint,
		fmt.Sprintf("traefik.http.services.%s.loadbalancer.server.port", sanitized): containerPort,
		"pontoon.project": projectName,
		"pontoon.managed": "true",
	}

	// only bother with cert resolver on real domains
	if !isLocalDomain(domain) {
		labels[fmt.Sprintf("traefik.http.routers.%s.tls.certresolver", sanitized)] = "letsencrypt"
	}

	return labels
}

func GenerateContainerLabels(tenantID, projectID, projectName, domain, containerPort string) map[string]string {
	labels := GenerateTraefikLabels(projectName, projectID, domain, containerPort)
	
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

