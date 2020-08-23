package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/gorilla/mux"
)

const servicename = "streamsinfo"

var clientID string = "uqpc0satolohmpkplj0q0zgon883qx"

var (
	redisAddr   = getEnv("REDIS_URL", "redis://user:@localhost:6379/0")
	sessManager *ConnManager
)

var streamers = `{
	"users": [
	  {"username": "thaina_"},
	  {"username": "blabalbee"},
	  {"username": "Smorodinova"},
	  {"username": "CekLena"},
	  {"username": "JowyBear"},
	  {"username": "pimpka74"},
	  {"username": "icytoxictv"},
	  {"username": "ustepuka"},
	  {"username": "AlenochkaBT"},
	  {"username": "ViktoriiShka"},
	  {"username": "irenchik"},
	  {"username": "lola_grrr"},
	  {"username": "Sensoria"},
	  {"username": "aisumaisu"},
	  {"username": "PANGCHOM"},
	  {"username": "Danucd"}
	]
}`

type ConnManager struct {
	redisConn redis.Conn
}

func (sm *ConnManager) Check(streamerName string) *StreamerInfo {
	mkey := streamerName
	data, err := redis.Bytes(sm.redisConn.Do("GET", mkey))
	if err != nil {
		log.Printf("cant get data for %s: (%s)", mkey, err)
		return nil
	}
	si := &StreamerInfo{}
	err = json.Unmarshal(data, si)
	if err != nil {
		log.Printf("cant unpack session data for %s: (%s)", mkey, err)
		return nil
	}
	return si
}

func (sm *ConnManager) Create(si *StreamerInfo) error {
	dataSerialized, _ := json.Marshal(si)
	mkey := si.ChannelName
	data, err := sm.redisConn.Do("SET", mkey, dataSerialized, "EX", 60)
	result, err := redis.String(data, err)
	if err != nil {
		return err
	}
	if result != "OK" {
		return fmt.Errorf("result not OK")
	}
	return nil
}

func NewConnManager(conn redis.Conn) *ConnManager {
	return &ConnManager{
		redisConn: conn,
	}
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

type StreamerNickName struct {
	Username string `json:"username"`
}

type Streamers struct {
	Streamer []StreamerNickName `json:"users"`
}

type StreamerInfo struct {
	ChannelName  string `json:"channel_name"`
	Viewers      string `json:"viewers"`
	StatusStream string `json:"status_stream"`
	Thumbnail    string `json:"thumbnail"`
}

func getStreamerInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	streamerName := params["streamerName"]

	// Check Redis
	si := sessManager.Check(streamerName)
	if si != nil {
		log.Println("Get from Redis")
		json.NewEncoder(w).Encode(si)
		return
	}

	timeout := time.Duration(5 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}

	//Get streamer info
	request, err := http.NewRequest("GET", "https://api.twitch.tv/kraken/users?login="+string(streamerName), nil)
	request.Header.Set("Accept", "application/vnd.twitchtv.v5+json")
	request.Header.Set("Client-ID", clientID)

	resp, err := client.Do(request)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server error"))
		log.Printf("Error: %s", err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server error"))
		log.Printf("Error: %s", err)
		return
	}

	var tsi TwitchStreamerInfo
	err = json.Unmarshal(body, &tsi)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server error"))
		log.Printf("Error: %s", err)
		return
	}

	if len(tsi.Users) == 0 {
		log.Printf("No data for User %s", streamerName)
		json.NewEncoder(w).Encode(&StreamerInfo{})
	}

	// Get stream status
	request, err = http.NewRequest("GET", "https://api.twitch.tv/kraken/streams/"+tsi.Users[0].ID, nil)
	request.Header.Set("Accept", "application/vnd.twitchtv.v5+json")
	request.Header.Set("Client-ID", clientID)

	resp, err = client.Do(request)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server error"))
		log.Printf("Error: %s", err)
		return
	}
	defer resp.Body.Close()

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server error"))
		log.Printf("Error: %s", err)
		return
	}

	var tss TwitchStreamStatus
	err = json.Unmarshal(body, &tss)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server error"))
		log.Printf("Error: %s", err)
		return
	}

	err = sessManager.Create(&StreamerInfo{
		ChannelName:  tsi.Users[0].Name,
		Viewers:      strconv.Itoa(tss.Stream.Viewers),
		StatusStream: "True",
		Thumbnail:    tss.Stream.Preview.Large,
	})
	if err != nil {
		log.Printf("cant get data for %s: (%s)", streamerName, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Return stream info
	json.NewEncoder(w).Encode(&StreamerInfo{ChannelName: tsi.Users[0].Name, Viewers: strconv.Itoa(tss.Stream.Viewers), StatusStream: "True", Thumbnail: tss.Stream.Preview.Large})
}

func getStreamData(streamerName, clientID string, wg *sync.WaitGroup) {
	defer wg.Done()

	timeout := time.Duration(5 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}
	//Get streamer info
	request, err := http.NewRequest("GET", "https://api.twitch.tv/kraken/users?login="+string(streamerName), nil)
	request.Header.Set("Accept", "application/vnd.twitchtv.v5+json")
	request.Header.Set("Client-ID", clientID)

	resp, err := client.Do(request)
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error: %s", err)
	}

	var tsi TwitchStreamerInfo
	err = json.Unmarshal(body, &tsi)
	if err != nil {
		log.Fatalf("Error: %s", err)
	}

	if len(tsi.Users) == 0 {
		log.Printf("No data for User %s", streamerName)
		return
	}

	// Get stream status
	request, err = http.NewRequest("GET", "https://api.twitch.tv/kraken/streams/"+tsi.Users[0].ID, nil)
	request.Header.Set("Accept", "application/vnd.twitchtv.v5+json")
	request.Header.Set("Client-ID", clientID)

	resp, err = client.Do(request)
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
	defer resp.Body.Close()

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error: %s", err)
	}

	var tss TwitchStreamStatus
	err = json.Unmarshal(body, &tss)
	if err != nil {
		log.Fatalf("Error: %s", err)
	}

	// Show stream info
	log.Println(tsi.Users[0].Name, tss.Stream.Game, tss.Stream.Viewers)
}

func showStreamersInfo(w http.ResponseWriter, r *http.Request) {
	var wg sync.WaitGroup

	t1 := time.Now()

	var strNames Streamers
	err := json.Unmarshal([]byte(streamers), &strNames)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server error"))
		log.Fatalf("Error: %s", err)
	}

	for _, streamerName := range strNames.Streamer {
		wg.Add(1)
		go getStreamData(streamerName.Username, clientID, &wg)

	}

	wg.Wait()
	log.Printf("Elapsed time: %v", time.Since(t1))
}

func getStreamersInfo(w http.ResponseWriter, r *http.Request) {
	var wg sync.WaitGroup

	t1 := time.Now()

	var strNames Streamers
	err := json.Unmarshal([]byte(streamers), &strNames)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server error"))
		log.Fatalf("Error: %s", err)
	}

	for _, streamerName := range strNames.Streamer {
		wg.Add(1)
		go getStreamData(streamerName.Username, clientID, &wg)
	}

	wg.Wait()
	log.Printf("Elapsed time: %v", time.Since(t1))
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func main() {
	redisConn, err := redis.DialURL(redisAddr)
	if err != nil {
		log.Fatalf("cant connect to redis")
	}
	sessManager = NewConnManager(redisConn)

	rs := mux.NewRouter()
	rs.HandleFunc("/streamers/", showStreamersInfo).Methods("GET")
	rs.HandleFunc("/streamers/{streamerName}", getStreamerInfo).Methods("GET")

	port := "8000"
	smux := http.NewServeMux()

	smux.Handle("/streamers/", rs)

	err = http.ListenAndServe(":"+port, smux)
	if err != nil {
		log.Fatalln(err)

	}
}
