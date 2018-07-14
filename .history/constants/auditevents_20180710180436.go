package constants

type AuditEvent string

const (
	Created AuditEvent = "created"
	Mod AuditEvent = "updated"
	Elite AuditEvent = 2
	User AuditEvent = 3
	Muted AuditEvent = 4
	Banned AuditEvent = 5
)
