package handler

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"

	"github.com/Ankesh2004/pontoon/internal/delivery/middleware"
	"github.com/Ankesh2004/pontoon/internal/usecase"
)

type AuthHandler struct {
	authUC      *usecase.AuthUseCase
	userUC      *usecase.UserUseCase
	states      *stateStore
	redisClient *goredis.Client
}

type stateStore struct {
	mu     sync.RWMutex
	states map[string]time.Time
}

func newStateStore() *stateStore {
	return &stateStore{
		states: make(map[string]time.Time),
	}
}

func (s *stateStore) Generate() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	state := hex.EncodeToString(bytes)

	s.mu.Lock()
	s.states[state] = time.Now().Add(10 * time.Minute)
	s.mu.Unlock()

	go func() {
		time.Sleep(10 * time.Minute)
		s.mu.Lock()
		delete(s.states, state)
		s.mu.Unlock()
	}()

	return state, nil
}

func (s *stateStore) Validate(state string) bool {
	s.mu.RLock()
	expiry, exists := s.states[state]
	s.mu.RUnlock()

	if !exists {
		return false
	}

	if time.Now().After(expiry) {
		s.mu.Lock()
		delete(s.states, state)
		s.mu.Unlock()
		return false
	}

	s.mu.Lock()
	delete(s.states, state)
	s.mu.Unlock()

	return true
}

func NewAuthHandler(authUC *usecase.AuthUseCase, userUC *usecase.UserUseCase, redisClient *goredis.Client) *AuthHandler {
	return &AuthHandler{
		authUC:      authUC,
		userUC:      userUC,
		states:      newStateStore(),
		redisClient: redisClient,
	}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	state, err := h.states.Generate()
	if err != nil {
		http.Error(w, "failed to generate state", http.StatusInternalServerError)
		return
	}

	authURL := h.authUC.GetAuthURL(state)
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

func (h *AuthHandler) Callback(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")
	if state == "" {
		http.Error(w, "missing state parameter", http.StatusBadRequest)
		return
	}

	if !h.states.Validate(state) {
		http.Error(w, "invalid or expired state", http.StatusBadRequest)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "missing code parameter", http.StatusBadRequest)
		return
	}

	token, err := h.authUC.HandleCallback(r.Context(), code)
	if err != nil {
		http.Error(w, "authentication failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     middleware.SessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400,
	})

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.userUC.GetUserByID(r.Context(), userID)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":       user.ID,
		"username": user.GitHubUsername,
		"avatar":   user.AvatarURL,
		"email":    user.Email,
	})
}

func (h *AuthHandler) WSTicket(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	ticket := uuid.New().String()
	key := "ws_ticket:" + ticket

	err := h.redisClient.Set(r.Context(), key, userID, 30*time.Second).Err()
	if err != nil {
		http.Error(w, "failed to create ticket", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ticket":     ticket,
		"expires_in": 30,
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     middleware.SessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})

	w.WriteHeader(http.StatusOK)
}
