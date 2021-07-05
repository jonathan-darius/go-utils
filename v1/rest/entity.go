package rest

// LogRequest types
type LogRequest struct {
	Method          string
	URL             interface{}
	Header          interface{}
	Body            map[string]interface{}
	Host            string
	Form            interface{}
	PostForm        interface{}
	MultipartForm   interface{}
	RemoteAddr      string
	PublicIPAddress string
	RequestURI      string
}

// LogData types
type LogData struct {
	Request  LogRequest `json:"request"`
	Response Response   `json:"response"`
}