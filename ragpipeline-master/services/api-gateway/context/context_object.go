// Package context provides a unified context object shared across all services via Redis.
package context

import "time"

// ContextObject represents the unified session context shared by all agents.
// It is stored in Redis and updated by various services during a session.
type ContextObject struct {
	SessionID      string           `json:"session_id"`
	UserID         string           `json:"user_id"`
	Board          string           `json:"board"`            // CBSE, ICSE, State
	Grade          string           `json:"grade"`            // 6-12
	Language       string           `json:"language"`         // English, Hindi
	CurrentChapter string           `json:"current_chapter"`
	CurrentSubject string           `json:"current_subject"`
	PracticeErrors []PracticeError  `json:"practice_errors"`
	PracticeScores map[string]float64 `json:"practice_scores"`
	SessionTurns   int              `json:"session_turns"`
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`
}

// PracticeError tracks a recurring mistake made by the student.
type PracticeError struct {
	Topic    string `json:"topic"`
	ErrorMsg string `json:"error_msg"`
	Count    int    `json:"count"`
}

// NewContextObject creates a new context object with the given session and user IDs.
func NewContextObject(sessionID, userID string) *ContextObject {
	now := time.Now()
	return &ContextObject{
		SessionID:      sessionID,
		UserID:         userID,
		PracticeScores: make(map[string]float64),
		PracticeErrors: make([]PracticeError, 0),
		SessionTurns:   0,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

// AddPracticeError records a practice error, incrementing the count if the same
// topic and error message already exist.
func (c *ContextObject) AddPracticeError(topic, errMsg string) {
	for i := range c.PracticeErrors {
		if c.PracticeErrors[i].Topic == topic && c.PracticeErrors[i].ErrorMsg == errMsg {
			c.PracticeErrors[i].Count++
			return
		}
	}
	c.PracticeErrors = append(c.PracticeErrors, PracticeError{
		Topic:    topic,
		ErrorMsg: errMsg,
		Count:    1,
	})
}

// RecordPracticeScore records a practice score for a given topic.
func (c *ContextObject) RecordPracticeScore(topic string, score float64) {
	if c.PracticeScores == nil {
		c.PracticeScores = make(map[string]float64)
	}
	c.PracticeScores[topic] = score
}

// IncrementTurns increments the session turn counter and updates the timestamp.
func (c *ContextObject) IncrementTurns() {
	c.SessionTurns++
	c.UpdatedAt = time.Now()
}

// UpdateBoard sets the curriculum board and updates the timestamp.
func (c *ContextObject) UpdateBoard(board string) {
	c.Board = board
	c.UpdatedAt = time.Now()
}

// UpdateGrade sets the grade and updates the timestamp.
func (c *ContextObject) UpdateGrade(grade string) {
	c.Grade = grade
	c.UpdatedAt = time.Now()
}

// UpdateLanguage sets the language and updates the timestamp.
func (c *ContextObject) UpdateLanguage(language string) {
	c.Language = language
	c.UpdatedAt = time.Now()
}

// UpdateCurrentChapter sets the current chapter and updates the timestamp.
func (c *ContextObject) UpdateCurrentChapter(chapter string) {
	c.CurrentChapter = chapter
	c.UpdatedAt = time.Now()
}

// UpdateCurrentSubject sets the current subject and updates the timestamp.
func (c *ContextObject) UpdateCurrentSubject(subject string) {
	c.CurrentSubject = subject
	c.UpdatedAt = time.Now()
}
