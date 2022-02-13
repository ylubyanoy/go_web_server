package api

import "time"

type TwitchAccessToken struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

type TwitchUsers struct {
	DisplayName string    `json:"display_name"`
	ID          string    `json:"_id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Bio         string    `json:"bio"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Logo        string    `json:"logo"`
}

type TwitchStreamerInfo struct {
	Total int           `json:"_total"`
	Users []TwitchUsers `json:"users"`
}

type StreamPreview struct {
	Small    string `json:"small"`
	Medium   string `json:"medium"`
	Large    string `json:"large"`
	Template string `json:"template"`
}

type StreamChannel struct {
	Mature                       bool      `json:"mature"`
	Status                       string    `json:"status"`
	BroadcasterLanguage          string    `json:"broadcaster_language"`
	BroadcasterSoftware          string    `json:"broadcaster_software"`
	DisplayName                  string    `json:"display_name"`
	Game                         string    `json:"game"`
	Language                     string    `json:"language"`
	ID                           int       `json:"_id"`
	Name                         string    `json:"name"`
	CreatedAt                    time.Time `json:"created_at"`
	UpdatedAt                    time.Time `json:"updated_at"`
	Partner                      bool      `json:"partner"`
	Logo                         string    `json:"logo"`
	VideoBanner                  string    `json:"video_banner"`
	ProfileBanner                string    `json:"profile_banner"`
	ProfileBannerBackgroundColor string    `json:"profile_banner_background_color"`
	URL                          string    `json:"url"`
	Views                        int       `json:"views"`
	Followers                    int       `json:"followers"`
	BroadcasterType              string    `json:"broadcaster_type"`
	Description                  string    `json:"description"`
	PrivateVideo                 bool      `json:"private_video"`
	PrivacyOptionsEnabled        bool      `json:"privacy_options_enabled"`
}

type TwitchStream struct {
	ID                int64         `json:"_id"`
	Game              string        `json:"game"`
	BroadcastPlatform string        `json:"broadcast_platform"`
	CommunityID       string        `json:"community_id"`
	CommunityIds      []interface{} `json:"community_ids"`
	Viewers           int           `json:"viewers"`
	VideoHeight       int           `json:"video_height"`
	AverageFps        int           `json:"average_fps"`
	Delay             int           `json:"delay"`
	CreatedAt         time.Time     `json:"created_at"`
	IsPlaylist        bool          `json:"is_playlist"`
	StreamType        string        `json:"stream_type"`
	Preview           StreamPreview `json:"preview"`
	Channel           StreamChannel `json:"channel"`
}

type TwitchStreamStatus struct {
	Stream TwitchStream `json:"stream"`
}
