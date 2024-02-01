package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
	"top_100_billboard_golang/repository/database"

	"golang.org/x/oauth2/google"
)

type NotificationType int

const (
	Register NotificationType = iota
	BillboardListUpdated
)

type fcmMessage struct {
	Message message `json:"message"`
}

type message struct {
	Token string      `json:"token"`
	Data  interface{} `json:"data"`
}

type fcmData struct {
	NotificationType string `json:"notification_type"`
	Body             string `json:"body"`
}

var (
	paidOAuthToken string
	freeOAuthToken string
)

func StartOAuthTokenGenerator() {
	err := populateToken()
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		// Expiration of OAuth token is 60 minutes, hence the 55 minutes hardcode
		case <-time.After(55 * time.Minute):
			err := populateToken()
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

func populateToken() error {
	token, err := generateOAuthToken(os.Getenv("PAID_JSON_PATH"))
	if err != nil {
		return err
	}
	paidOAuthToken = token

	token, err = generateOAuthToken(os.Getenv("FREE_JSON_PATH"))
	if err != nil {
		return err
	}
	freeOAuthToken = token

	return nil
}

func generateOAuthToken(jsonFilePath string) (string, error) {
	jsonBytes, err := os.ReadFile(jsonFilePath)
	if err != nil {
		return "", err
	}

	cred, err := google.CredentialsFromJSON(
		context.Background(),
		jsonBytes,
		"https://www.googleapis.com/auth/firebase.messaging",
	)
	if err != nil {
		return "", err
	}

	token, err := cred.TokenSource.Token()
	if err != nil {
		return "", err
	}

	return token.AccessToken, nil
}

func SendNotificationMessageToPaidUsers(artist string) error {
	tokens, err := database.UserWrapper.GetPaidUsersToken()
	if err != nil {
		return err
	}

	body := fmt.Sprintf(
		"%s is ranked 1st this week!",
		artist,
	)
	projectID := os.Getenv("PAID_PROJECT_ID")

	for _, token := range tokens {
		go func(userToken string) {
			_, err := SendNotification(
				BillboardListUpdated,
				userToken,
				body,
				projectID,
			)
			if err != nil {
				return
			}
		}(token)
	}

	return nil
}

func SendNotification(
	notificationType NotificationType,
	userToken,
	body,
	projectID string,
) (bool, error) {
	var notifType string
	switch notificationType {
	case Register:
		notifType = "register"
	case BillboardListUpdated:
		notifType = "list_updated"
	}

	fcmMessage := fcmMessage{
		Message: message{
			Token: userToken,
			Data: fcmData{
				NotificationType: notifType,
				Body:             body,
			},
		},
	}

	jsonBytes, err := json.Marshal(&fcmMessage)
	if err != nil {
		return false, err
	}

	url := fmt.Sprintf(os.Getenv("FCM_URL_FORMAT"), projectID)
	r, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return false, err
	}

	switch projectID {
	case os.Getenv("PAID_PROJECT_ID"):
		r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", paidOAuthToken))
	case os.Getenv("FREE_PROJECT_ID"):
		r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", freeOAuthToken))
	}
	r.Header.Set("Content-type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(r)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 404:
		return true, nil
	case 401:
		return false, fmt.Errorf("cannot authorize oauth credentials")
	case 400:
		if notificationType == Register {
			return false, fmt.Errorf("invalid token")
		}
	}

	return false, nil
}
