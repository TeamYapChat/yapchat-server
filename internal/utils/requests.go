package utils

type UpdateUserRequest struct {
	Username string `json:"username,omitempty"  example:"john_doe"`
	ImageURL string `json:"image_url,omitempty" example:"https://example.com/profile_picture.jpg"`
}
