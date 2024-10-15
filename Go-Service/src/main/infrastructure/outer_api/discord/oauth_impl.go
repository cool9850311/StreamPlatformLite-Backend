package discord

import (
	"bytes"
	"encoding/json"
	"errors"
	"Go-Service/src/main/application/dto"
	"net/http"
	"net/url"
	"io"
	"context"
)

type DiscordOAuthImpl struct{}

func NewDiscordOAuthImpl() *DiscordOAuthImpl {
	return &DiscordOAuthImpl{}
}

func (d *DiscordOAuthImpl) GetAccessToken(ctx context.Context, clientID string, clientSecret string, code string, redirectURI string) (string, error) {
	data := url.Values{}
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)

	req, err := http.NewRequestWithContext(ctx, "POST", "https://discord.com/api/oauth2/token", bytes.NewBufferString(data.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("failed to get access token: " + string(body))
	}

	var tokenResponse struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		return "", err
	}

	return tokenResponse.AccessToken, nil
}

func (d *DiscordOAuthImpl) GetGuildMemberData(ctx context.Context, accessToken string, guildID string) (*dto.DiscordGuildMemberDTO, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://discord.com/api/users/@me/guilds/"+guildID+"/member", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to get guild member data: " + string(body))
	}

	var guildMemberData dto.DiscordGuildMemberDTO
	if err := json.Unmarshal(body, &guildMemberData); err != nil {
		return nil, err
	}

	return &guildMemberData, nil
}
