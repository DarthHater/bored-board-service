package constants

type AuditEvent string

const (
	Creat AuditEvent = "created"
	Mod AuditEvent = "updated"
	Elite AuditEvent = "deleted"
)
