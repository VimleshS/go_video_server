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
	user := vlh.sessionUser(r)
	/*User not present in session*/
	if user == nil {
		roothandler(w, r)
		return
	}

	tmpl, err := template.ParseFiles("templates/index.html", "templates/list.html")
	if err != nil {
		http.Error(w, http.StatusText(400), http.StatusBadRequest)
		return
	}

	vlh.generateSaltForURLEncryption(w, r)

	videoFilesAndUserInfo := struct {
		VideoFiles []GroupedFileInfo
		UserInfo   User
	}{vlh.groupByParent(), *user}

	tmpl.ExecuteTemplate(w, "layout", videoFilesAndUserInfo)
}

func (vlh videoListHandler) sessionUser(r *http.Request) *User {
	session, _ := store.Get(r, "user-details")
	_user := session.Values["user"]
	if _user != nil {
		return _user.(*User)
	}
	return nil
	/*
		return &User{
			Picture: "https://lh6.googleusercontent.com/-TUpg87ezNDw/AAAAAAAAAAI/AAAAAAAAACQ/6g6K__7LDaQ/photo.jpg",
			Email:   "vimlesh.sharma@synerzip.com",
		}
	*/
}

func (vlh videoListHandler) generateSaltForURLEncryption(w http.ResponseWriter, r *http.Request) error {
	session, _ := store.Get(r, "user-details")
	session.Values["pubkey"] = videoURLCrypto{}.randomString(SaltLen)
	err := session.Save(r, w)
	return err
}

func (vlh videoListHandler) groupByParent() []GroupedFileInfo {
	// updateRec("")

	// files := vlh.getVideoFilesInDirectory()
	files := vlh.VideoFilesFromDB()
	groupedVideos := []GroupedFileInfo{}

	accTitleInfo := GroupedFileInfo{}
	for _, file := range files {
		if file.IsDirectory {
			if file.ID > 1 {
				groupedVideos = append(groupedVideos, accTitleInfo)
				accTitleInfo = GroupedFileInfo{}
			}
			accTitleInfo.ID = file.ID
			accTitleInfo.IsDirectory = file.IsDirectory
			accTitleInfo.Name = file.Name
			accTitleInfo.Path = file.Path
			accTitleInfo.Desc = file.Desc
		} else {
			accTitleInfo.Childs = append(accTitleInfo.Childs, file)
		}
	}

	if len(accTitleInfo.Childs) > 0 {
		groupedVideos = append(groupedVideos, accTitleInfo)
	}

	return groupedVideos
}

func (vlh videoListHandler) getVideoFilesInDirectory() []FileInfo {
	fileList := []FileInfo{}
	id := 1
	filepath.Walk("./static/videos/", func(path string, info os.FileInfo, err error) error {
		path = strings.TrimPrefix(path, "static/videos/")
		namepart := strings.Split(path, "/")
		name := namepart[len(namepart)-1]
		if len(path) > 0 {
			fi := FileInfo{ID: id, Name: name, Path: path, IsDirectory: info.IsDir()}
			id++
			fileList = append(fileList, fi)
		}
		return nil
	})
	return fileList

}

func (vlh videoListHandler) VideoFilesFromDB() []FileInfo {
	return readvideoInfo()
}

func (videoListHandler) destroySession(w http.ResponseWriter, r *http.Request) {
	//DELETE in session and set cookie to zero as well
	cookie := http.Cookie{
		Name:   "user-details",
		Value:  "",
		MaxAge: -1,
	}
	http.SetCookie(w, &cookie)

	session, _ := store.Get(r, "user-details")
	session.Options.MaxAge = -1
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
