package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Ankesh2004/pontoon/internal/domain"
	"github.com/Ankesh2004/pontoon/internal/usecase"
)

type WebhookHandler struct {
	webhookUC *usecase.WebhookUseCase
}

func NewWebhookHandler(webhookUC *usecase.WebhookUseCase) *WebhookHandler {
	return &WebhookHandler{
		webhookUC: webhookUC,
	}
}

type GitHubWebhookPayload struct {
	Action     string `json:"action"`
	Ref        string `json:"ref"`
	Repository struct {
		FullName string `json:"full_name"`
		CloneURL string `json:"clone_url"`
	} `json:"repository"`
	HeadCommit struct {
		ID string `json:"id"`
	} `json:"head_commit"`
}

func (h *WebhookHandler) Handle(w http.ResponseWriter, r *http.Request) {
	signature := r.Header.Get("X-Hub-Signature-256")
	if signature == "" {
		http.Error(w, "missing signature", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	projectID := r.URL.Query().Get("project_id")
	if projectID == "" {
		http.Error(w, "project_id query parameter required", http.StatusBadRequest)
		return
	}

	if err := h.verifySignature(signature, body, projectID); err != nil {
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}

	var payload GitHubWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	if payload.Action != "" && payload.Action != "created" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if !strings.HasPrefix(payload.Ref, "refs/heads/") {
		w.WriteHeader(http.StatusOK)
		return
	}

	branch := strings.TrimPrefix(payload.Ref, "refs/heads/")

	deployment, err := h.webhookUC.TriggerDeploymentFromWebhook(
		r.Context(),
		projectID,
		branch,
		payload.HeadCommit.ID,
	)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			w.WriteHeader(http.StatusOK)
			return
		}
		http.Error(w, fmt.Sprintf("failed to trigger deployment: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"deployment_id": deployment.ID,
		"status":        deployment.Status,
	})
}

func (h *WebhookHandler) verifySignature(signature string, body []byte, projectID string) error {
	if !strings.HasPrefix(signature, "sha256=") {
		return errors.New("invalid signature format")
	}

	signatureHex := strings.TrimPrefix(signature, "sha256=")
	signatureBytes, err := hex.DecodeString(signatureHex)
	if err != nil {
		return errors.New("invalid signature encoding")
	}

	secret, err := h.webhookUC.GetWebhookSecret(projectID)
	if err != nil {
		return fmt.Errorf("failed to get webhook secret: %w", err)
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expectedMAC := mac.Sum(nil)

	if !hmac.Equal(signatureBytes, expectedMAC) {
		return errors.New("signature mismatch")
	}

	return nil
}
