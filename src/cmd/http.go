package main


import (
	"http"
	"os"
	"log"
	"encoding/base64"
)

const (
	DEFAULT_VERB = "GET"
)


//Hold information about how to connect and
//authenticate to the server
type HTTPClient struct {
	addr, method, user, password string
	basicAuth                    bool
	client                       *http.Client
}

//HTTPClient constructor. If method is "", DEFAULT_VER is
//then used.
func NewHTTPClient(addr, method string) (c *HTTPClient) {
	m := DEFAULT_VERB

	if method != "" {
		m = method
	}
	c = &HTTPClient{
		addr:      addr,
		method:    m,
		user:      "",
		password:  "",
		basicAuth: false,
		client:    new(http.Client),
	}
	return
}

//Used to set auth information for HTTP Basic Authentication.
func (c *HTTPClient) Auth(usr, passwd string) {
	c.user = usr
	c.password = passwd
	c.basicAuth = true
}

//Uses base64 to encode "user: password" for
//Http Header Authorization.
func authInfo(user, password string) string {

	srcs := []byte(user + ": " + password)
	dsts := make([]byte, base64.StdEncoding.EncodedLen(len(srcs)))

	base64.StdEncoding.Encode(dsts, srcs)
	return string(dsts)
}

type Error string

func (e Error) String() string {
	return string(e)
}
//Perform the HTTP method against the target host.
//Auth is handled if Auth was previously invoked to set
//user info.
func (c *HTTPClient) DoRequest() (response *http.Response, err os.Error) {
	//Recover if things goes really bad
	defer func() {
		if e := recover(); e != nil {
			response = nil
			err = e.(Error)
			log.Print(err)
		}
	}()

	response, _, err = c.client.Get(c.addr)

	if err != nil {
		log.Printf("Error performing Request: %v", err.String())
		return nil, err
	}
	if response.StatusCode == http.StatusUnauthorized && c.basicAuth {
		var req *http.Request = new(http.Request)
		var h http.Header = map[string][]string{}

		h.Add("Authorization", authInfo(c.user, c.password))

		req.Header = h
		req.Method = c.method
		req.ProtoMajor = 1
		req.ProtoMinor = 1
		req.URL, _ = http.ParseURL(c.addr)

		_, err = c.client.Do(req)

		if err != nil {
			log.Println(err)
			return
		}

	}
	return
}
