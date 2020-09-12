package models

// StreamerNickName is streamer username
type StreamerNickName struct {
	Username string `json:"username"`
}

// Streamers is list of streamers
type Streamers struct {
	Streamer []StreamerNickName `json:"users"`
}

// var streamersList = `{
// 	"users": [
// 	  {"username": "thaina_"},
// 	  {"username": "blabalbee"},
// 	  {"username": "Smorodinova"},
// 	  {"username": "CekLena"},
// 	  {"username": "JowyBear"},
// 	  {"username": "pimpka74"},
// 	  {"username": "icytoxictv"},
// 	  {"username": "ustepuka"},
// 	  {"username": "AlenochkaBT"},
// 	  {"username": "ViktoriiShka"},
// 	  {"username": "irenchik"},
// 	  {"username": "lola_grrr"},
// 	  {"username": "Sensoria"},
// 	  {"username": "aisumaisu"},
// 	  {"username": "PANGCHOM"},
// 	  {"username": "Danucd"}
// 	]
// }`

// StreamerInfo is struct of common stream information
type StreamerInfo struct {
	ChannelName  string `json:"channel_name"`
	Game         string `json:"game"`
	Viewers      int    `json:"viewers"`
	StatusStream string `json:"status_stream"`
	Thumbnail    string `json:"thumbnail"`
}
