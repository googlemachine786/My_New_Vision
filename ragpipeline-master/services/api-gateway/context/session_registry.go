// Package context provides a unified context object shared across all services via Redis.
package context

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

const (
	// ContextKeyPrefix is the Redis key prefix for session context objects.
	ContextKeyPrefix = "rag:ctx:"

	// ContextChannelPrefix is the Redis Pub/Sub channel prefix for context updates.
	ContextChannelPrefix = "rag:ctx:pubsub:"

	// DefaultContextTTL is the default TTL for context objects in Redis.
	DefaultContextTTL = 2 * time.Hour
)

// SessionRegistry manages Redis-backed session context storage with pub/sub notifications.
type SessionRegistry struct {
	redis   *redis.Client
	ttl     time.Duration
	pubsub  map[string]*redis.PubSub
	mu      sync.RWMutex
}

// NewSessionRegistry creates a new session registry with the given Redis client.
func NewSessionRegistry(redisClient *redis.Client, opts ...RegistryOption) *SessionRegistry {
	registry := &SessionRegistry{
		redis:  redisClient,
		ttl:    DefaultContextTTL,
		pubsub: make(map[string]*redis.PubSub),
	}

	for _, opt := range opts {
		opt(registry)
	}

	return registry
}

// RegistryOption configures a SessionRegistry.
type RegistryOption func(*SessionRegistry)

// WithTTL sets the TTL for context objects.
func WithTTL(ttl time.Duration) RegistryOption {
	return func(r *SessionRegistry) {
		r.ttl = ttl
	}
}

// Get retrieves a session context by ID.
// Returns nil if the context does not exist.
func (r *SessionRegistry) Get(ctx context.Context, sessionID string) (*ContextObject, error) {
	key := contextKey(sessionID)

	data, err := r.redis.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get context for session %s: %w", sessionID, err)
	}

	var ctxObj ContextObject
	if err := json.Unmarshal(data, &ctxObj); err != nil {
		return nil, fmt.Errorf("failed to unmarshal context for session %s: %w", sessionID, err)
	}

	return &ctxObj, nil
}

// Set stores or updates a session context.
// It also publishes an update notification to subscribers.
func (r *SessionRegistry) Set(ctx context.Context, ctxObj *ContextObject) error {
	if ctxObj == nil {
		return fmt.Errorf("context object is nil")
	}

	key := contextKey(ctxObj.SessionID)
	ctxObj.UpdatedAt = time.Now()

	data, err := json.Marshal(ctxObj)
	if err != nil {
		return fmt.Errorf("failed to marshal context for session %s: %w", ctxObj.SessionID, err)
	}

	if err := r.redis.Set(ctx, key, data, r.ttl).Err(); err != nil {
		return fmt.Errorf("failed to set context for session %s: %w", ctxObj.SessionID, err)
	}

	// Publish update notification
	r.publishUpdate(ctx, ctxObj.SessionID, ctxObj)

	log.Debug().Str("session_id", ctxObj.SessionID).Msg("Session context updated")
	return nil
}

// Delete removes a session context.
func (r *SessionRegistry) Delete(ctx context.Context, sessionID string) error {
	key := contextKey(sessionID)
	if err := r.redis.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete context for session %s: %w", sessionID, err)
	}

	// Close any existing subscription
	r.unsubscribe(sessionID)

	log.Debug().Str("session_id", sessionID).Msg("Session context deleted")
	return nil
}

// Subscribe creates a subscription for context updates on a given session.
// The caller should call the returned cancel function when done.
func (r *SessionRegistry) Subscribe(ctx context.Context, sessionID string) (<-chan *ContextObject, context.CancelFunc, error) {
	channel := contextChannel(sessionID)

	r.mu.Lock()
	defer r.mu.Unlock()

	// Close existing subscription if any
	r.unsubscribe(sessionID)

	pubsub := r.redis.Subscribe(ctx, channel)
	r.pubsub[sessionID] = pubsub

	ch := make(chan *ContextObject, 10)

	subCtx, cancel := context.WithCancel(context.Background())

	go func() {
		defer close(ch)
		defer pubsub.Close()

		for {
			select {
			case <-subCtx.Done():
				return
			default:
				msg, err := pubsub.ReceiveMessage(subCtx)
				if err != nil {
					if subCtx.Err() != nil {
						return
					}
					log.Warn().Err(err).Str("session_id", sessionID).Msg("Error receiving context update")
					continue
				}

				var ctxObj ContextObject
				if err := json.Unmarshal([]byte(msg.Payload), &ctxObj); err != nil {
					log.Warn().Err(err).Str("session_id", sessionID).Msg("Failed to unmarshal context update")
					continue
				}

				select {
				case ch <- &ctxObj:
				default:
					log.Warn().Str("session_id", sessionID).Msg("Context update channel full, dropping update")
				}
			}
		}
	}()

	return ch, cancel, nil
}

// publishUpdate publishes a context update to the session's Pub/Sub channel.
func (r *SessionRegistry) publishUpdate(ctx context.Context, sessionID string, ctxObj *ContextObject) {
	channel := contextChannel(sessionID)

	data, err := json.Marshal(ctxObj)
	if err != nil {
		log.Warn().Err(err).Str("session_id", sessionID).Msg("Failed to marshal context for publishing")
		return
	}

	if err := r.redis.Publish(ctx, channel, data).Err(); err != nil {
		log.Warn().Err(err).Str("session_id", sessionID).Msg("Failed to publish context update")
	}
}

// unsubscribe closes an existing Pub/Sub subscription for a session.
func (r *SessionRegistry) unsubscribe(sessionID string) {
	if pubsub, exists := r.pubsub[sessionID]; exists {
		pubsub.Close()
		delete(r.pubsub, sessionID)
	}
}

// EnsureContext retrieves a context object, creating a new one if it doesn't exist.
func (r *SessionRegistry) EnsureContext(ctx context.Context, sessionID, userID string) (*ContextObject, error) {
	ctxObj, err := r.Get(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	if ctxObj == nil {
		ctxObj = NewContextObject(sessionID, userID)
		if err := r.Set(ctx, ctxObj); err != nil {
			return nil, fmt.Errorf("failed to create context for session %s: %w", sessionID, err)
		}
		log.Debug().Str("session_id", sessionID).Str("user_id", userID).Msg("Created new session context")
	}

	return ctxObj, nil
}

// GetRecentErrors retrieves the most recent practice errors for a session.
func (r *SessionRegistry) GetRecentErrors(ctx context.Context, sessionID string) ([]PracticeError, error) {
	ctxObj, err := r.Get(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if ctxObj == nil {
		return nil, nil
	}

	return ctxObj.PracticeErrors, nil
}

// FormatErrorsForPrompt formats practice errors into a human-readable string for prompt inclusion.
func FormatErrorsForPrompt(errors []PracticeError) string {
	if len(errors) == 0 {
		return "No recent mistakes recorded."
	}

	var sb strings.Builder
	sb.WriteString("Recent mistakes to be mindful of:\n")
	for _, e := range errors {
		if e.Count > 1 {
			sb.WriteString(fmt.Sprintf("- %s: %s (occurred %d times)\n", e.Topic, e.ErrorMsg, e.Count))
		} else {
			sb.WriteString(fmt.Sprintf("- %s: %s\n", e.Topic, e.ErrorMsg))
		}
	}
	return sb.String()
}

// FormatContextForPrompt builds a context summary string for inclusion in prompts.
func FormatContextForPrompt(ctx *ContextObject) string {
	if ctx == nil {
		return ""
	}

	var sb strings.Builder

	if ctx.Board != "" {
		sb.WriteString(fmt.Sprintf("Board: %s. ", ctx.Board))
	}
	if ctx.Grade != "" {
		sb.WriteString(fmt.Sprintf("Grade: %s. ", ctx.Grade))
	}
	if ctx.Language != "" {
		sb.WriteString(fmt.Sprintf("Language: %s. ", ctx.Language))
	}
	if ctx.CurrentSubject != "" {
		sb.WriteString(fmt.Sprintf("Subject: %s. ", ctx.CurrentSubject))
	}
	if ctx.CurrentChapter != "" {
		sb.WriteString(fmt.Sprintf("Current chapter: %s. ", ctx.CurrentChapter))
	}

	return sb.String()
}

func contextKey(sessionID string) string {
	return fmt.Sprintf("%s%s", ContextKeyPrefix, sessionID)
}

func contextChannel(sessionID string) string {
	return fmt.Sprintf("%s%s", ContextChannelPrefix, sessionID)
}
