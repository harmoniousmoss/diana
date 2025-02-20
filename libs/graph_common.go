package libs

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"myfiberproject/config"
)

// TokenResponse represents the structure of the response when fetching an access token.
type TokenResponse struct {
	AccessToken string `json:"access_token"`
}

// Email represents an individual email message.
type Email struct {
	ID               string `json:"id"`
	ReceivedDateTime string `json:"receivedDateTime"`
	Subject          string `json:"subject"`
	From             struct {
		EmailAddress struct {
			Name    string `json:"name"`
			Address string `json:"address"`
		} `json:"emailAddress"`
	} `json:"from"`
	Body struct {
		ContentType string `json:"contentType"`
		Content     string `json:"content"`
	} `json:"body"`
}

// EmailResponse represents the structure of the email response from Microsoft Graph API.
type EmailResponse struct {
	Value    []Email `json:"value"`
	NextLink string  `json:"@odata.nextLink"`
}

// GetMicrosoftAccessToken retrieves the access token for Microsoft Graph API.
func GetMicrosoftAccessToken() (string, error) {
	tenantID := config.GetEnv("TENANT_ID", "")
	clientID := config.GetEnv("CLIENT_ID", "")
	clientSecret := config.GetEnv("CLIENT_SECRET", "")
	tokenURL := fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", tenantID)

	data := "grant_type=client_credentials" +
		"&client_id=" + clientID +
		"&client_secret=" + clientSecret +
		"&scope=https://graph.microsoft.com/.default"

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to get access token: %s", body)
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", err
	}

	return tokenResp.AccessToken, nil
}
