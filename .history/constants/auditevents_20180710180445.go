package constants

type AuditEvent string

const (
	Created AuditEvent = "created"
	Mod AuditEvent = "updated"
	Elite AuditEvent = "deleted"
)
