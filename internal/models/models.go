package models

// StreamerNickName is streamer username
type StreamerNickName struct {
	Username string `json:"username"`
}

// Streamers is list of streamers
type Streamers struct {
	Streamer []StreamerNickName `json:"users"`
}

// StreamerInfo is struct of common stream information
type StreamerInfo struct {
	ChannelName  string `json:"channel_name"`
	Game         string `json:"game"`
	Viewers      int    `json:"viewers"`
	StatusStream string `json:"status_stream"`
	Thumbnail    string `json:"thumbnail"`
}
