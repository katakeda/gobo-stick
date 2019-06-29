package main

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

type User struct {
	id       int
	username string
	password string
	email    string
}

type OauthClient struct {
	config *oauth2.Config
}

func (auth *OauthClient) init() {
	auth.config = &oauth2.Config{
		ClientID:     os.Getenv("CLIENT_ID"),
		ClientSecret: os.Getenv("CLIENT_SECRET"),
		RedirectURL:  os.Getenv("REDIRECT_URL"),
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
}

var auth OauthClient

var redisClient = redis.NewClient(&redis.Options{
	Addr:     "localhost:6000",
	Password: "",
	DB:       0,
})

func main() {
	// Load env variables
	godotenv.Load()

	// Initialize OAuth config
	auth.init()

	// Init Router
	router := mux.NewRouter()

	// Routes
	router.HandleFunc("/api/sign-in-with-google", handleSignInRequest).Methods("GET")
	router.HandleFunc("/api/auth/google/callback", handleCallback).Methods("GET")
	router.HandleFunc("/api/profile", getProfile).Methods("POST")

	log.Fatal(http.ListenAndServe(":8888", router))
}

func handleSignInRequest(w http.ResponseWriter, r *http.Request) {
	// Create an anti-forgery state token
	securekey := securecookie.GenerateRandomKey(32)
	state := base64.URLEncoding.EncodeToString(securekey)
	url := auth.config.AuthCodeURL(state)

	// Set state token as cookie
	expires := time.Now().Add(365 * 24 * time.Hour)
	cookie := http.Cookie{Name: "oauth_state", Value: state, Expires: expires}
	redisClient.Set(state, state, expires.Sub(expires))

	http.SetCookie(w, &cookie)
	http.Redirect(w, r, url, 302)
}

func handleCallback(w http.ResponseWriter, r *http.Request) {
	// Get code and state from query parameter
	code := r.FormValue("code")
	state := r.FormValue("state")

	// Validate state token
	if _, err := redisClient.Get(state).Result(); err != nil {
		log.Println("Unmatched state")
		return
	}

	// Exchange authorization code for token
	token, err := auth.config.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Fatal(err)
		return
	}

	// Fetch Google's openid configuration JSON
	openid, err := http.Get("https://accounts.google.com/.well-known/openid-configuration")
	if err != nil {
		log.Fatal(err)
		return
	}
	defer openid.Body.Close()
	data1, _ := ioutil.ReadAll(openid.Body)
	var config map[string]interface{}
	json.Unmarshal(data1, &config)

	// Fetch userinfo
	endpoint := fmt.Sprintf("%v", config["userinfo_endpoint"])
	client := auth.config.Client(oauth2.NoContext, token)
	response, err := client.Get(endpoint)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer response.Body.Close()
	data2, _ := ioutil.ReadAll(response.Body)
	var userinfo map[string]interface{}
	json.Unmarshal(data2, &userinfo)

	// Authenticate user
	var user User
	db, err := sql.Open("postgres", "hasura:hasura@tcp(localhost:5432)/hasura")
	if err != nil {
		log.Fatal(err)
		return
	}
	row := db.QueryRow(`SELECT * FROM user WHERE email=$1;`, userinfo["email"])
	row.Scan(&user.id, &user.username, &user.password, &user.email)
	log.Printf("%+v", user)
	defer db.Close()
}

func getProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode("")
}
