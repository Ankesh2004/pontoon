package tasks

import "encoding/json"

const (
	TypeDeploy      = "deploy"
	TypeReapStuck   = "deployments:reap"
	TypeRollback    = "deployments:rollback"
	TypePruneImages = "images:prune"
)

type DeployPayload struct {
	DeploymentID string            `json:"deployment_id"`
	ProjectID    string            `json:"project_id"`
	UserID       string            `json:"user_id"`
	RepoURL      string            `json:"repo_url"`
	Branch       string            `json:"branch"`
	CommitSHA    string            `json:"commit_sha"`
	Domain       string            `json:"domain"`
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

type RollbackPayload struct {
	DeploymentID string            `json:"deployment_id"`
	ProjectID    string            `json:"project_id"`
	UserID       string            `json:"user_id"`
	Domain       string            `json:"domain"`
	TargetImage  string            `json:"target_image"`
}

func (p *RollbackPayload) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

func UnmarshalRollbackPayload(data []byte) (*RollbackPayload, error) {
	var payload RollbackPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	return &payload, nil
}
