package constants

type AuditEvent int

const (
	Admin Role = 0
	Mod Role = 1
	Elite Role = 2
	User Role = 3
	Muted Role = 4
	Banned Role = 5
)
