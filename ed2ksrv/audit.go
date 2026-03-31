package ed2ksrv

import (
	"sort"
	"time"
)

const maxAuditEntries = 200

// AuditEntry records an admin-visible management event.
type AuditEntry struct {
	Time       time.Time `json:"time"`
	Action     string    `json:"action"`
	Resource   string    `json:"resource"`
	ResourceID string    `json:"resource_id,omitempty"`
	RemoteAddr string    `json:"remote_addr,omitempty"`
	Status     string    `json:"status"`
	Detail     string    `json:"detail,omitempty"`
}

// AuditSnapshot returns a reverse-chronological copy of the audit log.
func (s *Server) AuditSnapshot() []AuditEntry {
	s.mu.RLock()
	entries := append([]AuditEntry(nil), s.auditLog...)
	s.mu.RUnlock()
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Time.After(entries[j].Time)
	})
	return entries
}

// appendAudit records one management event in memory.
func (s *Server) appendAudit(entry AuditEntry) {
	if entry.Time.IsZero() {
		entry.Time = time.Now()
	}
	if entry.Status == "" {
		entry.Status = "ok"
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.auditLog = append(s.auditLog, entry)
	if len(s.auditLog) > maxAuditEntries {
		s.auditLog = append([]AuditEntry(nil), s.auditLog[len(s.auditLog)-maxAuditEntries:]...)
	}
}
