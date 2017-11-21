package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

type page struct {
	WsEndPoint string
	VideoURL   string
	Name       string
}

type videoInfo struct {
	Name string
	Url  string
}

func decorate(h http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("-------------------------static handler----------------------")
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

	})
}

func main() {
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", decorate(fs)))
	fs1 := http.FileServer(http.Dir("resourses"))
	http.Handle("/resourses/", http.StripPrefix("/resourses/", fs1))

	http.HandleFunc("/list", videolistHandler)
	http.HandleFunc("/play", playHandler)
	http.HandleFunc("/me", videoplayHandler)

	err := http.ListenAndServe(":9090", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

}

func videolistHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/index.html", "templates/list.html")
	if err != nil {
		log.Println(err.Error())
		http.Error(w, http.StatusText(500), 500)
		return
	}

	files := getVideoFilesInDirectory()
	data := struct {
		VideoFiles []string
	}{files}

	//generate random key
	pubKey := randomString(6)
	fmt.Println("Pub key", pubKey)

	// set public key in cookie for decrypting names and play list
	expiration := time.Now().Add(60 * time.Minute)
	cookie := http.Cookie{
		Name:    "pubkey",
		Value:   pubKey,
		Expires: expiration,
	}
	http.SetCookie(w, &cookie)

	tmpl.ExecuteTemplate(w, "layout", data)
}

func playHandler(w http.ResponseWriter, r *http.Request) {
	file := r.URL.Query().Get("file")
	videopath := fmt.Sprintf("videos/%s", file)
	fmt.Println(videopath)

	var pubKey string
	c, _ := r.Cookie("pubkey")
	fmt.Printf(" decrypt value from cookie %v  \n", c.Value)

	if c == nil {
		pubKey = c.Value
	} else {
		pubKey = c.Value
	}

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

func getVideoFilesInDirectory() []string {
	videofiles := []string{}
	files, err := ioutil.ReadDir("./static/videos/")
	if err != nil {
		log.Println("error while reading files.. ", err.Error())
	}
	for _, file := range files {
		videofiles = append(videofiles, file.Name())
	}
	return videofiles
}

func videoplayHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("-----------------handler------------")
	pubKey := randomString(6)
	fmt.Println("Pub key", pubKey)

	// videosFromDirec := []videoInfo{}
	// for _, file := range getVideoFilesInDirectory() {
	// 	// fmt.Println(file)
	// 	fmt.Println(randomString(6))

	// 	videopath := fmt.Sprintf("videos/%s", file)
	// 	videoURL, _ := encrypt(CIPHER_KEY, videopath)
	// 	evideoURL := "/static/" + videoURL

	// 	vi := videoInfo{
	// 		Name: file,
	// 		Url:  evideoURL,
	// 	}
	// 	videosFromDirec = append(videosFromDirec, vi)
	// }

	videopath := "videos/weekly-checkin call -2017-09-15.mp4"
	// videoURL, err := encrypt(CIPHER_KEY, videopath)
	videoURL := doEncrypt(pubKey, videopath)
	evideoURL := "/static/" + videoURL

	/* 		if err != nil {
	   			log.Println(err)
	   		}
	*/
	p := page{
		WsEndPoint: r.Host,
		VideoURL:   evideoURL,
	}

	// for cookies
	expiration := time.Now().Add(10 * time.Minute)
	cookie := http.Cookie{
		Name:    "pubkey",
		Value:   pubKey,
		Expires: expiration,
	}
	http.SetCookie(w, &cookie)
	//cookies

	tmpl, err := template.ParseFiles("templates/index.html", "templates/input.html")
	if err != nil {
		log.Println(err.Error())
		http.Error(w, http.StatusText(500), 500)
		return
	}
	//experiment
	buf := bytes.Buffer{}
	tmpl.ExecuteTemplate(&buf, "layout", p)
	w.Write(buf.Bytes())

	// if err := tmpl.ExecuteTemplate(w, "layout", p); err != nil {
	// 	log.Println(err.Error())
	// 	http.Error(w, http.StatusText(500), 500)
	// }
}

func ConnWs(w http.ResponseWriter, r *http.Request) {
	ws, err := websocket.Upgrade(w, r, nil, 1024, 1024)
	if _, ok := err.(websocket.HandshakeError); ok {
		http.Error(w, "Not a websocket handshake", 400)
		return
	} else if err != nil {
		log.Println(err)
		return
	}

	// var img64 []byte

	res := map[string]interface{}{}
	for {
		// messageType, p, err := ws.ReadMessage()
		// fmt.Println(messageType)
		// fmt.Println(string(p))
		// fmt.Println(err)

		if err = ws.ReadJSON(&res); err != nil {
			fmt.Printf("%v \n", res)

			if err.Error() == "EOF" {
				return
			}
			// ErrShortWrite means a write accepted fewer bytes than requested then failed to return an explicit error.
			if err.Error() == "unexpected EOF" {
				return
			}
			fmt.Println("Read : " + err.Error())
			return
		}

		res["a"] = "a"
		log.Println(res)

		f, _ := os.Open("./videos/big.mp4")
		wswriter, _ := ws.NextWriter(websocket.BinaryMessage)
		io.Copy(wswriter, f)
		f.Close()
		wswriter.Close()

		// reader := bufio.NewReader(f)
		// buf := make([]byte, 1024*20)
		// byt, err := reader.ReadBytes(8000)
		// ws.WriteMessage(websocket.BinaryMessage, byt)

		// // for {
		// files, _ := ioutil.ReadDir("./videos")
		// for _, f := range files {

		// 	video, _ := ioutil.ReadFile("./videos/" + f.Name())
		// 	ws.WriteMessage(websocket.BinaryMessage, video)

		// 	/*Below code works*/
		// 	// str := base64.StdEncoding.EncodeToString(img64)
		// 	// res["img64"] = str

		// 	// if err = ws.WriteJSON(&res); err != nil {
		// 	// 	fmt.Println("watch dir - Write : " + err.Error())
		// 	// }

		// 	/*Efficient only for small data*/
		// 	/*TextMessage*/
		// 	/*Remember ruby multipart upload, always encoded into base64 */

		// 	// _dst := make([]byte, base64.StdEncoding.EncodedLen(len(img64)))
		// 	// base64.StdEncoding.Encode(_dst, img64)

		// 	/*For larger data use NewEncoder */
		// 	// _dst := &bytes.Buffer{}
		// 	// encoder := base64.NewEncoder(base64.StdEncoding, _dst)
		// 	// encoder.Write(img64)
		// 	// encoder.Close()

		// 	// ws.WriteMessage(websocket.TextMessage, _dst.Bytes())

		// 	/*BinaryMessage*/
		// 	/*utf-8 encoding*/
		// 	// ws.WriteMessage(websocket.BinaryMessage, img64)

		// 	time.Sleep(50 * time.Millisecond)
		// }

		time.Sleep(50 * time.Millisecond)
		// }
	}
}
