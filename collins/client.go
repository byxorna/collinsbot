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

func (c *Client) doGet(route string) (*http.Response, error) {
	url := fmt.Sprintf("%s%s", c.config.Host, route)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.config.Username, c.config.Password)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("User-Agent", "byxorna/collinsbot")
	return c.httpClient.Do(req)
}

//TODO pull out this logic into something reusable
func (c *Client) Get(tag string) (*Asset, error) {
	resp, err := c.doGet(fmt.Sprintf("/api/asset/%s", tag))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Got response %d from %s: %s", resp.StatusCode, resp.Request.URL, body)
	}
	var collinsResp GenericResponse
	err = json.Unmarshal(body, &collinsResp)
	if err != nil {
		return nil, err
	}
	if collinsResp.Status != "success:ok" {
		return nil, fmt.Errorf("Got bad response from Collins: %s", collinsResp.Status)
	}

	//try and parse the Data as an Asset
	var a Asset
	err = json.Unmarshal(collinsResp.Data, &a)
	if err != nil {
		return nil, err
	}
	return &a, nil
}
