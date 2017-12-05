package main

type page struct {
	WsEndPoint string
	VideoURL   string
	Name       string
}

type FileInfo struct {
	ID          int
	Name        string
	IsDirectory bool
	Path        string
}

type GroupedFileInfo struct {
	ID          int
	Name        string
	IsDirectory bool
	Path        string
	Childs      []FileInfo
}

// Credentials which stores google ids.
type Credentials struct {
	Cid     string `json:"cid"`
	Csecret string `json:"csecret"`
}

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
