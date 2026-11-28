package botclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/Alievs-corp/hypedomino/server/internal/account"
)

var ErrSignInRefused = errors.New("botclient: bot sign-in refused")

type Session struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	RequestKey   string `json:"request_key"`
	AccountID    string `json:"account_id"`
}

type botSignInRequest struct {
	Provider string `json:"provider"`
	BotToken string `json:"id_token"`
	DeviceID string `json:"device_id"`
}

func SignIn(ctx context.Context, client *http.Client, baseURL, botToken, device string) (Session, error) {
	body, err := json.Marshal(botSignInRequest{Provider: account.BotProvider, BotToken: botToken, DeviceID: device})
	if err != nil {
		return Session{}, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimSuffix(baseURL, "/")+"/auth/signin", bytes.NewReader(body))
	if err != nil {
		return Session{}, err
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := client.Do(request)
	if err != nil {
		return Session{}, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		var refusal struct {
			Error string `json:"error"`
		}
		json.NewDecoder(response.Body).Decode(&refusal)
		return Session{}, fmt.Errorf("%w: %d %s", ErrSignInRefused, response.StatusCode, refusal.Error)
	}
	var session Session
	if err := json.NewDecoder(response.Body).Decode(&session); err != nil {
		return Session{}, err
	}
	return session, nil
}
