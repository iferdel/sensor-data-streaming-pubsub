package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/spf13/viper"
)

type LoginRequest struct {
	// otp stands for one-time password
	Otp string `json:"otp"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func FetchAccessToken() (*LoginResponse, error) {
	api_url := viper.GetString("api_url")
	client := &http.Client{}
	r, err := http.NewRequest("POST", api_url+"/v1/auth/refresh", bytes.NewBuffer([]byte{}))
	r.Header.Add("X-Refresh-Token", viper.GetString("refresh_token"))
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.New("invalid refresh token")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var creds LoginResponse
	err = json.Unmarshal(body, &creds)
	return &creds, err
}
