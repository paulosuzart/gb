package main


import (
	"http"
	"os"
	"log"
	"encoding/base64"
)

const (
	GET          = "GET"
	POST         = "POST"
	DEFAULT_VERB = GET
)

type Cookie struct {
	Name, Value string
}

//Hold information about how to connect and
//authenticate to the server
type HTTPClient struct {
	addr, method, user, password string
	basicAuth                    bool
	contentType                  string
	cookie                       Cookie
}

//HTTPClient constructor. If method is "", DEFAULT_VERB is
//then used.
func NewHTTPClient(addr, contentType string, cookie Cookie) (c *HTTPClient) {
	var m string
	if contentType != "" {
		m = POST
	} else {
		m = DEFAULT_VERB
	}

	c = &HTTPClient{
		addr:        addr,
		method:      m,
		user:        "",
		password:    "",
		contentType: contentType,
		basicAuth:   false,
		cookie:      cookie,
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

func (self *HTTPClient) defaultRequest() (req *http.Request, err os.Error) {
	var h http.Header = map[string][]string{}
	req = new(http.Request)
	req.Method = self.method
	if self.contentType != "" {
		headers := map[string]string{"Content-Type": self.contentType}
		for k, v := range headers {
			h.Add(k, v)
		}
		req.Header = h
	}

	req.ProtoMajor = 1
	req.ProtoMinor = 1
	if self.cookie.Name != "" && self.cookie.Value != "" {
		req.Cookie = make([]*http.Cookie, 1)
		req.Cookie[0] = &http.Cookie{Name: self.cookie.Name, Value: self.cookie.Value}
	}

	if req.URL, err = http.ParseURL(self.addr); err != nil {
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

	req, err := c.defaultRequest()
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
