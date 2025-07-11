package discord

import (
	"Go-Service/src/main/application/dto"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
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
		return "", errors.New("failed to get access token: " + string(body) + " | request data: " + data.Encode())
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
		// If JSON unmarshaling fails, try to extract essential fields manually
		// This makes the system more resilient to Discord API changes
		var rawData map[string]interface{}
		if unmarshalErr := json.Unmarshal(body, &rawData); unmarshalErr != nil {
			return nil, errors.New("failed to parse guild member data: " + err.Error() + " | raw response: " + string(body))
		}

		// Extract essential fields that we need for the application to work
		guildMemberData = dto.DiscordGuildMemberDTO{}

		// Extract basic user data
		if user, ok := rawData["user"].(map[string]interface{}); ok {
			guildMemberData.User = dto.DiscordUserDTO{
				ID:         getStringFromInterface(user["id"]),
				Username:   getStringFromInterface(user["username"]),
				Avatar:     getStringFromInterface(user["avatar"]),
				GlobalName: getStringFromInterface(user["global_name"]),
			}
		}

		// Extract roles
		if roles, ok := rawData["roles"].([]interface{}); ok {
			guildMemberData.Roles = make([]string, len(roles))
			for i, role := range roles {
				guildMemberData.Roles[i] = getStringFromInterface(role)
			}
		}

		// Extract other essential fields
		guildMemberData.Pending = getBoolFromInterface(rawData["pending"])
		guildMemberData.Mute = getBoolFromInterface(rawData["mute"])
		guildMemberData.Deaf = getBoolFromInterface(rawData["deaf"])

		// For non-essential fields, use safe extraction
		if nick, ok := rawData["nick"]; ok && nick != nil {
			nickStr := getStringFromInterface(nick)
			guildMemberData.Nick = &nickStr
		}

		if bio, ok := rawData["bio"]; ok {
			guildMemberData.Bio = getStringFromInterface(bio)
		}
	}

	return &guildMemberData, nil
}

// Helper functions for safe type conversion
func getStringFromInterface(v interface{}) string {
	if v == nil {
		return ""
	}
	if str, ok := v.(string); ok {
		return str
	}
	return ""
}

func getBoolFromInterface(v interface{}) bool {
	if v == nil {
		return false
	}
	if b, ok := v.(bool); ok {
		return b
	}
	return false
}
