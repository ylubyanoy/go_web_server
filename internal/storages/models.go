package storages

// StreamerNickName is streamer username
type StreamerNickName struct {
	Username string
}

// Streamers is list of streamers
type Streamers struct {
	Streamer []StreamerNickName
}

// StreamerInfo is struct of common stream information
type StreamerInfo struct {
	ChannelName  string
	Game         string
	Viewers      int
	StatusStream string
	Thumbnail    string
}
