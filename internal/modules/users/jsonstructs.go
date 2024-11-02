package users

type RegisterRequest struct {
	UserName        string `json:"username"`
	Email           string `json:"email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
}

type RegisterResponse struct {
	UserName string `json:"username"`
	UserSlug string `json:"userslug"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRespone struct {
	AccessToken string `json:"access_token"`
	UserSlug    string `json:"userslug"`
}
