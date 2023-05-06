package users

type User struct {
	ID       int64  `json:"id"`
	UserName string `json:"username"`
	UserSlug string `json:"userslug"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

var (
	CreateUserQuery = `INSERT INTO users ( username, userslug, email, password, joined) VALUES ( $1, $2, $3, $4, $5)`
	GetUserQuery    = `SELECT id, userslug, password FROM users WHERE email = $1`
	IsUniqueSlug    = `SELECT count(id) FROM users WHERE userslug = $1`
	IsUniqueEmail   = `SELECT count(id) FROM users WHERE email = $1`
)
