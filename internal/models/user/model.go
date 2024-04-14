package user

type User struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
	IsAdmin  bool   `json:"is_admin"`
}
