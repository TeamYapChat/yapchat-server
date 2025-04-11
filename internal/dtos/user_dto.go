package dtos

type UpdateUserRequest struct {
	Username string `json:"username,omitempty"  example:"john_doe"`
	ImageURL string `json:"image_url,omitempty" example:"https://example.com/profile_picture.jpg"`
	Status   string `json:"status,omitempty"    example:"online"`
}

type UserResponse struct {
	ID        string `json:"id"                   example:"123"`
	Username  string `json:"username"             example:"john_doe"`
	ImageURL  string `json:"image_url,omitempty"  example:"https://example.com/profile_picture.jpg"`
	IsOnline  bool   `json:"is_online"            example:"true"`
	CreatedAt string `json:"created_at,omitempty" example:"1970-01-01T00:00:00Z"`
}
