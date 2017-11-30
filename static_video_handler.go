package main

import (
	"fmt"
	"net/http"
)

func processEncryptedVideoURL(h http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// playedTime, err := r.Cookie("played_time")
		// if err != nil {
		// 	fmt.Println(err.Error())
		// }
		// fmt.Printf(" value from cookie %v  \n", playedTime.Value)

		pubKey, err := retrivePublicKey(r)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		videoURL := videoURLCrypto{Pubkey: pubKey, Source: r.URL.Path}.doDecrypt()
		r.URL.Path = videoURL

		fmt.Printf("---> In static handler \n path %-11s \n key %-11s \n "+
			"url %-11s \n ", r.URL.Path, pubKey, videoURL)

		h.ServeHTTP(w, r)
	})
}

func retrivePublicKey(r *http.Request) (string, error) {
	session, err := store.Get(r, "user-details")
	if err != nil {
		return "", err
	}
	pubKey := session.Values["pubkey"]
	if pubKey == nil {
		return "", fmt.Errorf("public key not found")
	}
	return pubKey.(string), nil
}
