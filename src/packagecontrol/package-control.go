package packagecontrol

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sync"

	"github.com/moul/http2curl"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

const (
	searchURL  = "https://packagecontrol.io/search/"
	packageURL = "https://packagecontrol.io/packages/"
)

type Packages struct {
	Packages []Package `json:"packages"`
}

type Package struct {
	Name                   string `json:"name"`
	HighlightedDescription string `json:"highlighted_description"`
	Installs               int    `json:"unique_installs"`
}

func (p *Package) GetName() string {
	return StripNewlines(p.Name)
}

func (p *Package) GetInstalls() int {
	return p.Installs
}

func (p *Package) FormattedInstalls() string {
	pprint := message.NewPrinter(language.English)
	return pprint.Sprintf("%d", p.GetInstalls())
}

func (p *Package) GetURL() string {
	return fmt.Sprintf("https://packagecontrol.io/packages/%s", p.GetName())
}

type PackageDetails struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Homepage    string   `json:"homepage"`
	Installs    Installs `json:"installs"`
}

func (p *PackageDetails) GetName() string {
	return StripNewlines(p.Name)
}

func (p *PackageDetails) GetInstalls() int {
	return p.Installs.Total
}

func (p *PackageDetails) FormattedInstalls() string {
	pprint := message.NewPrinter(language.English)
	return pprint.Sprintf("%d", p.GetInstalls())
}

func (p *PackageDetails) GetURL() string {
	return StripNewlines(p.Homepage)
}

type Installs struct {
	Total   int `json:"total"`
	Windows int `json:"windows"`
	Osx     int `json:"osx"`
	Linux   int `json:"linux"`
}

// Client represents PackageControl HTTP client
type Client struct {
	m *sync.Mutex

	SearchURL  string
	PackageURL string
	Debugging  bool
	Client     *http.Client
}

// NewClient creates new PackageControl client
func NewClient(client *http.Client) *Client {
	if client == nil {
		client = http.DefaultClient
	}

	return &Client{
		SearchURL:  searchURL,
		PackageURL: packageURL,
		Client:     client,
		m:          &sync.Mutex{},
	}
}

// SetDebug toggles the debug mode.
func (c *Client) SetDebug(debug bool) {
	c.m.Lock()
	defer c.m.Unlock()

	c.Debugging = debug
}

// Debug debug logger
func (c *Client) Debug(message string, req *http.Request, err error) {
	if c.Debugging {
		if req != nil {
			if command, err := http2curl.GetCurlCommand(req); err == nil {
				log.Printf("[ DEBUG ] %v\n", message)
				log.Printf("[ DEBUG ] %v\n", command)
			}
		} else {
			fmt.Printf("[ DEBUG ] %v\n", message)
		}
		if err != nil {
			fmt.Printf("[ ERROR ] %v\n", err)
		}
	}
}

// NewSearchRequest prepares new request
func (c *Client) NewSearchRequest(method string, search string) (*http.Request, error) {
	u, err := url.Parse(c.SearchURL + search + ".json")
	if err != nil {
		return nil, err
	}

	return c.baseRequest(method, u)
}

// NewPackageRequest prepares new request
func (c *Client) NewPackageRequest(method string, sublimePackage string) (*http.Request, error) {
	u, err := url.Parse(c.PackageURL + sublimePackage + ".json")
	if err != nil {
		return nil, err
	}

	return c.baseRequest(method, u)
}

func (c *Client) baseRequest(method string, u *url.URL) (*http.Request, error) {
	req, err := http.NewRequest(method, u.String(), nil)
	if err != nil {
		c.Debug("Sending HTTP Request", req, nil)
		return nil, err
	}

	return req, nil
}

// Do executes common (non-streaming) request.
func (c *Client) Do(ctx context.Context, req *http.Request, destination interface{}) error {
	if ctx != nil {
		req = req.WithContext(ctx)
	}

	c.Debug("Sending HTTP Request", req, nil)

	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	if destination == nil {
		return nil
	}

	return c.parseResponse(destination, resp.Body)
}

func (c *Client) parseResponse(destination interface{}, body io.Reader) error {
	var err error

	if w, ok := destination.(io.Writer); ok {
		_, err = io.Copy(w, body)
	} else {
		decoder := json.NewDecoder(body)
		err = decoder.Decode(destination)
	}

	return err
}
