package constants

// Role determines what user role someone has.
type Role int

// Constants for user roles.
const (
	Admin Role = 0
	Mod Role = 1
	Elite Role = 2
	User Role = 3
	Muted Role = 4
	Banned Role = 5
)

// Constants for JWT claim values.
const (
	Expires = "exp"
	IssuedAt = "iat"
	UserName = "user"
	UserID = "id"
	UserRole = "role"
)
