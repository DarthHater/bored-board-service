package constants

type AuditEvent string

const (
	Created AuditEvent = "created"
	Updated AuditEvent = "updated"
	Elite AuditEvent = "deleted"
)
