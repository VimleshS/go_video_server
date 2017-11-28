package main

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type page struct {
	WsEndPoint string
	VideoURL   string
	Name       string
}

type FileInfo struct {
	Name        string
	IsDirectory bool
	Path        string
}

// Credentials which stores google ids.
type Credentials struct {
	Cid     string `json:"cid"`
	Csecret string `json:"csecret"`
}

var (
	cred  Credentials
	conf  *oauth2.Config
	state string
	store = sessions.NewCookieStore([]byte("video-server-#$%"))
)

// User is a retrieved and authentiacted user.
type User struct {
	Sub           string `json:"sub"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Profile       string `json:"profile"`
	Picture       string `json:"picture"`
	Email         string `json:"email"`
	EmailVerified string `json:"email_verified"`
	Gender        string `json:"gender"`
}

func init() {
	file, err := ioutil.ReadFile("./creds.json")
	if err != nil {
		fmt.Printf("File error: %v\n", err)
		os.Exit(1)
	}
	json.Unmarshal(file, &cred)

	conf = &oauth2.Config{
		ClientID:     cred.Cid,
		ClientSecret: cred.Csecret,
		RedirectURL:  "http://127.0.0.1:9090/auth",
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email", // You have to select your own scope from here -> https://developers.google.com/identity/protocols/googlescopes#google_sign-in
		},
		Endpoint: google.Endpoint,
	}
	gob.Register(&User{})
}

func decorate(h http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("-------------------------static handler----------------------")
		/*
			c, _ := r.Cookie("pubkey")
			fmt.Printf(" value from cookie %v  \n", c.Value)
			if c != nil {
				fmt.Println("r.URL.Path", r.URL.Path)
				fmt.Println("c.Value", c.Value)

				videoURL := doDecrypt(c.Value, r.URL.Path)
				r.URL.Path = videoURL

				fmt.Println("new r.URL.Path", r.URL.Path)

				h.ServeHTTP(w, r)
			}
		*/

		session, err := store.Get(r, "user-details")
		if err != nil {
			http.Error(w, "Video File not found", 500)
			return
		}
		pubKey := session.Values["pubkey"]
		fmt.Printf(" value from cookie %v  \n", pubKey)
		if pubKey == nil {
			http.Error(w, "Video File not found", 500)
			return

		}
		_pubKey := pubKey.(string)
		videoURL := doDecrypt(_pubKey, r.URL.Path)
		r.URL.Path = videoURL
		h.ServeHTTP(w, r)
	})
}

// https://skarlso.github.io/2016/06/12/google-signin-with-go/
// https://github.com/google/google-api-go-client/blob/master/GettingStarted.md
// https://github.com/golang/oauth2
//http://localhost:9090/oauth_redirect_uri

func main() {
	r := mux.NewRouter()
	//static handlers
	fs := http.FileServer(http.Dir("static"))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", decorate(fs)))
	fs1 := http.FileServer(http.Dir("resourses"))
	r.PathPrefix("/resourses/").Handler(http.StripPrefix("/resourses/", fs1))

	//flow handlers
	r.HandleFunc("/", roothandler)
	r.HandleFunc("/login", loginHandler).Methods("GET")
	r.HandleFunc("/auth", authHandler).Methods("GET")
	r.HandleFunc("/list", videolistHandler).Methods("GET")
	r.HandleFunc("/play", playHandler).Methods("GET")

	srv := &http.Server{
		Handler: r,
		Addr:    "127.0.0.1:9090",

		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}

func roothandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/index.html", "templates/login.html")
	if err != nil {
		log.Println(err.Error())
		http.Error(w, http.StatusText(500), 500)
		return
	}
	//generate state key
	state := randomString(6)

	// set public key in cookie for decrypting names and play list
	/*
		expiration := time.Now().Add(10 * time.Minute)
		cookie := http.Cookie{
			Name:    "state",
			Value:   state,
			Expires: expiration,
		}
		http.SetCookie(w, &cookie)
	*/
	session, _ := store.Get(r, "user-details")
	session.Values["state"] = state
	session.Save(r, w)

	data := struct {
		State string
	}{state}

	tmpl.ExecuteTemplate(w, "layout", data)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")
	session, _ := store.Get(r, "user-details")
	sessionState := session.Values["state"]
	if sessionState != state {
		http.Error(w, "Request from from an unknown source", 400)
		return
	}
	authURL := conf.AuthCodeURL(state)
	http.Redirect(w, r, authURL, http.StatusFound)
}
func authHandler(w http.ResponseWriter, r *http.Request) {
	// Handle the exchange code to initiate a transport.
	session, _ := store.Get(r, "user-details")
	retrievedState := session.Values["state"]
	// retrievedState := session.Get("state")
	if retrievedState != r.URL.Query().Get("state") {
		http.Error(w, http.StatusText(500), 500)
		return
	}

	fmt.Println("------------- AUTH URL --------------------------")
	fmt.Println(r.URL)
	code := r.URL.Query().Get("code")

	tok, err := conf.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, http.StatusText(500), 500)
		return
	}

	client := conf.Client(oauth2.NoContext, tok)
	email, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		log.Println(err.Error())
		http.Error(w, http.StatusText(400), http.StatusBadRequest)
		return
	}
	defer email.Body.Close()
	data, _ := ioutil.ReadAll(email.Body)
	log.Println("Email body: ", string(data))

	user := User{}
	json.Unmarshal(data, &user)
	session.Values["user"] = user
	session.Save(r, w)
	videolistHandler(w, r)
}

func videolistHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/index.html", "templates/list.html")
	if err != nil {
		log.Println(err.Error())
		http.Error(w, http.StatusText(400), http.StatusBadRequest)
		return
	}

	//generate random key
	pubKey := randomString(6)
	fmt.Println("Pub key", pubKey)
	session, _ := store.Get(r, "user-details")
	session.Values["pubkey"] = pubKey
	user := session.Values["user"].(User)
	session.Save(r, w)

	files := getVideoFilesInDirectory()
	data := struct {
		VideoFiles []FileInfo
		UserInfo   User
	}{files,
		user}

	// // set public key in cookie for decrypting names and play list
	// expiration := time.Now().Add(60 * time.Minute)
	// cookie := http.Cookie{
	// 	Name:    "pubkey",
	// 	Value:   pubKey,
	// 	Expires: expiration,
	// }
	// http.SetCookie(w, &cookie)

	tmpl.ExecuteTemplate(w, "layout", data)
}

func playHandler(w http.ResponseWriter, r *http.Request) {
	file := r.URL.Query().Get("file")
	videopath := fmt.Sprintf("videos/%s", file)
	fmt.Println(videopath)

	var pubKey string
	/*
		c, _ := r.Cookie("pubkey")
		fmt.Printf(" decrypt value from cookie %v  \n", c.Value)

		if c == nil {
			pubKey = c.Value
		} else {
			pubKey = c.Value
		}
	*/
	session, _ := store.Get(r, "user-details")
	_pubKey := session.Values["pubkey"]
	if _pubKey == nil {
		http.Error(w, "(Play) Error Video File not found", 500)
		return

	}
	pubKey = _pubKey.(string)

	videoURL := doEncrypt(pubKey, videopath)
	evideoURL := "/static/" + videoURL

	tmpl, err := template.ParseFiles("templates/index.html", "templates/play.html")
	if err != nil {
		log.Println(err.Error())
		http.Error(w, http.StatusText(500), 500)
		return
	}
	p := page{
		VideoURL: evideoURL,
		Name:     file,
	}
	tmpl.ExecuteTemplate(w, "layout", p)
}

func getVideoFilesInDirectory() []FileInfo {
	fileList := []FileInfo{}
	filepath.Walk("./static/videos/", func(path string, info os.FileInfo, err error) error {
		path = strings.TrimPrefix(path, "static/videos/")
		if len(path) > 0 {
			fi := FileInfo{Name: path, Path: path, IsDirectory: info.IsDir()}
			fileList = append(fileList, fi)
		}
		return nil
	})
	return fileList
}
