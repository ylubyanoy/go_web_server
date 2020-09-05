package internal

import (
	"encoding/json"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/ylubyanoy/go_web_server/internal/api/twitch_api"
	"github.com/ylubyanoy/go_web_server/internal/models"
	"github.com/ylubyanoy/go_web_server/internal/storages"
	"github.com/ylubyanoy/go_web_server/internal/storages/redis_store"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

var streamersList = `{
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

var clientID string = "uqpc0satolohmpkplj0q0zgon883qx"

// func getStreamData(streamerName, clientID string, wg *sync.WaitGroup, streamers *[]storages.StreamerInfo, mutex *sync.Mutex) {
// 	defer wg.Done()

// 	timeout := time.Duration(5 * time.Second)
// 	client := http.Client{
// 		Timeout: timeout,
// 	}
// 	//Get streamer info
// 	request, err := http.NewRequest("GET", "https://api.twitch.tv/kraken/users?login="+string(streamerName), nil)
// 	request.Header.Set("Accept", "application/vnd.twitchtv.v5+json")
// 	request.Header.Set("Client-ID", clientID)

// 	resp, err := client.Do(request)
// 	if err != nil {
// 		log.Fatalf("Error: %s", err)
// 	}
// 	defer resp.Body.Close()

// 	body, err := ioutil.ReadAll(resp.Body)
// 	if err != nil {
// 		log.Fatalf("Error: %s", err)
// 	}

// 	var tsi twc.TwitchStreamerInfo
// 	err = json.Unmarshal(body, &tsi)
// 	if err != nil {
// 		log.Fatalf("Error: %s", err)
// 	}

// 	if len(tsi.Users) == 0 {
// 		log.Printf("No data for User %s", streamerName)
// 		return
// 	}

// 	// Get stream status
// 	request, err = http.NewRequest("GET", "https://api.twitch.tv/kraken/streams/"+tsi.Users[0].ID, nil)
// 	request.Header.Set("Accept", "application/vnd.twitchtv.v5+json")
// 	request.Header.Set("Client-ID", clientID)

// 	resp, err = client.Do(request)
// 	if err != nil {
// 		log.Fatalf("Error: %s", err)
// 	}
// 	defer resp.Body.Close()

// 	body, err = ioutil.ReadAll(resp.Body)
// 	if err != nil {
// 		log.Fatalf("Error: %s", err)
// 	}

// 	var tss twc.TwitchStreamStatus
// 	err = json.Unmarshal(body, &tss)
// 	if err != nil {
// 		log.Fatalf("Error: %s", err)
// 	}

// 	if tss.Stream.Viewers == 0 {
// 		log.Printf("No stream data for User %s", streamerName)
// 		return
// 	}

// 	si := storages.StreamerInfo{
// 		ChannelName:  tsi.Users[0].Name,
// 		Game:         tss.Stream.Game,
// 		Viewers:      tss.Stream.Viewers,
// 		StatusStream: "true",
// 		Thumbnail:    tss.Stream.Preview.Large,
// 	}
// 	mutex.Lock()
// 	*streamers = append(*streamers, si)
// 	mutex.Unlock()
// }

// BusinessLogic is main func for business logic for app
func BusinessLogic(logger *zap.SugaredLogger, redisAddr string, port string, shutdown chan<- error) *http.Server {

	sessManager, err := redis_store.New(redisAddr)
	if err != nil {
		logger.Fatalw("Can't connect to Redis", "err", err)
	}
	logger.Info("Connected to Redis")

	r := mux.NewRouter()
	r.HandleFunc("/streamers/", handleStreamersInfo(logger.With("handler", "getStreamersInfo"))).Methods("GET")
	r.HandleFunc("/streamers/{streamerName}", handleStreamerInfo(logger.With("handler", "getStreamerInfo"), sessManager)).Methods("GET")

	server := http.Server{
		Addr:    net.JoinHostPort("", port),
		Handler: r,
	}

	logger.Info("Ready to start the server...")
	go func() {
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			shutdown <- err
		}
	}()

	return &server
}

func handleStreamersInfo(logger *zap.SugaredLogger) func(http.ResponseWriter, *http.Request) {
	return func(
		w http.ResponseWriter, r *http.Request) {
		logger.Info("Received a call StreamersInfo")

		w.Header().Set("Content-Type", "application/json")

		var wg sync.WaitGroup

		t1 := time.Now()

		var strNames models.Streamers
		err := json.Unmarshal([]byte(streamersList), &strNames)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server error"))
			logger.Fatalw("Error: %s", err)
		}

		var mutex = &sync.Mutex{}
		var streamers []models.StreamerInfo

		for _, streamerName := range strNames.Streamer {
			wg.Add(1)
			go func(streamerName, clientID string, wg *sync.WaitGroup, streamers *[]models.StreamerInfo, mutex *sync.Mutex) {
				defer wg.Done()

				apiClient := twitch_api.NewTwitchClient()

				tsi, err := apiClient.GetStreamerInfo(streamerName, clientID)
				if err != nil {
					logger.Fatalf("Error: %s", err)
				}

				if len(tsi.Users) == 0 {
					logger.Infof("No data for User %s", streamerName)
					return
				}

				tss, err := apiClient.GetStreamStatus(tsi.Users[0].ID, clientID)
				if err != nil {
					logger.Fatalf("Error: %s", err)
				}

				if tss.Stream.Viewers == 0 {
					logger.Infof("No stream data for User %s", streamerName)
					return
				}

				si := models.StreamerInfo{
					ChannelName:  tsi.Users[0].Name,
					Game:         tss.Stream.Game,
					Viewers:      tss.Stream.Viewers,
					StatusStream: "true",
					Thumbnail:    tss.Stream.Preview.Large,
				}
				mutex.Lock()
				*streamers = append(*streamers, si)
				mutex.Unlock()

			}(streamerName.Username, clientID, &wg, &streamers, mutex)
		}

		wg.Wait()
		logger.Infof("Elapsed time: %v", time.Since(t1))

		// Return streamers info
		json.NewEncoder(w).Encode(streamers)

	}
}

func handleStreamerInfo(logger *zap.SugaredLogger, sessManager *redis_store.ConnManager) func(http.ResponseWriter, *http.Request) {
	return func(
		w http.ResponseWriter, r *http.Request) {
		logger.Info("Received a call StreamerInfo")

		w.Header().Set("Content-Type", "application/json")
		params := mux.Vars(r)
		streamerName := params["streamerName"]

		// Check Redis
		si := sessManager.Check(streamerName)
		if si != nil {
			logger.Info("Get from Redis")
			json.NewEncoder(w).Encode(si)
			return
		}

		apiClient := twitch_api.NewTwitchClient()

		tsi, err := apiClient.GetStreamerInfo(streamerName, clientID)
		if err != nil {
			logger.Fatalf("Error: %s", err)
		}

		if len(tsi.Users) == 0 {
			logger.Infof("No data for User %s", streamerName)
			return
		}

		tss, err := apiClient.GetStreamStatus(tsi.Users[0].ID, clientID)
		if err != nil {
			logger.Fatalf("Error: %s", err)
		}

		if tss.Stream.Viewers == 0 {
			logger.Infof("No stream data for User %s", streamerName)
			return
		}

		// Save to Redis
		err = sessManager.Create(&storages.StreamerInfo{
			ChannelName:  tsi.Users[0].Name,
			Game:         tss.Stream.Game,
			Viewers:      tss.Stream.Viewers,
			StatusStream: "true",
			Thumbnail:    tss.Stream.Preview.Large,
		})
		if err != nil {
			logger.Infow("Can't set data for %s: (%s)", streamerName, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Return stream info
		json.NewEncoder(w).Encode(&storages.StreamerInfo{
			ChannelName:  tsi.Users[0].Name,
			Game:         tss.Stream.Game,
			Viewers:      tss.Stream.Viewers,
			StatusStream: "true",
			Thumbnail:    tss.Stream.Preview.Large,
		})
	}
}
