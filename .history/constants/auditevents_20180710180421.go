package constants

type AuditEvent string

const (
	Admin AuditEvent = 0
	Mod AuditEvent = 1
	Elite AuditEvent = 2
	User AuditEvent = 3
	Muted AuditEvent = 4
	Banned AuditEvent = 5
)
