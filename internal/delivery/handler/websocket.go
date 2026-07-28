package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	goredis "github.com/redis/go-redis/v9"

	"github.com/Ankesh2004/pontoon/internal/domain"
	infraredis "github.com/Ankesh2004/pontoon/internal/infrastructure/redis"
	"github.com/Ankesh2004/pontoon/internal/usecase"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WebSocketHandler struct {
	deploymentUC *usecase.DeploymentUseCase
	redisClient  *goredis.Client
}

func NewWebSocketHandler(
	deploymentUC *usecase.DeploymentUseCase,
	redisClient *goredis.Client,
) *WebSocketHandler {
	return &WebSocketHandler{
		deploymentUC: deploymentUC,
		redisClient:  redisClient,
	}
}

func (h *WebSocketHandler) StreamLogs(w http.ResponseWriter, r *http.Request) {
	deploymentID := chi.URLParam(r, "deploymentId")
	if deploymentID == "" {
		http.Error(w, "deployment id required", http.StatusBadRequest)
		return
	}

	ticket := r.URL.Query().Get("ticket")
	if ticket == "" {
		http.Error(w, "ticket query parameter required", http.StatusBadRequest)
		return
	}

	key := "ws_ticket:" + ticket
	userID, err := h.redisClient.GetDel(r.Context(), key).Result()
	if err != nil {
		http.Error(w, "invalid or expired ticket", http.StatusUnauthorized)
		return
	}

	deployment, err := h.deploymentUC.GetDeployment(r.Context(), userID, deploymentID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) || err.Error() == domain.ErrNotFound.Error() {
			http.Error(w, "deployment not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, domain.ErrForbidden) || err.Error() == domain.ErrForbidden.Error() {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		http.Error(w, "failed to get deployment", http.StatusInternalServerError)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	if isTerminalStatus(deployment.Status) {
		finalMsg := infraredis.LogMessage{
			DeploymentID: deploymentID,
			Line:         fmt.Sprintf("[DEPLOYMENT %s]", strings.ToUpper(string(deployment.Status))),
			Timestamp:    time.Now().Unix(),
		}
		conn.WriteJSON(finalMsg)
		return
	}

	channel := fmt.Sprintf("deployment:%s:logs", deploymentID)
	pubsub := h.redisClient.Subscribe(r.Context(), channel)
	defer pubsub.Close()

	ch := pubsub.Channel()

	conn.SetReadDeadline(time.Now().Add(time.Hour))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(time.Hour))
		return nil
	})

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	statusTicker := time.NewTicker(5 * time.Second)
	defer statusTicker.Stop()

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				return
			}
		}
	}()

	for {
		select {
		case msg := <-ch:
			var logMsg infraredis.LogMessage
			if err := json.Unmarshal([]byte(msg.Payload), &logMsg); err != nil {
				continue
			}

			if err := conn.WriteJSON(logMsg); err != nil {
				return
			}

		case <-ticker.C:
			if err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(10*time.Second)); err != nil {
				return
			}

		case <-statusTicker.C:
			deployment, err := h.deploymentUC.GetDeployment(r.Context(), userID, deploymentID)
			if err != nil {
				return
			}

			if isTerminalStatus(deployment.Status) {
				finalMsg := infraredis.LogMessage{
					DeploymentID: deploymentID,
					Line:         fmt.Sprintf("[DEPLOYMENT %s]", strings.ToUpper(string(deployment.Status))),
					Timestamp:    time.Now().Unix(),
				}
				conn.WriteJSON(finalMsg)
				return
			}

		case <-done:
			return

		case <-r.Context().Done():
			return
		}
	}
}

func isTerminalStatus(status domain.DeploymentStatus) bool {
	return status == domain.DeploymentStatusLive ||
		status == domain.DeploymentStatusFailed ||
		status == domain.DeploymentStatusStopped
}
