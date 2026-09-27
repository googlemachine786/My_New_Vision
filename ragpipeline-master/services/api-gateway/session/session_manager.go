package session

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

const (
	SessionKeyPrefix = "rag:session:"
	DefaultTTL       = 30 * time.Minute
	MaxTurns         = 20
)

// SessionManager manages conversation sessions in Redis
type SessionManager struct {
	redis   *redis.Client
	ttl     time.Duration
	maxTurns int
}

// NewSessionManager creates a new session manager
func NewSessionManager(redisClient *redis.Client, ttl time.Duration, maxTurns int) *SessionManager {
	if ttl <= 0 {
		ttl = DefaultTTL
	}
	if maxTurns <= 0 {
		maxTurns = MaxTurns
	}

	return &SessionManager{
		redis:    redisClient,
		ttl:      ttl,
		maxTurns: maxTurns,
	}
}

// Session represents a conversation session
type Session struct {
	ID        string           `json:"id"`
	UserID    string           `json:"user_id"`
	Turns     []ConversationTurn `json:"turns"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
}

// ConversationTurn represents a single query/response pair
type ConversationTurn struct {
	Query        string    `json:"query"`
	Response     string    `json:"response"`
	Timestamp    time.Time `json:"timestamp"`
	Grade        string    `json:"grade,omitempty"`
	Subject      string    `json:"subject,omitempty"`
}

// CreateSession creates a new session
func (sm *SessionManager) CreateSession(ctx context.Context, userID string) (*Session, error) {
	session := &Session{
		ID:        uuid.New().String(),
		UserID:    userID,
		Turns:     make([]ConversationTurn, 0),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := sm.saveSession(ctx, session); err != nil {
		return nil, err
	}

	log.Debug().Str("session_id", session.ID).Msg("Created session")
	return session, nil
}

// GetSession retrieves a session by ID
func (sm *SessionManager) GetSession(ctx context.Context, sessionID string) (*Session, error) {
	key := sessionKey(sessionID)

	data, err := sm.redis.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	return &session, nil
}

// AddTurn adds a conversation turn to a session
func (sm *SessionManager) AddTurn(ctx context.Context, sessionID string, turn ConversationTurn) error {
	session, err := sm.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}

	if session == nil {
		return fmt.Errorf("session %s not found", sessionID)
	}

	// Enforce max turns
	if len(session.Turns) >= sm.maxTurns {
		// Remove oldest turn
		session.Turns = session.Turns[1:]
	}

	session.Turns = append(session.Turns, turn)
	session.UpdatedAt = time.Now()

	return sm.saveSession(ctx, session)
}

// GetHistory retrieves recent conversation history
func (sm *SessionManager) GetHistory(ctx context.Context, sessionID string) ([]ConversationTurn, error) {
	session, err := sm.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	if session == nil {
		return nil, nil
	}

	return session.Turns, nil
}

// DeleteSession removes a session
func (sm *SessionManager) DeleteSession(ctx context.Context, sessionID string) error {
	key := sessionKey(sessionID)
	return sm.redis.Del(ctx, key).Err()
}

// EnsureSession ensures a session exists, creating one if needed
func (sm *SessionManager) EnsureSession(ctx context.Context, sessionID, userID string) (*Session, error) {
	if sessionID == "" {
		// Create new session
		session, err := sm.CreateSession(ctx, userID)
		if err != nil {
			return nil, err
		}
		return session, nil
	}

	session, err := sm.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	if session == nil {
		// Session expired or doesn't exist, create new one with same ID
		session = &Session{
			ID:        sessionID,
			UserID:    userID,
			Turns:     make([]ConversationTurn, 0),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := sm.saveSession(ctx, session); err != nil {
			return nil, err
		}
	}

	return session, nil
}

func (sm *SessionManager) saveSession(ctx context.Context, session *Session) error {
	key := sessionKey(session.ID)

	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	return sm.redis.Set(ctx, key, data, sm.ttl).Err()
}

func sessionKey(sessionID string) string {
	return fmt.Sprintf("%s%s", SessionKeyPrefix, sessionID)
}
