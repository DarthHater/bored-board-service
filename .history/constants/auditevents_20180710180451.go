package constants

type AuditEvent string

const (
	Created AuditEvent = "created"
	Updated AuditEvent = "updated"
	Deleted AuditEvent = "deleted"
)
