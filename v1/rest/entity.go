package rest

// ActivityRequest types
type ActivityRequest struct {
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

// APIActivity types
type APIActivity struct {
	Request  ActivityRequest `json:"request"`
	Response Response        `json:"response"`
}
