package twitch_api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ylubyanoy/go_web_server/internal/api"
)

const clientID = "uqpc0satolohmpkplj0q0zgon883qx"
const getStreamerInfoURL = "https://api.twitch.tv/kraken/users?login="
const getStreamStatusURL = "https://api.twitch.tv/kraken/streams/"

// TwitchClient is a Twitch API client
type TwitchClient struct {
	url string
}

// NewTwitchClient creates a new joke client
func NewTwitchClient() *TwitchClient {
	return &TwitchClient{}
}

// GetStreamerInfo get Streamer information
func (tc *TwitchClient) GetStreamerInfo(streamerName, clientID string) (*api.TwitchStreamerInfo, error) {

	urlPath := getStreamerInfoURL + streamerName

	timeout := time.Duration(5 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}

	request, err := http.NewRequest("GET", urlPath, nil)
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

	request, err := http.NewRequest("GET", urlPath, nil)
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
