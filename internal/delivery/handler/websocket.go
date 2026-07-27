package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	goredis "github.com/redis/go-redis/v9"

	"github.com/Ankesh2004/pontoon/internal/delivery/middleware"
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
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	deploymentID := r.URL.Query().Get("deployment_id")
	if deploymentID == "" {
		http.Error(w, "deployment_id query parameter required", http.StatusBadRequest)
		return
	}

	deployment, err := h.deploymentUC.GetDeployment(r.Context(), userID, deploymentID)
	if err != nil {
		if err == domain.ErrNotFound {
			http.Error(w, "deployment not found", http.StatusNotFound)
			return
		}
		if err == domain.ErrForbidden {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		http.Error(w, "failed to get deployment", http.StatusInternalServerError)
		return
	}

	_ = deployment // Deployment validated, now stream logs

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

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

		case <-done:
			return

		case <-r.Context().Done():
			return
		}
	}
}

func (h *WebSocketHandler) Close() error {
	return h.redisClient.Close()
}
