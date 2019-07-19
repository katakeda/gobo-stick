package main

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

// User ...
type User struct {
	id       int
	username string
	password string
	email    string
}

func main() {
	// Load env variables
	godotenv.Load()

	// Initialize OAuth config
	auth.init()

	// Initialize Redis config
	redisClient.init()

	// Initialize OpenID config
	if err := openid.init(); err != nil {
		log.Fatal(err)
	}

	// Initialize Database config
	if err := db.init(); err != nil {
		log.Fatal(err)
	}

	// Init Router
	router := mux.NewRouter()

	// Routes
	router.HandleFunc("/api/sign-in-with-google", handleSignInRequest).Methods("GET")
	router.HandleFunc("/api/auth/google/callback", handleCallback).Methods("GET")
	router.HandleFunc("/api/profile", getProfile).Methods("GET")

	log.Fatal(http.ListenAndServe(os.Getenv("HTTP_PORT"), router))
}

func handleSignInRequest(w http.ResponseWriter, r *http.Request) {
	// Create an anti-forgery state token
	securekey := securecookie.GenerateRandomKey(32)
	state := base64.URLEncoding.EncodeToString(securekey)
	url := auth.config.AuthCodeURL(state)

	// Store state token for validation
	redisClient.client.Set(state, state, time.Minute)

	http.Redirect(w, r, url, 302)
}

func handleCallback(w http.ResponseWriter, r *http.Request) {
	// Get code and state from query parameter
	code := r.FormValue("code")
	state := r.FormValue("state")

	// Validate state token
	if err := redisClient.client.Get(state).Err(); err != nil {
		http.Redirect(w, r, os.Getenv("APP_URL"), 302)
		return
	}

	// Exchange authorization code for token
	token, err := auth.config.Exchange(oauth2.NoContext, code)
	if err != nil {
		http.Redirect(w, r, os.Getenv("APP_URL"), 302)
		return
	}

	// Fetch userinfo
	client := auth.config.Client(oauth2.NoContext, token)
	response, err := client.Get(openid.get("userinfo_endpoint"))
	if err != nil {
		http.Redirect(w, r, os.Getenv("APP_URL"), 302)
		return
	}
	defer response.Body.Close()
	data, _ := ioutil.ReadAll(response.Body)
	var userinfo map[string]interface{}
	json.Unmarshal(data, &userinfo)

	// Authenticate user
	var user User
	if err := db.client.PingContext(oauth2.NoContext); err != nil {
		http.Redirect(w, r, os.Getenv("APP_URL"), 302)
		return
	}

	defer db.client.Close()
	row := db.client.QueryRow("SELECT id, email FROM user WHERE email=?", userinfo["email"])
	row.Scan(&user.id, &user.email)
	if user.email == "" {
		http.Redirect(w, r, os.Getenv("APP_URL"), 302)
		return
	}

	// Set access token key in cookie and key-value pair in Redis
	securekey := base64.URLEncoding.EncodeToString(securecookie.GenerateRandomKey(32))
	expires := time.Now().Add(time.Hour)
	cookie := http.Cookie{
		Name:     os.Getenv("SESSION_NAME"),
		Value:    securekey,
		Path:     "/",
		Expires:  expires,
		HttpOnly: true,
	}
	redisClient.client.Set(securekey, token.AccessToken, time.Hour)
	http.SetCookie(w, &cookie)
	http.Redirect(w, r, os.Getenv("APP_URL"), 302)
}

func getProfile(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(os.Getenv("SESSION_NAME"))
	if err != nil {
		http.Redirect(w, r, os.Getenv("APP_URL"), 302)
		return
	}

	// Exchange cookie value with access token
	accessToken, err := redisClient.client.Get(cookie.Value).Result()
	if err != nil {
		http.Redirect(w, r, os.Getenv("APP_URL"), 302)
		return
	}

	// Fetch userinfo
	response, err := http.Get(openid.get("userinfo_endpoint") + "?access_token=" + accessToken)
	if err != nil {
		http.Redirect(w, r, os.Getenv("APP_URL"), 302)
		return
	}
	defer response.Body.Close()
	data, _ := ioutil.ReadAll(response.Body)
	var userinfo map[string]interface{}
	json.Unmarshal(data, &userinfo)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userinfo)
}
