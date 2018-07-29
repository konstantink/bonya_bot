package en

import (
	"bytes"
	"container/list"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

const (
	//EnAddress = "http://quest.ua/%s"
	// TODO: initially should be read from environment
	// EnAddress server address
	EnAddress = "http://%s/%s"

	// LoginEndpoint login endpoint
	LoginEndpoint = "login/signin?json=1"

	// LevelInfoEndpoint level information endpoint
	LevelInfoEndpoint = "GameEngines/Encounter/Play/%d?json=1"

	// SendCodeEndpoint send code endpoint
	SendCodeEndpoint
	// SendBonusCodeEndpoint

	DEBUG = true
)

const (
	captcha = iota + 1
	incorrectLogin
	incorrectUser
	ipBlacklisted
	serverFault
	bruteForce
)

// ServerError map with possible login errors from EN server
var ServerError = map[int32]string{
	captcha:        "Captcha input is required",
	incorrectLogin: "Incorrect login/password",
	incorrectUser:  "Incorrect user",
	ipBlacklisted:  "IP is blaclisted",
	serverFault:    "Server fault",
	bruteForce:     "Brootforce?",
}

type enResponse interface {
	createFromResponse(resp *http.Response) error
}

// APIAuthResponse struct that represents the response from EN serverFault
// after authorisation.
// `Ok` - indicates the success of the authorisation request
// `Cookies` - cookies with session information, should be used for all further requests
// `Result` - json with the Result
// `StatusCode` - http status code
// `Description` - error string in case request was not successful
type APIAuthResponse struct {
	Ok          bool
	Cookies     []*http.Cookie
	Result      json.RawMessage
	StatusCode  int
	Description string
}

func (apiResp *APIAuthResponse) createFromResponse(resp *http.Response) error {
	var (
		buf      []byte
		err      error
		respBody map[string]interface{}
	)
	buf, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	err = json.Unmarshal(buf, &respBody)
	if err != nil {
		return err
	}

	apiResp.Ok = respBody["Error"].(float64) == 0
	if !apiResp.Ok {
		apiResp.Description = ServerError[int32(respBody["Error"].(float64))]
	} else {
		apiResp.Description = ""
	}
	apiResp.Result = buf
	apiResp.Cookies = resp.Cookies()
	apiResp.StatusCode = resp.StatusCode

	return nil
}

// newAuthResponse creates new instance of APIAuthResponse from the http.Response
// that comes from the EN server, or empty instance if nil is passed
func newAuthResponse(response *http.Response) *APIAuthResponse {
	var authResponse = &APIAuthResponse{}

	if response == nil {
		return authResponse
	}

	if err := authResponse.createFromResponse(response); err != nil {
		log.Printf("Problem while creating auth response object: %q", err)
	}
	return authResponse
}

// API represents object that contains useful data to operate with
// EN server and information about current game state
type API struct {
	Username      string       `json:"Login"`
	Password      string       `json:"Password"`
	Client        *http.Client `json:"-"`
	CurrentGameID int32        `json:"-"`
	CurrentLevel  *Level       `json:"-"`
	Domain        string       `json:"-"`
	Levels        *list.List   `json:"-"`
}

func (api *API) makeRequest2(method string, url string, body io.Reader) (*http.Response, error) {
	request, err := http.NewRequest(method, url, body)
	request.Header.Add("Content-Type", "application/json")
	if err != nil {
		log.Printf("Failed to create request: %s", err)
		return nil, err
	}

	if DEBUG {
		log.Printf("[DEBUG] Sending request to %s", url)
	}
	response, err := api.Client.Do(request)
	if err != nil {
		fmt.Printf("Failed to post request: %q", err)
		return nil, err
	}

	return response, nil
}

func (api *API) verifyAuthResponse(response *http.Response) error {
	var (
		buf      []byte
		err      error
		ok       bool
		respBody map[string]interface{}
	)
	buf, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	err = json.Unmarshal(buf, &respBody)
	if err != nil {
		return err
	}

	ok = respBody["Error"].(float64) == 0
	if !ok {
		return errors.New(ServerError[int32(respBody["Error"].(float64))])
	}
	api.Client.Jar.SetCookies(response.Request.URL, response.Cookies())

	// apiResp.Result = buf
	// apiResp.Cookies = resp.Cookies()
	// apiResp.StatusCode = resp.StatusCode

	return nil
}

// Login2 new version of Login function
func (api *API) Login2(username string, password string) error {
	var (
		body = bytes.NewBufferString("")
		url  = fmt.Sprintf(EnAddress, api.Domain, LoginEndpoint)
	)
	if err := json.NewEncoder(body).Encode(map[string]string{
		"Login":    username,
		"Password": password,
	}); err != nil {
		return err
	}

	resp, err := api.makeRequest2("POST", url, body)
	if err != nil {
		return err
	}

	if err := api.verifyAuthResponse(resp); err != nil {
		log.Printf("[ERROR] Failed to login to %s: %s", api.Domain, err)
		return err
	}

	log.Printf("[INFO] Successfully logged in to %s with user %s", api.Domain, username)
	// auth := newAuthResponse(response)
	return nil
}

func (api *API) makeRequest(url string, payload interface{}) (*http.Response, error) {
	var buf bytes.Buffer

	if err := json.NewEncoder(&buf).Encode(payload); err != nil {
		return nil, err
	}
	log.Printf("Payload: %s", &buf)
	resp, err := api.Client.Post(url, "application/json", &buf)
	if err != nil {
		fmt.Printf("Failed to post request: %q", err)
		return nil, err
	}

	return resp, nil
}

func parseLevelJSON(body io.ReadCloser) (*LevelResponse, error) {
	var (
		lvl *LevelResponse
		err error
	)
	defer body.Close()
	respBody, _ := ioutil.ReadAll(body)
	lvl = new(LevelResponse)
	err = json.Unmarshal(respBody, &lvl)
	if err != nil {
		log.Println("Error:", err)
		return &LevelResponse{}, err
	}

	return lvl, nil
}

// Login returns an error in case login to the EN server failed,
// otherwise returns nil
func (api *API) Login() error {
	var (
		authResponse *APIAuthResponse
		resp         *http.Response
		err          error
	)

	resp, err = api.makeRequest(fmt.Sprintf(EnAddress, api.Domain, LoginEndpoint), api)
	authResponse = newAuthResponse(resp)
	log.Println("Login response: ", authResponse, err)
	if err != nil {
		log.Print(err)
		return err
	}
	if !authResponse.Ok {
		log.Printf("Failed to login to server: %s", authResponse.Description)
		return errors.New(authResponse.Description)
	}
	log.Printf("Successfully logged in on server %q as user %q", api.Domain, api.Username)
	return err
}

// GetLevelInfo returns pointer to the LevelResponse object
// with level information or empty object and the occurred error
func (api *API) GetLevelInfo() (*Level, error) {
	//gameUrl := "http://demo.en.cx/GameEngines/Encounter/Play/25733?json=1"
	var (
		gameURL = fmt.Sprintf(EnAddress, api.Domain, fmt.Sprintf(LevelInfoEndpoint, api.CurrentGameID))
		lvl     *Level
		err     error
	)

	request, err := http.NewRequest("GET", gameURL, nil)
	// url, err := url.Parse(gameURL)
	// for _, cookie := range api.Client.Jar.Cookies(url) {
	// 	request.AddCookie(cookie)
	// }
	// resp, err := api.Client.Get(gameURL)
	resp, err := api.Client.Do(request)
	if err != nil {
		log.Println("Error on GET request:", err)
		return NewLevel(nil), err
	}

	if strings.HasPrefix(resp.Header["Content-Type"][0], "text/html") {
		if resp.StatusCode == 504 {
			log.Println("Timeout on server")
		} else {
			// log.Println("Incorrect cookies, need to re-login")
			buf, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()
			log.Printf("%s", buf)
		}
		return NewLevel(nil), errors.New("Incorrect cookies, need to re-login")
	}
	// buf, err := ioutil.ReadAll(resp.Body)
	// if err != nil {
	// 	return nil, err
	// }
	// defer resp.Body.Close()
	// log.Printf("===> Body: %s", buf)

	lvl = NewLevel(resp)
	//lvl, err = parseLevelJSON(resp.Body)
	if lvl == nil {
		return lvl, errors.New("No level info")
	}

	return lvl, err
}

type sendCodeResponse struct {
}

// SendCode sends post request to EN server, returns level information
// or error
func (api *API) SendCode(code string) (*Level, error) {
	var (
		codeURL = fmt.Sprintf(EnAddress, api.Domain, fmt.Sprintf(SendCodeEndpoint, api.CurrentGameID))
		resp    *http.Response
		body    SendCodeRequest
		lvl     *Level
		//bodyJSON []byte
		err error
	)

	body = SendCodeRequest{
		codeRequest: codeRequest{
			LevelID:     api.CurrentLevel.LevelID,
			LevelNumber: api.CurrentLevel.Number},
		LevelAction: code,
	}

	resp, err = api.makeRequest(codeURL, body)
	//bodyJSON, err = json.Marshal(body)
	if err != nil {
		log.Println("Error while serializing body:", err)
		return nil, err
	}
	//
	//resp, err = en.Client.Post(codeURL, "application/json", bytes.NewBuffer(bodyJSON))
	//if err != nil {
	//	log.Println("Error while preforming request:", err)
	//	return nil, err
	//}

	lvl = NewLevel(resp)
	//lvl, err = parseLevelJSON(resp.Body)
	return lvl, err
}

// SendBonusCode sends post request with bonus code to EN server,
// returns level information or error
func (api *API) SendBonusCode() {

}
