package internal

import (
	"encoding/json"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/ylubyanoy/go_web_server/internal/api/twitch_api"
	"github.com/ylubyanoy/go_web_server/internal/data"
	"github.com/ylubyanoy/go_web_server/internal/handlers"
	"github.com/ylubyanoy/go_web_server/internal/models"
	"github.com/ylubyanoy/go_web_server/internal/services"
	"github.com/ylubyanoy/go_web_server/internal/storages"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

var clientID string = "uqpc0satolohmpkplj0q0zgon883qx"

// BusinessLogic is main func for business logic for app
func BusinessLogic(logger *zap.SugaredLogger, storage storages.KeyStorage, port string, db data.Repository, shutdown chan<- error) *http.Server {

	validator := data.NewValidation()

	// authService contains all methods that help in authorizing a user request
	authService := services.NewAuthService(logger)

	// UserHandler encapsulates all the services related to user
	uh := handlers.NewAuthHandler(logger.With("handler", "AuthHandler"), validator, db, authService)

	r := mux.NewRouter()

	s := r.Methods(http.MethodGet).Subrouter()
	s.HandleFunc("/streamers/", handleStreamersInfo(logger.With("handler", "getStreamersInfo"))).Methods("POST")
	s.HandleFunc("/streamers/{streamerName}", handleStreamerInfo(logger.With("handler", "getStreamerInfo"), storage)).Methods("GET")
	s.Use(uh.MiddlewareValidateAccessToken)

	postR := r.Methods(http.MethodPost).Subrouter()
	postR.HandleFunc("/signup", uh.Signup)
	postR.HandleFunc("/login", uh.Login)
	postR.Use(uh.MiddlewareValidateUser)

	mailR := r.PathPrefix("/verify").Methods(http.MethodPost).Subrouter()
	mailR.HandleFunc("/mail", uh.VerifyMail)
	mailR.Use(uh.MiddlewareValidateVerificationData)

	// used the PathPrefix as workaround for scenarios where all the
	// get requests must use the ValidateAccessToken middleware except
	// the /refresh-token request which has to use ValidateRefreshToken middleware
	refToken := r.PathPrefix("/refresh-token").Subrouter()
	refToken.HandleFunc("", uh.RefreshToken)
	refToken.Use(uh.MiddlewareValidateRefreshToken)

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
		err := json.NewDecoder(r.Body).Decode(&strNames)
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

func handleStreamerInfo(logger *zap.SugaredLogger, storage storages.KeyStorage) func(http.ResponseWriter, *http.Request) {
	return func(
		w http.ResponseWriter, r *http.Request) {
		logger.Info("Received a call StreamerInfo")

		w.Header().Set("Content-Type", "application/json")
		params := mux.Vars(r)
		streamerName := params["streamerName"]

		// Check Redis
		si := storage.Check(streamerName)
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
		err = storage.Create(&models.StreamerInfo{
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
		json.NewEncoder(w).Encode(&models.StreamerInfo{
			ChannelName:  tsi.Users[0].Name,
			Game:         tss.Stream.Game,
			Viewers:      tss.Stream.Viewers,
			StatusStream: "true",
			Thumbnail:    tss.Stream.Preview.Large,
		})
	}
}
