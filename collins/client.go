package collins

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

type Config struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Host     string `json:"host"`
}

type Client struct {
	config     Config
	httpClient http.Client
}

func NewFromJson(file string) (*Client, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var c Config
	err = json.NewDecoder(f).Decode(&c)
	if err != nil {
		return nil, err
	}
	return &Client{config: c, httpClient: http.Client{}}, nil
}

func New(username string, password string, host string) *Client {
	c := Config{
		Username: username,
		Password: password,
		Host:     host,
	}
	return &Client{config: c, httpClient: http.Client{}}
}

func (c *Client) doGet(route string) (*json.RawMessage, error) {
	url := fmt.Sprintf("%s%s", c.config.Host, route)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.config.Username, c.config.Password)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("User-Agent", "byxorna/collinsbot")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	// parse it into a GenericResponse and hand it back to the client to decode as it sees fit
	var collinsResp GenericResponse
	err = json.Unmarshal(body, &collinsResp)
	if err != nil {
		return nil, err
	}
	if collinsResp.Status != "success:ok" || resp.StatusCode != http.StatusOK {
		// this thing is an error response, so lets unmarshal it so we can pull the error message out
		var errResp ErrorResponse
		err = json.Unmarshal(body, &errResp)
		if err != nil {
			return &collinsResp.Data, err
		}
		return &collinsResp.Data, errResp.Error()
	}
	// just return the json.RawMessage so the caller can decode
	return &collinsResp.Data, nil
}

func (c *Client) Get(tag string) (*Asset, error) {
	rawjson, err := c.doGet(fmt.Sprintf("/api/asset/%s", tag))
	if err != nil {
		return nil, err
	}
	//try and parse the Data as an Asset
	var a Asset
	err = json.Unmarshal(*rawjson, &a)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (c *Client) Link(asset Asset) string {
	return c.LinkFromTag(asset.Asset.Tag)
}
func (c *Client) LinkFromTag(tag string) string {
	return fmt.Sprintf("%s/asset/%s", c.config.Host, tag)
}
func (c *Client) LinkFromAttribute(attribute string, value string) string {
	return fmt.Sprintf("%s/resources?%s=%s", c.config.Host, attribute, value)
}
