package models

type AuthResponse struct {
	Status       string `json:"status"`
	Login        string `json:"login,omitempty"`
	AccessToken  string `json:"accessToken,omitempty"`
	RefreshToken string `json:"refreshToken,omitempty"`
}
