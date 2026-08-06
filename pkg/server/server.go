package server

import (
	"log"
	"net/http"

	"github.com/iavianm/raindrop_go_bot/pkg/raindrop"
	"github.com/iavianm/raindrop_go_bot/pkg/storage"
)

type Server struct {
	raindrop *raindrop.Client
	storage  *storage.Storage
}

func NewServer(rd *raindrop.Client, db *storage.Storage) *Server {
	return &Server{
		raindrop: rd,
		storage:  db,
	}
}

func (s *Server) handleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	if code == "" {
		log.Printf("callback without code: %s", r.URL.RawQuery)
		http.Error(w, "missing code", http.StatusBadRequest)
		return
	}

	chatID, err := s.storage.PopState(state)
	if err != nil {
		log.Printf("callback with bad state %q: %v", state, err)
		http.Error(w, "invalid state", http.StatusInternalServerError)
		return
	}
	if chatID == 0 {
		log.Printf("callback with unknown state: %q", state)
		http.Error(w, "unknown state", http.StatusBadRequest)
		return
	}

	token, err := s.raindrop.Authorize(code)
	if err != nil {
		log.Printf("authorize failed (chat_id=%d): %v", chatID, err)
		http.Error(w, "authorization failed", http.StatusInternalServerError)
		return
	}

	if err := s.storage.SaveTokens(chatID, token); err != nil {
		log.Printf("save tokens failed (chat_id=%d): %v", chatID, err)
		http.Error(w, "failed to store tokens", http.StatusInternalServerError)
		return
	}

	log.Printf("authorized chat_id=%d expires_in=%d", chatID, token.ExpiresIn)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("Authorization successful. You can return to the bot.")); err != nil {
		log.Println(err)
	}
}

func (s *Server) Start() {
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", s.handleCallback)

	log.Println("Server starting on port 8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
