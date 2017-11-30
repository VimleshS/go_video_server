package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"golang.org/x/oauth2"
)

func authHandler(w http.ResponseWriter, r *http.Request) {
	// Handle the exchange code to initiate a transport.
	session, _ := store.Get(r, "user-details")
	retrievedState := session.Values["state"]
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
	session.Values["user"] = &user
	session.Save(r, w)
	videoListHandler{}.ServeHTTP(w, r)
}
