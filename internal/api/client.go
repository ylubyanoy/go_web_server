package api

// Client interacts with Twitch API
type Client interface {
	GetStreamStatus(ID, clientID string) (*TwitchStreamStatus, error)
	GetStreamerInfo(streamerName, clientID string) (*TwitchStreamerInfo, error)
}
