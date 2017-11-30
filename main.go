// https://skarlso.github.io/2016/06/12/google-signin-with-go/
// https://github.com/google/google-api-go-client/blob/master/GettingStarted.md
// https://github.com/golang/oauth2
//http://localhost:9090/oauth_redirect_uri

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
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	cred  Credentials
	conf  *oauth2.Config
	state string
	store = sessions.NewCookieStore([]byte("video-server-#$%"))
)

func init() {
	/*
		logwritter, err := syslog.New(syslog.LOG_INFO, "VIDEO_SERVER_APP")
		if err != nil {
			fmt.Printf("error in initializing syslog error: %v\n", err)
			os.Exit(1)
		}
		logwritter.Info("HI VIMLESH")
	*/

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

func main() {
	r := mux.NewRouter()
	//static handlers
	fs := http.FileServer(http.Dir("static"))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", processEncryptedVideoURL(fs)))
	fs1 := http.FileServer(http.Dir("resourses"))
	r.PathPrefix("/resourses/").Handler(http.StripPrefix("/resourses/", fs1))

	//flow handlers
	r.HandleFunc("/", roothandler)
	r.HandleFunc("/login", loginHandler).Methods("GET")
	r.HandleFunc("/auth", authHandler).Methods("GET")
	r.Handle("/list", videoListHandler{}).Methods("GET")
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
	state := videoURLCrypto{}.randomString(6)

	// set public key in cookie for decrypting names and play list
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

func playHandler(w http.ResponseWriter, r *http.Request) {
	file := r.URL.Query().Get("file")
	videopath := fmt.Sprintf("videos/%s", file)
	fmt.Println(videopath)

	var pubKey string
	session, _ := store.Get(r, "user-details")
	_pubKey := session.Values["pubkey"]
	if _pubKey == nil {
		http.Error(w, "(Play) Error Video File not found", 500)
		return
	}
	pubKey = _pubKey.(string)
	videoURL := videoURLCrypto{Pubkey: pubKey, Source: videopath}.doEncrypt()
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
