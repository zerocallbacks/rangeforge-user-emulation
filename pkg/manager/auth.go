package manager

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// AdminAccount stores administrative credentials and authentication metadata.
type AdminAccount struct {
	Username     string `json:"username"`
	PasswordHash string `json:"password_hash"`
	Salt         string `json:"salt"`
	UpdatedAt    string `json:"updated_at"`
}

// SessionInfo holds active session data for an authenticated admin.
type SessionInfo struct {
	Username  string
	ExpiresAt time.Time
}

// SessionStore manages active authenticated sessions.
type SessionStore struct {
	mu       sync.RWMutex
	sessions map[string]SessionInfo
}

// NewSessionStore initializes the session store.
func NewSessionStore() *SessionStore {
	return &SessionStore{
		sessions: make(map[string]SessionInfo),
	}
}

// CreateSession generates a cryptographically secure token for an authenticated user.
func (s *SessionStore) CreateSession(username string, duration time.Duration) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	b := make([]byte, 24)
	_, _ = rand.Read(b)
	token := hex.EncodeToString(b)

	s.sessions[token] = SessionInfo{
		Username:  username,
		ExpiresAt: time.Now().UTC().Add(duration),
	}
	return token
}

// ValidateSession verifies if a session token is active and valid.
func (s *SessionStore) ValidateSession(token string) (string, bool) {
	if token == "" {
		return "", false
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	sess, ok := s.sessions[token]
	if !ok {
		return "", false
	}
	if time.Now().UTC().After(sess.ExpiresAt) {
		return "", false
	}
	return sess.Username, true
}

// InvalidateSession removes a session token.
func (s *SessionStore) InvalidateSession(token string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, token)
}

// HashPassword generates a salted SHA-256 hash.
func HashPassword(password, salt string) string {
	h := sha256.New()
	h.Write([]byte(salt + ":" + password))
	return hex.EncodeToString(h.Sum(nil))
}

// GenerateSalt creates a random 16-byte hex string.
func GenerateSalt() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// VerifyPassword checks if a plain password matches the stored salted hash.
func VerifyPassword(password, salt, storedHash string) bool {
	computed := HashPassword(password, salt)
	return subtle.ConstantTimeCompare([]byte(computed), []byte(storedHash)) == 1
}

// initAdminAuth loads or initializes the admin account credentials.
func (s *Server) initAdminAuth() {
	s.adminMu.Lock()
	defer s.adminMu.Unlock()

	authFile := filepath.Join("configs", "admin_auth.json")
	if s.config.ProfilesDir != "" {
		cfgDir := filepath.Dir(s.config.ProfilesDir)
		if strings.Contains(cfgDir, "test") {
			authFile = filepath.Join(cfgDir, "admin_auth.json")
		}
	}
	s.authFile = authFile

	data, err := os.ReadFile(authFile)
	if err == nil && len(data) > 0 {
		var acc AdminAccount
		if err := json.Unmarshal(data, &acc); err == nil && acc.Username != "" && acc.PasswordHash != "" {
			s.adminAuth = acc
			log.Printf("[MANAGER] Loaded Admin Account credentials for user '%s'", acc.Username)
			return
		}
	}

	// Default fallback admin account
	salt := GenerateSalt()
	defaultUser := "admin"
	defaultPass := "rangeforge"
	s.adminAuth = AdminAccount{
		Username:     defaultUser,
		Salt:         salt,
		PasswordHash: HashPassword(defaultPass, salt),
		UpdatedAt:    time.Now().UTC().Format(time.RFC3339),
	}

	_ = os.MkdirAll(filepath.Dir(authFile), 0755)
	if outBytes, err := json.MarshalIndent(s.adminAuth, "", "  "); err == nil {
		_ = os.WriteFile(authFile, outBytes, 0644)
	}
	log.Printf("[MANAGER] Initialized default Admin Account ('%s') at %s", defaultUser, authFile)
}

// extractToken extracts authentication token from cookie or Authorization header.
func extractToken(r *http.Request) string {
	if c, err := r.Cookie("rf_auth_token"); err == nil && c.Value != "" {
		return c.Value
	}
	if c, err := r.Cookie("rangeforge_auth_token"); err == nil && c.Value != "" {
		return c.Value
	}
	authHdr := r.Header.Get("Authorization")
	if strings.HasPrefix(strings.ToLower(authHdr), "bearer ") {
		return strings.TrimSpace(authHdr[7:])
	}
	return ""
}

func (s *Server) handleAuthLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := decodeJSONBody(r, &req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid JSON: " + err.Error(),
		})
		return
	}

	s.adminMu.RLock()
	currentAdmin := s.adminAuth
	s.adminMu.RUnlock()

	if !strings.EqualFold(req.Username, currentAdmin.Username) || !VerifyPassword(req.Password, currentAdmin.Salt, currentAdmin.PasswordHash) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid administrative username or password",
		})
		return
	}

	token := s.sessions.CreateSession(currentAdmin.Username, 24*time.Hour)

	http.SetCookie(w, &http.Cookie{
		Name:     "rf_auth_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":   "authenticated",
		"token":    token,
		"username": currentAdmin.Username,
	})
}

func (s *Server) handleAuthLogout(w http.ResponseWriter, r *http.Request) {
	token := extractToken(r)
	if token != "" {
		s.sessions.InvalidateSession(token)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "rf_auth_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		MaxAge:   -1,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "logged_out",
	})
}

func (s *Server) handleAuthStatus(w http.ResponseWriter, r *http.Request) {
	token := extractToken(r)
	username, valid := s.sessions.ValidateSession(token)

	s.adminMu.RLock()
	configuredUser := s.adminAuth.Username
	s.adminMu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"authenticated":   valid,
		"username":        username,
		"configured_user": configuredUser,
	})
}

func (s *Server) handleAuthChangeCredentials(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		OldPassword string `json:"old_password"`
		NewUsername string `json:"new_username"`
		NewPassword string `json:"new_password"`
	}
	if err := decodeJSONBody(r, &req); err != nil {
		writeJSONError(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	s.adminMu.Lock()
	defer s.adminMu.Unlock()

	if !VerifyPassword(req.OldPassword, s.adminAuth.Salt, s.adminAuth.PasswordHash) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Current administrator password verification failed",
		})
		return
	}

	newUsername := strings.TrimSpace(req.NewUsername)
	if newUsername == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "New administrator username cannot be empty",
		})
		return
	}

	newPassword := strings.TrimSpace(req.NewPassword)
	if len(newPassword) < 6 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "New password must be at least 6 characters in length",
		})
		return
	}

	newSalt := GenerateSalt()
	newHash := HashPassword(newPassword, newSalt)

	s.adminAuth.Username = newUsername
	s.adminAuth.Salt = newSalt
	s.adminAuth.PasswordHash = newHash
	s.adminAuth.UpdatedAt = time.Now().UTC().Format(time.RFC3339)

	_ = os.MkdirAll(filepath.Dir(s.authFile), 0755)
	if outBytes, err := json.MarshalIndent(s.adminAuth, "", "  "); err == nil {
		_ = os.WriteFile(s.authFile, outBytes, 0644)
	}

	// Create fresh session for new user
	token := s.sessions.CreateSession(newUsername, 24*time.Hour)
	http.SetCookie(w, &http.Cookie{
		Name:     "rf_auth_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400,
	})

	log.Printf("[MANAGER] Admin credentials successfully changed. New username: '%s'", newUsername)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":   "success",
		"username": newUsername,
		"token":    token,
		"message":  fmt.Sprintf("Administrator account credentials updated successfully for '%s'", newUsername),
	})
}
