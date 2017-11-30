package main

import (
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type videoListHandler struct {
}

func (vlh videoListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/index.html", "templates/list.html")
	if err != nil {
		http.Error(w, http.StatusText(400), http.StatusBadRequest)
		return
	}

	vlh.generateSaltForURLEncryption(w, r)

	videoFilesAndUserInfo := struct {
		VideoFiles []FileInfo
		UserInfo   User
	}{vlh.getVideoFilesInDirectory(), *vlh.sessionUser(r)}

	tmpl.ExecuteTemplate(w, "layout", videoFilesAndUserInfo)
}

func (vlh videoListHandler) sessionUser(r *http.Request) *User {
	session, _ := store.Get(r, "user-details")
	_user := session.Values["user"]
	if _user != nil {
		return _user.(*User)
	}
	return &User{
		Picture: "https://lh6.googleusercontent.com/-TUpg87ezNDw/AAAAAAAAAAI/AAAAAAAAACQ/6g6K__7LDaQ/photo.jpg",
		Email:   "vimlesh.sharma@synerzip.com",
	}
}

func (vlh videoListHandler) generateSaltForURLEncryption(w http.ResponseWriter, r *http.Request) error {
	session, _ := store.Get(r, "user-details")
	session.Values["pubkey"] = videoURLCrypto{}.randomString(SaltLen)
	err := session.Save(r, w)
	return err
}

func (vlh videoListHandler) getVideoFilesInDirectory() []FileInfo {
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

func (videoListHandler) destroySession(w http.ResponseWriter, r *http.Request) {
	//DELETE in session and set cookie to zero as well
	cookie := http.Cookie{
		Name:   "user-details",
		Value:  "",
		MaxAge: -1,
	}
	http.SetCookie(w, &cookie)
}
