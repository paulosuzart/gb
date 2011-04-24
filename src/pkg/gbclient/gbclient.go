package gbclient


import (
	"http"
	"os"
	"log"
)


type HTTPClient struct {
	Addr, Method, User, Password string
	basicAuth                    bool
	client                       *http.Client
}

func NewHTTPClient(addr, method string) (c *HTTPClient) {
	c = &HTTPClient{
		Addr:      addr,
		Method:    method,
		User:      "",
		Password:  "",
		basicAuth: false,
		client:    new(http.Client),
	}
	return
}

func (c *HTTPClient) Auth(usr, passwd string) {
	c.User = usr
	c.Password = passwd
	c.basicAuth = true
}


func (c *HTTPClient) DoRequest() (response *http.Response, err os.Error) {

	response, _, err = c.client.Get(c.Addr)

	if err != nil {
		log.Print(err.String())
		return nil, err
	}

	if response.StatusCode == http.StatusUnauthorized {
		var req *http.Request = new(http.Request)
		var h http.Header = map[string][]string{}
		h.Add("Authorization", "Basic dGVzdGU6b3BlbiB0ZXN0ZQ==")
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
