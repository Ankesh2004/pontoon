package worker

import "encoding/json"

const (
	TypeDeploy = "deploy"
)

type DeployPayload struct {
	DeploymentID string `json:"deployment_id"`
	ProjectID    string `json:"project_id"`
	UserID       string `json:"user_id"`
	RepoURL      string `json:"repo_url"`
	Branch       string `json:"branch"`
	CommitSHA    string `json:"commit_sha"`
	Domain       string `json:"domain"`
	EnvVars      map[string]string `json:"env_vars"`
}

func (p *DeployPayload) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

func UnmarshalDeployPayload(data []byte) (*DeployPayload, error) {
	var payload DeployPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	return &payload, nil
}
