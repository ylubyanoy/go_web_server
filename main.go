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

// type Book struct {
// 	ID     string  `json:"id"`
// 	Title  string  `json:"title"`
// 	Author *Author `json:"author"`
// }

// type Author struct {
// 	Firstname string `json:"firstname"`
// 	Lastname  string `json:"lastname"`
// }

// var books []Book

// var apiKey *string // 9ad6a77f98b24af492a46b7e450cc379

// var tpl = template.Must(template.ParseFiles("index.html"))

// //NewsAPIError represents the JSON response that is received from the News API whenever a request fails
// type NewsAPIError struct {
// 	Status  string `json:"status"`
// 	Code    string `json:"code"`
// 	Message string `json:"message"`
// }

// // Source is source for articles
// type Source struct {
// 	ID   interface{} `json:"id"`
// 	Name string      `json:"name"`
// }

// // Article is news article
// type Article struct {
// 	Source      Source    `json:"source"`
// 	Author      string    `json:"author"`
// 	Title       string    `json:"title"`
// 	Description string    `json:"description"`
// 	URL         string    `json:"url"`
// 	URLToImage  string    `json:"urlToImage"`
// 	PublishedAt time.Time `json:"publishedAt"`
// 	Content     string    `json:"content"`
// }

// // Results is the current page of results for the query
// type Results struct {
// 	Status       string    `json:"status"`
// 	TotalResults int       `json:"totalResults"`
// 	Articles     []Article `json:"articles"`
// }

// // Search is represents each search query made by the user
// type Search struct {
// 	SearchKey  string
// 	NextPage   int
// 	TotalPages int
// 	Results    Results
// }

// // FormatPublishedDate is format PublishedAt datefield
// func (a *Article) FormatPublishedDate() string {
// 	year, month, day := a.PublishedAt.Date()
// 	return fmt.Sprintf("%v %d, %d", month, day, year)
// }

// // IsLastPage is get last page state
// func (s *Search) IsLastPage() bool {
// 	return s.NextPage >= s.TotalPages
// }

// // CurrentPage is get current page number
// func (s *Search) CurrentPage() int {
// 	if s.NextPage == 1 {
// 		return s.NextPage
// 	}

// 	return s.NextPage - 1
// }

// // PreviousPage get previous page number
// func (s *Search) PreviousPage() int {
// 	return s.CurrentPage() - 1
// }

// // SearchHandler handler for search page
// func searchHandler(w http.ResponseWriter, r *http.Request) {
// 	u, err := url.Parse(r.URL.String())
// 	if err != nil {
// 		w.WriteHeader(http.StatusInternalServerError)
// 		w.Write([]byte("Internal Server error"))
// 		return
// 	}

// 	params := u.Query()
// 	searchKey := params.Get("q")
// 	page := params.Get("page")
// 	if page == "" {
// 		page = "1"
// 	}

// 	search := &Search{}
// 	search.SearchKey = searchKey

// 	next, err := strconv.Atoi(page)
// 	if err != nil {
// 		http.Error(w, "Unexpected server error", http.StatusInternalServerError)
// 		return
// 	}

// 	search.NextPage = next
// 	pageSize := 20

// 	endpoint := fmt.Sprintf("https://newsapi.org/v2/everything?q=%s&pageSize=%d&page=%d&apiKey=%s&sortBy=publishedAt&language=en", url.QueryEscape(search.SearchKey), pageSize, search.NextPage, *apiKey)
// 	resp, err := http.Get(endpoint)
// 	if err != nil {
// 		w.WriteHeader(http.StatusInternalServerError)
// 		return
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode != 200 {
// 		newError := &NewsAPIError{}
// 		err := json.NewDecoder(resp.Body).Decode(newError)
// 		if err != nil {
// 			http.Error(w, "Unexpected server error", http.StatusInternalServerError)
// 			return
// 		}

// 		http.Error(w, newError.Message, http.StatusInternalServerError)
// 		return
// 	}

// 	err = json.NewDecoder(resp.Body).Decode(&search.Results)
// 	if err != nil {
// 		w.WriteHeader(http.StatusInternalServerError)
// 		return
// 	}

// 	search.TotalPages = int(math.Ceil(float64(search.Results.TotalResults / pageSize)))

// 	if ok := !search.IsLastPage(); ok {
// 		search.NextPage++
// 	}

// 	err = tpl.Execute(w, search)
// 	if err != nil {
// 		log.Println(err)
// 	}

// }

// // IndexHandler handler for main page
// func indexHandler(w http.ResponseWriter, r *http.Request) {
// 	tpl.Execute(w, nil)
// }

// func getBooks(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(books)
// }

// func getBook(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("Content-Type", "application/json")
// 	params := mux.Vars(r)
// 	for _, item := range books {
// 		if item.ID == params["id"] {
// 			json.NewEncoder(w).Encode(item)
// 			return
// 		}
// 	}
// 	json.NewEncoder(w).Encode(&Book{})
// }

// func createBook(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("Content-Type", "application/json")
// 	var book Book
// 	_ = json.NewDecoder(r.Body).Decode(&book)
// 	book.ID = strconv.Itoa(rand.Intn(1000000))
// 	books = append(books, book)
// 	json.NewEncoder(w).Encode(book)
// }

// func updateBook(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("Content-Type", "application/json")
// 	params := mux.Vars(r)

// 	for index, item := range books {
// 		if item.ID == params["id"] {
// 			books = append(books[:index], books[index+1:]...)
// 			var book Book
// 			_ = json.NewDecoder(r.Body).Decode(&book)
// 			book.ID = params["id"]
// 			books = append(books, book)
// 			json.NewEncoder(w).Encode(book)
// 			return
// 		}
// 	}
// 	json.NewEncoder(w).Encode(books)
// }

// func deleteBook(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("Content-Type", "application/json")
// 	params := mux.Vars(r)

// 	for index, item := range books {
// 		if item.ID == params["id"] {
// 			books = append(books[:index], books[index+1:]...)
// 			break
// 		}
// 	}
// 	json.NewEncoder(w).Encode(books)
// }

var clientID string = "uqpc0satolohmpkplj0q0zgon883qx"

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

var (
	redisAddr   = getEnv("REDIS_URL", "redis://user:@localhost:6379/0")
	sessManager *ConnManager
)

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
	// books = append(books, Book{ID: "1", Title: "Война и Мир", Author: &Author{Firstname: "Лев", Lastname: "Толстой"}})
	// books = append(books, Book{ID: "2", Title: "Преступление и наказание", Author: &Author{Firstname: "Фёдор", Lastname: "Достоевский"}})

	// r := mux.NewRouter()
	// r.HandleFunc("/books/", getBooks).Methods("GET")
	// r.HandleFunc("/books/{id}", getBook).Methods("GET")
	// r.HandleFunc("/books/", createBook).Methods("POST")
	// r.HandleFunc("/books/{id}", updateBook).Methods("PUT")
	// r.HandleFunc("/books/{id}", deleteBook).Methods("DELETE")

	rs := mux.NewRouter()
	rs.HandleFunc("/streamers/", showStreamersInfo).Methods("GET")
	rs.HandleFunc("/streamers/{streamerName}", getStreamerInfo).Methods("GET")

	// apiKey = flag.String("apikey", "", "Newsapi.org access key")
	// flag.Parse()
	// if *apiKey == "" {
	// 	log.Fatal("apiKey must be set")
	// }

	port := "8000"
	smux := http.NewServeMux()

	// fs := http.FileServer(http.Dir("assets"))
	// smux.Handle("/assets/", http.StripPrefix("/assets/", fs))

	smux.Handle("/streamers/", rs)

	// smux.Handle("/books/", r)
	// smux.HandleFunc("/search", searchHandler)
	// smux.HandleFunc("/", indexHandler)

	err = http.ListenAndServe(":"+port, smux)
	if err != nil {
		log.Fatalln(err)

	}
}

type StreamerNickName struct {
	Username string `json:"username"`
}

type Streamers struct {
	Streamer []StreamerNickName `json:"users"`
}

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

type StreamerInfo struct {
	ChannelName  string `json:"channel_name"`
	Viewers      string `json:"viewers"`
	StatusStream string `json:"status_stream"`
	Thumbnail    string `json:"thumbnail"`
}
