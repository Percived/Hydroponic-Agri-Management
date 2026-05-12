package auth

type LoginRequest struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Password string `json:"password" binding:"required,min=6,max=64"`
}

type CreateUserRequest struct {
	Username string   `json:"username" binding:"required,min=3,max=32"`
	Password string   `json:"password" binding:"required,min=6,max=64"`
	Nickname string   `json:"nickname" binding:"max=64"`
	Phone    string   `json:"phone" binding:"max=32"`
	Email    string   `json:"email" binding:"max=64"`
	Roles    []string `json:"roles" binding:"required,min=1"`
}

type UpdateUserRequest struct {
	Nickname *string   `json:"nickname" binding:"omitempty,max=64"`
	Phone    *string   `json:"phone" binding:"omitempty,max=32"`
	Email    *string   `json:"email" binding:"omitempty,max=64"`
	Roles    *[]string `json:"roles"`
}

type UpdateUserStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=ENABLED DISABLED"`
}

type CreateRoleRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=32"`
	Description string `json:"description" binding:"max=64"`
}

type UpdateRoleRequest struct {
	Description string `json:"description" binding:"max=64"`
}
