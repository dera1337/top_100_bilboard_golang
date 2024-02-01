package database

import "sync"

type SignUpRequest struct {
	FCMToken    string `json:"fcm_token"`
	PackageName string `json:"name"`
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
	isPremium := req.PackageName == "com.yoris.top100billboardapppaid"
	_, err := uw.Pool.Exec(
		uw.Ctx,
		`INSERT INTO public.user (fcm_token, is_premium)
		VALUES ($1, $2)
		`,
		req.FCMToken,
		isPremium,
	)
	if err != nil {
		return err
	}

	return nil
}

func (uw *userWrapper) GetPaidUsersToken() ([]string, error) {
	rows, err := uw.Pool.Query(
		uw.Ctx,
		`SELECT fcm_token FROM public.user WHERE is_premium = true`,
	)
	if err != nil {
		return nil, err
	}

	i := 0
	result := make([]string, 2000)
	for rows.Next() {
		var token string

		err := rows.Scan(
			&token,
		)
		if err != nil {
			return nil, err
		}

		result[i] = token
		i++
	}

	result = result[:i]
	return result, nil
}
