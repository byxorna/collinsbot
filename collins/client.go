package collins

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
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

func (c *Client) doGet(url string) (*json.RawMessage, error) {
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
	//log.Printf("%s\n", body)
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
	rawjson, err := c.doGet(fmt.Sprintf("%s/api/asset/%s", c.config.Host, tag))
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

func (c *Client) Find(params map[string]string, attrs map[string]string) ([]Asset, error) {
	//TODO add pagination
	/*
		var (
			page = 0
			size = 25
			sort = "DESC"
		)
	*/

	route := c.LinkFromAttributesApi(params, attrs)
	log.Printf("fetching %s\n", route)
	rawjson, err := c.doGet(route)
	if err != nil {
		return nil, err
	}
	var res PagedAssetResponse
	// data is a {"Pagination":...,"Data":...}
	err = json.Unmarshal(*rawjson, &res)
	if err != nil {
		return nil, err
	}
	//TODO loop pages

	return res.Data, nil

}

func (c *Client) Link(asset Asset) string {
	return c.LinkFromTag(asset.Asset.Tag)
}
func (c *Client) LinkFromTag(tag string) string {
	return fmt.Sprintf("%s/asset/%s", c.config.Host, tag)
}

// return a link to the web UI for a single attribute key=value pair
func (c *Client) LinkFromAttributeWeb(attribute string, value string) string {
	return fmt.Sprintf("%s/resources?%s=%s", c.config.Host, url.QueryEscape(attribute), url.QueryEscape(value))
}

// return a link to a given query in the API
func (c *Client) LinkFromAttributesApi(params map[string]string, attrs map[string]string) string {
	return c.linkFromAttributes("/api/assets", params, attrs)
}

// return a link to a given query in the web UI
func (c *Client) LinkFromAttributesWeb(params map[string]string, attrs map[string]string) string {
	//We shouldnt treat attrs and params differently, as there is no attribute=KEY;VALUE on /resources
	for k, v := range attrs {
		params[k] = v
	}
	return c.linkFromAttributes("/resources", params, map[string]string{})
}
func (c *Client) linkFromAttributes(route string, params map[string]string, attrs map[string]string) string {
	queryParams := make([]string, len(params)+len(attrs))
	i := 0
	for k, v := range params {
		queryParams[i] = fmt.Sprintf("%s=%s", url.QueryEscape(k), url.QueryEscape(v))
		i = i + 1
	}
	for k, v := range attrs {
		queryParams[i] = fmt.Sprintf("attribute=%s", url.QueryEscape(fmt.Sprintf("%s;%s", k, v)))
		i = i + 1
	}
	return fmt.Sprintf("%s%s?%s", c.config.Host, route, strings.Join(queryParams, "&"))
}
