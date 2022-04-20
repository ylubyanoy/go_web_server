package internal

import (
	"encoding/json"
	"net"
	"net/http"
	"regexp"
	"sync"
	"time"

	"github.com/ylubyanoy/go_web_server/internal/api/twitch_api"
	"github.com/ylubyanoy/go_web_server/internal/config"
	"github.com/ylubyanoy/go_web_server/internal/data"
	"github.com/ylubyanoy/go_web_server/internal/handlers"
	"github.com/ylubyanoy/go_web_server/internal/models"
	"github.com/ylubyanoy/go_web_server/internal/services"
	"github.com/ylubyanoy/go_web_server/internal/storages"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// BusinessLogic is main func for business logic for app
func BusinessLogic(logger *zap.SugaredLogger, storage storages.KeyStorage, cfg *config.Config, db data.Repository, shutdown chan<- error) *http.Server {

	validator := data.NewValidation()

	// authService contains all methods that help in authorizing a user request
	authService := services.NewAuthService(logger)

	// UserHandler encapsulates all the services related to user
	uh := handlers.NewAuthHandler(logger.With("handler", "AuthHandler"), validator, db, authService)

	r := mux.NewRouter()

	s := r.Methods(http.MethodGet, http.MethodPost).Subrouter()
	s.HandleFunc("/streamers/", handleStreamersInfo(logger.With("handler", "getStreamersInfo"), storage, cfg)).Methods("POST")
	s.HandleFunc("/api/v2/streamers/", handleStreamersInfo(logger.With("handler", "getStreamersInfo"), storage, cfg)).Methods("POST")
	s.HandleFunc("/streamers/{streamerName}", handleStreamerInfo(logger.With("handler", "getStreamerInfo"), storage, cfg)).Methods("GET")
	s.HandleFunc("/api/v2/streamers/{streamerName}", handleStreamerInfo(logger.With("handler", "getStreamerInfo"), storage, cfg)).Methods("GET")
	s.Use(uh.MiddlewareValidateAccessToken)

	postR := r.Methods(http.MethodPost).Subrouter()
	postR.HandleFunc("/signup", uh.Signup)
	postR.HandleFunc("/login", uh.Login)
	postR.Use(uh.MiddlewareValidateUser)

	mailR := r.PathPrefix("/verify").Methods(http.MethodPost).Subrouter()
	mailR.HandleFunc("/mail", uh.VerifyMail)
	mailR.Use(uh.MiddlewareValidateVerificationData)

	refToken := r.PathPrefix("/refresh-token").Subrouter()
	refToken.HandleFunc("", uh.RefreshToken)
	refToken.Use(uh.MiddlewareValidateRefreshToken)

	server := http.Server{
		Addr:    net.JoinHostPort("", cfg.Port),
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

func handleStreamersInfo(logger *zap.SugaredLogger, storage storages.KeyStorage, cfg *config.Config) func(http.ResponseWriter, *http.Request) {
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
			go func(streamerName string, wg *sync.WaitGroup, streamers *[]models.StreamerInfo, cfg *config.Config, mutex *sync.Mutex) {
				defer wg.Done()

				// Check Redis
				si := storage.Check(streamerName)
				if si != nil {
					logger.Info("Get from Redis for" + streamerName)
					json.NewEncoder(w).Encode(si)
					return
				}

				apiClient := twitch_api.NewTwitchClient()

				// Get token
				token := storage.CheckToken("token")
				if token == "" {
					logger.Info("Get new token from Twitch")

					token, err := apiClient.GetAccessToken(cfg.ClientID, cfg.ClientSecret)
					if err != nil {
						logger.Errorf("Error: %s", err)
						return
					}

					// Save to Redis
					err = storage.CreateToken(token, cfg.TokenExpiresTime)
					if err != nil {
						logger.Infow("Can't set data (%s)", err)
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
				}

				tsi, err := apiClient.GetStreamerInfov2(streamerName, cfg.ClientID, token)
				if err != nil {
					logger.Fatalf("Error: %s", err)
				}

				if len(tsi.Data) == 0 {
					logger.Infof("No data for User %s", streamerName)
					return
				}

				tss, err := apiClient.GetStreamStatusv2(tsi.Data[0].ID, cfg.ClientID, token)
				if err != nil {
					logger.Fatalf("Error: %s", err)
				}

				if len(tss.Data) == 0 || tss.Data[0].ViewerCount == 0 {
					logger.Infof("No stream data for User %s", streamerName)
					return
				}

				// Change url for {640} {360}
				re, _ := regexp.Compile("{width}")
				re2, _ := regexp.Compile("{height}")
				ThumbnailURL := tss.Data[0].ThumbnailURL
				thumbnail := re.ReplaceAllString(ThumbnailURL, "640")
				thumbnail = re2.ReplaceAllString(thumbnail, "360")

				si = &models.StreamerInfo{
					ChannelName:  tsi.Data[0].DisplayName,
					Game:         tss.Data[0].GameName,
					Viewers:      int(tss.Data[0].ViewerCount),
					StatusStream: "true",
					Thumbnail:    thumbnail,
				}

				mutex.Lock()
				*streamers = append(*streamers, *si)
				mutex.Unlock()

				// Save to Redis
				err = storage.Create(si, cfg.StreamerDataExpiresTime)
				if err != nil {
					logger.Infow("Can't set data for %s: (%s)", streamerName, err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

			}(streamerName.Username, &wg, &streamers, cfg, mutex)
		}

		wg.Wait()
		logger.Infof("Elapsed time: %v", time.Since(t1))

		// Return streamers info
		json.NewEncoder(w).Encode(streamers)
	}
}

func handleStreamerInfo(logger *zap.SugaredLogger, storage storages.KeyStorage, cfg *config.Config) func(http.ResponseWriter, *http.Request) {
	return func(
		w http.ResponseWriter, r *http.Request) {
		logger.Info("Received a call StreamerInfo")

		w.Header().Set("Content-Type", "application/json")
		params := mux.Vars(r)
		streamerName := params["streamerName"]

		// Check Redis
		si := storage.Check(streamerName)
		if si != nil {
			logger.Info("Get from Redis for" + streamerName)
			json.NewEncoder(w).Encode(si)
			return
		}

		apiClient := twitch_api.NewTwitchClient()

		// Get token
		token := storage.CheckToken("token")
		if token == "" {
			logger.Info("Get new token from Twitch")

			token, err := apiClient.GetAccessToken(cfg.ClientID, cfg.ClientSecret)
			if err != nil {
				logger.Errorf("Error: %s", err)
				return
			}

			// Save to Redis
			err = storage.CreateToken(token, cfg.TokenExpiresTime)
			if err != nil {
				logger.Infow("Can't set data (%s)", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		tsi, err := apiClient.GetStreamerInfov2(streamerName, cfg.ClientID, token)
		if err != nil {
			logger.Errorf("Error: %s", err)
			return
		}
		if len(tsi.Data) == 0 {
			logger.Infof("No data for User %s", streamerName)
			return
		}

		tss, err := apiClient.GetStreamStatusv2(streamerName, cfg.ClientID, token)
		if err != nil {
			logger.Errorf("Error: %s", err)
			return
		}

		if len(tss.Data) == 0 || tss.Data[0].ViewerCount == 0 {
			logger.Infof("No stream data for User %s", streamerName)
			return
		}

		// Change url for {640} {360}
		re, _ := regexp.Compile("{width}")
		re2, _ := regexp.Compile("{height}")
		ThumbnailURL := tss.Data[0].ThumbnailURL
		thumbnail := re.ReplaceAllString(ThumbnailURL, "640")
		thumbnail = re2.ReplaceAllString(thumbnail, "360")

		StreamInfo := &models.StreamerInfo{
			ChannelName:  tsi.Data[0].DisplayName,
			Game:         tss.Data[0].GameName,
			Viewers:      int(tss.Data[0].ViewerCount),
			StatusStream: "true",
			Thumbnail:    thumbnail,
		}

		// Save to Redis
		err = storage.Create(StreamInfo, cfg.StreamerDataExpiresTime)
		if err != nil {
			logger.Infow("Can't set data for %s: (%s)", streamerName, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Return stream info
		json.NewEncoder(w).Encode(StreamInfo)
	}
}
