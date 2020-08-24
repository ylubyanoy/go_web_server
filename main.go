package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	twc "github.com/ylubyanoy/go_web_server/internal"
	"go.uber.org/zap"

	"github.com/garyburd/redigo/redis"
	"github.com/gorilla/mux"
)

const servicename = "streamsinfo"

var clientID string = "uqpc0satolohmpkplj0q0zgon883qx"

var sessManager *ConnManager

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

	var tsi twc.TwitchStreamerInfo
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
		return
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

	var tss twc.TwitchStreamStatus
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
		log.Printf("Can't get data for %s: (%s)", streamerName, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Return stream info
	json.NewEncoder(w).Encode(&StreamerInfo{
		ChannelName:  tsi.Users[0].Name,
		Viewers:      strconv.Itoa(tss.Stream.Viewers),
		StatusStream: "True",
		Thumbnail:    tss.Stream.Preview.Large,
	})
}

func getStreamData(streamerName, clientID string, wg *sync.WaitGroup, streamers *[]StreamerInfo, mutex *sync.Mutex) {
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

	var tsi twc.TwitchStreamerInfo
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

	var tss twc.TwitchStreamStatus
	err = json.Unmarshal(body, &tss)
	if err != nil {
		log.Fatalf("Error: %s", err)
	}

	si := StreamerInfo{
		ChannelName:  tsi.Users[0].Name,
		Viewers:      strconv.Itoa(tss.Stream.Viewers),
		StatusStream: "True",
		Thumbnail:    tss.Stream.Preview.Large,
	}
	mutex.Lock()
	*streamers = append(*streamers, si)
	mutex.Unlock()
}

func getStreamersInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var wg sync.WaitGroup

	t1 := time.Now()

	var strNames Streamers
	err := json.Unmarshal([]byte(streamers), &strNames)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server error"))
		log.Fatalf("Error: %s", err)
	}

	var mutex = &sync.Mutex{}
	var streamers []StreamerInfo
	for _, streamerName := range strNames.Streamer {
		wg.Add(1)
		go getStreamData(streamerName.Username, clientID, &wg, &streamers, mutex)
	}

	wg.Wait()
	log.Printf("Elapsed time: %v", time.Since(t1))

	// Return streamers info
	json.NewEncoder(w).Encode(streamers)
}

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	appLoger := logger.Sugar().Named(servicename)
	appLoger.Info("The application is starting...")

	redisAddr := getEnv("REDIS_URL", "redis://user:@localhost:6379/0")
	redisConn := &redis.Pool{
		MaxIdle:     10,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			redisConn, err := redis.DialURL(redisAddr)
			if err != nil {
				appLoger.Fatalw("Can't connect to Redis", "err", err)
			}
			return redisConn, nil
		},
	}
	sessManager = NewConnManager(redisConn)
	rc := redisConn.Get()
	_, err := redis.String(rc.Do("PING"))
	if err != nil {
		appLoger.Fatalw("Can't connect to Redis", "err", err)
	}
	rc.Close()
	appLoger.Info("Connected to Redis")

	rs := mux.NewRouter()
	rs.HandleFunc("/streamers/", getStreamersInfo).Methods("GET")
	rs.HandleFunc("/streamers/{streamerName}", getStreamerInfo).Methods("GET")

	port := "8000"
	smux := http.NewServeMux()

	smux.Handle("/streamers/", rs)

	appLoger.Info("Server are ready")
	err = http.ListenAndServe(":"+port, smux)
	if err != nil {
		appLoger.Fatalw("Got an error from the server", "err", err)
	}

	appLoger.Info("The application is stopped.")
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
