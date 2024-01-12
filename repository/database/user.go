package database

import "sync"

type SignUpRequest struct {
	FCMToken  string `json:"fcm_token"`
	IsPremium bool   `json:"is_premium"`
}

type SignUpResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type RefreshTokenResponse struct {
	AccessToken string `json:"access_token"`
}

type userWrapper struct {
	*dbConn
	sync.RWMutex
}

var (
	UserWrapper userWrapper
)

func (uw *userWrapper) InsertUser(req *SignUpRequest) error {
	_, err := uw.Pool.Exec(
		uw.Ctx,
		`INSERT INTO public.user (fcm_token, is_premium)
		VALUES ($1, $2)
		`,
		req.FCMToken,
		req.IsPremium,
	)
	if err != nil {
		return err
	}

	return nil
}
