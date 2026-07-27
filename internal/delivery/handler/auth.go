package handler

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"
	"time"

	"github.com/Ankesh2004/pontoon/internal/usecase"
)

type AuthHandler struct {
	authUC *usecase.AuthUseCase
	states *stateStore
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

func NewAuthHandler(authUC *usecase.AuthUseCase) *AuthHandler {
	return &AuthHandler{
		authUC: authUC,
		states: newStateStore(),
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
		Name:     "pontoon_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"token":"` + token + `"}`))
}
