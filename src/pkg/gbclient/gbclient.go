//Groups http client related abstractions
package gbclient


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
	Addr, Method, User, Password string
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
		Addr:      addr,
		Method:    m,
		User:      "",
		Password:  "",
		basicAuth: false,
		client:    new(http.Client),
	}
	return
}

//Used to set auth information for HTTP Basic Authentication.
func (c *HTTPClient) Auth(usr, passwd string) {
	c.User = usr
	c.Password = passwd
	c.basicAuth = true
}

type Error string
func (e Error) String() string {
	return string(e)
}
//Perform the HTTP method against the target host.
//Auth is handled if Auth was previously invoked to set
//user info.
func (c *HTTPClient) DoRequest() (response *http.Response, err os.Error) {
	defer func() {
		if e := recover(); e != nil {
			response = nil
			err = e.(Error)
			log.Panic(err)
		}
	}()
	
	
	response, _, err = c.client.Get(c.Addr)

	if err != nil {
		log.Print(err.String())
		return nil, err
	}

	log.Print(c.basicAuth)
	if response.StatusCode == http.StatusUnauthorized && c.basicAuth {
		var req *http.Request = new(http.Request)
		var h http.Header = map[string][]string{}

		srcs := []byte(c.User + ": " + c.Password)
		dsts := make([]byte, base64.StdEncoding.EncodedLen(len(srcs)))
		
		base64.StdEncoding.Encode(dsts, srcs)

		h.Add("Authorization", "Basic " + string(dsts))

		req.Header = h
		req.Method = "GET"
		req.ProtoMajor = 1
		req.ProtoMinor = 1
		req.URL, _ = http.ParseURL(c.Addr)

		_, err = c.client.Do(req)

		if err != nil {
			log.Println("ERR")
			return
		}

	}
	return
}
