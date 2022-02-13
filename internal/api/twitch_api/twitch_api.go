package twitch_api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ylubyanoy/go_web_server/internal/api"
)

const (
	clientID           = "uqpc0satolohmpkplj0q0zgon883qx"
	clientSecret       = "tkdw28jvktekj56gw5k4m2qrwcdvcc"
	getStreamerInfoURL = "https://api.twitch.tv/kraken/users?login="
	getStreamStatusURL = "https://api.twitch.tv/kraken/streams/"
)

// TwitchClient is a Twitch API client
type TwitchClient struct {
	url string
}

// NewTwitchClient creates a new joke client
func NewTwitchClient() *TwitchClient {
	return &TwitchClient{}
}

// GetAccessToken get Streamer information
func (tc *TwitchClient) GetAccessToken(clientID, clientSecret string) (string, error) {

	urlPath := "https://id.twitch.tv/oauth2/token?grant_type=client_credentials"
	urlRedirect := url.QueryEscape("https://8a88-46-72-17-208.ngrok.io")

	payload := strings.NewReader(fmt.Sprintf("client_id=%s&client_secret=%s&redirect_uri=%s&code=%s", clientID, clientSecret, urlRedirect, clientSecret))

	timeout := time.Duration(5 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}

	request, _ := http.NewRequest("POST", urlPath, payload)
	request.Header.Set("Client-ID", clientID)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Cookie", "twitch.lohp.countryCode=RU")

	resp, err := client.Do(request)
	if err != nil {
		return "", err
	} else if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request status: %s", http.StatusText(resp.StatusCode))
	}
	defer resp.Body.Close()

	var tat api.TwitchAccessToken
	err = json.NewDecoder(resp.Body).Decode(&tat)
	if err != nil {
		return "", err
	}

	return tat.AccessToken, nil

}

// GetStreamerInfo get Streamer information
func (tc *TwitchClient) GetStreamerInfo(streamerName, clientID string) (*api.TwitchStreamerInfo, error) {

	urlPath := getStreamerInfoURL + streamerName

	timeout := time.Duration(5 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}

	request, _ := http.NewRequest("GET", urlPath, nil)
	request.Header.Set("Accept", "application/vnd.twitchtv.v5+json")
	request.Header.Set("Client-ID", clientID)

	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	} else if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request status: %s", http.StatusText(resp.StatusCode))
	}
	defer resp.Body.Close()

	var tsi api.TwitchStreamerInfo
	err = json.NewDecoder(resp.Body).Decode(&tsi)
	if err != nil {
		return nil, err
	}

	return &tsi, nil
}

// GetStreamStatus get Stream status for streamer
func (tc *TwitchClient) GetStreamStatus(ID, clientID string) (*api.TwitchStreamStatus, error) {

	urlPath := getStreamStatusURL + ID

	timeout := time.Duration(5 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}

	request, _ := http.NewRequest("GET", urlPath, nil)
	request.Header.Set("Accept", "application/vnd.twitchtv.v5+json")
	request.Header.Set("Client-ID", clientID)

	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	} else if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request status: %s", http.StatusText(resp.StatusCode))
	}
	defer resp.Body.Close()

	var tss api.TwitchStreamStatus
	err = json.NewDecoder(resp.Body).Decode(&tss)
	if err != nil {
		return nil, err
	}

	return &tss, nil
}
