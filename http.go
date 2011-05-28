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
	contentType                  string
}

//HTTPClient constructor. If method is "", DEFAULT_VERB is
//then used.
func NewHTTPClient(addr, method, contentType string) (c *HTTPClient) {
	m := DEFAULT_VERB

	if method != "" {
		m = method
	}
	c = &HTTPClient{
		addr:        addr,
		method:      m,
		user:        "",
		password:    "",
		contentType: contentType,
		basicAuth:   false,
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

func defaultRequest(url string, headers map[string]string) (req *http.Request, err os.Error) {
	var h http.Header = map[string][]string{}
	req = new(http.Request)
	for k, v := range headers {
		h.Add(k, v)
	}
	req.Header = h
	req.ProtoMajor = 1
	req.ProtoMinor = 1
	if req.URL, err = http.ParseURL(url); err != nil {
		return
	}
	return

}

var gbTransport *http.Transport = &http.Transport{DisableKeepAlives: true}
//Perform the HTTP method against the target host.
//Auth is handled if Auth was previously invoked to set
//user info.
func (c *HTTPClient) DoRequest() (response *http.Response, err os.Error) {
	//Recover if things goes really bad
	defer func() {
		if e := recover(); e != nil {
			log.Print(e)
		}
	}()

	req, err := defaultRequest(c.addr, map[string]string{"Content-Type": c.contentType})
	if err != nil {
		return
	}
	response, err = gbTransport.RoundTrip(req)

	if err != nil {
		log.Printf("Error performing Request: %v", err.String())
		return nil, err
	}
	if response.StatusCode == http.StatusUnauthorized && c.basicAuth {

		req.Header.Add("Authorization", authInfo(c.user, c.password))

		_, err = gbTransport.RoundTrip(req) //c.client.Do(req)

		if err != nil {
			log.Println(err)
			return
		}

	}
	return
}
