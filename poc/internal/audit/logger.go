package audit

import (
	"sync"
	"time"
)

type LogEntry struct {
	ID          string    `json:"id"`
	Timestamp   time.Time `json:"timestamp"`
	Type        string    `json:"type"` // "query", "execution", "error"
	User        string    `json:"user"`
	Query       string    `json:"query,omitempty"`
	Command     string    `json:"command"`
	Namespace   string    `json:"namespace,omitempty"`
	Safety      string    `json:"safety,omitempty"`
	Success     bool      `json:"success"`
	Error       string    `json:"error,omitempty"`
	Duration    int64     `json:"duration_ms,omitempty"`
	RemoteAddr  string    `json:"remote_addr,omitempty"`
}

type Logger struct {
	entries []LogEntry
	mu      sync.RWMutex
}

func NewLogger() *Logger {
	return &Logger{
		entries: make([]LogEntry, 0),
	}
}

func (l *Logger) LogQuery(query, command, safety, user string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	entry := LogEntry{
		ID:        generateID(),
		Timestamp: time.Now(),
		Type:      "query",
		User:      user,
		Query:     query,
		Command:   command,
		Safety:    safety,
		Success:   true,
	}

	l.entries = append(l.entries, entry)
}

func (l *Logger) LogExecution(command, namespace string, success bool, user string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	entry := LogEntry{
		ID:        generateID(),
		Timestamp: time.Now(),
		Type:      "execution",
		User:      user,
		Command:   command,
		Namespace: namespace,
		Success:   success,
	}

	l.entries = append(l.entries, entry)
}

func (l *Logger) LogError(command, namespace, errorMsg, user string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	entry := LogEntry{
		ID:        generateID(),
		Timestamp: time.Now(),
		Type:      "error",
		User:      user,
		Command:   command,
		Namespace: namespace,
		Success:   false,
		Error:     errorMsg,
	}

	l.entries = append(l.entries, entry)
}

func (l *Logger) GetRecentLogs(limit int) []LogEntry {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if len(l.entries) == 0 {
		return []LogEntry{}
	}

	start := len(l.entries) - limit
	if start < 0 {
		start = 0
	}

	// Return a copy to avoid race conditions
	result := make([]LogEntry, len(l.entries[start:]))
	copy(result, l.entries[start:])
	
	return result
}

func (l *Logger) GetLogsByUser(user string, limit int) []LogEntry {
	l.mu.RLock()
	defer l.mu.RUnlock()

	var userLogs []LogEntry
	for i := len(l.entries) - 1; i >= 0 && len(userLogs) < limit; i-- {
		if l.entries[i].User == user {
			userLogs = append(userLogs, l.entries[i])
		}
	}

	return userLogs
}

func generateID() string {
	// Simple ID generation for PoC - use proper UUID in production
	return time.Now().Format("20060102-150405") + "-" + generateRandomString(6)
}

func generateRandomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, n)
	for i := range result {
		result[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(result)
}