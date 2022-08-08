package models

type AuthRequest struct {
	Auth         bool   `swaggerignore:"true"`
	UpdateTokens bool   `swaggerignore:"true"`
	Login        string `json:"login"`
	Password     string `json:"password"`
}

func (ar *AuthRequest) IsValid() (r bool) {
	r = false
	if len(ar.Login) > 0 && len(ar.Password) > 0 {
		r = true
	}

	return
}

type AuthTokenRequest struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}
