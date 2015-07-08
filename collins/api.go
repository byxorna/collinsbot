package collins

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

type Config struct {
	username string
	password string
	host     string
}

type Api struct {
	config Config
	client http.Client
}

func NewFromJson(file string) (*Api, error) {
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
	return &Api{config: c, client: http.Client{}}, nil
}

func New(username string, password string, host string) *Api {
	c := Config{
		username: username,
		password: password,
		host:     host,
	}
	return &Api{config: c, client: http.Client{}}
}

func (c *Api) doGet(route string) (*http.Response, error) {
	url := fmt.Sprintf("%s%s", c.config.host, route)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.config.username, c.config.password)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("User-Agent", "byxorna/collinsbot")
	return c.client.Do(req)
}

//TODO pull out this logic into something reusable
func (c *Api) Get(tag string) (*Asset, error) {
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
