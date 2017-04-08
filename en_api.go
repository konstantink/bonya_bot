package main

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
	EnAddress = "http://demo.en.cx/%s"

	// LoginEndpoint login endpoint
	LoginEndpoint = "login/signin?json=1"

	// LevelInfoEndpoint level information endpoint
	LevelInfoEndpoint = "GameEngines/Encounter/Play/%d?json=1"
	SendCodeEndpoint
	// SendBonusCodeEndpoint
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

// EnAPIAuthResponse struct that represents the response from EN serverFault
// after authorisation.
// `Ok` - indicates the success of the authorisation request
// `Cookies` - cookies with session information, should be used for all further requests
// `Result` - json with the Result
// `StatusCode` - http status code
// `Description` - error string in case request was not successful
type EnAPIAuthResponse struct {
	Ok          bool
	Cookies     []*http.Cookie
	Result      json.RawMessage
	StatusCode  int
	Description string
}

func (apiResp *EnAPIAuthResponse) createFromResponse(resp *http.Response) error {
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

// NewAuthResponse creates new instance of EnAPIAuthResponse
func NewAuthResponse() *EnAPIAuthResponse {
	return &EnAPIAuthResponse{}
}

// EnAPI represents object that contains useful data to operate with
// EN server and information about current game state
type EnAPI struct {
	Username      string       `json:"Login"`
	Password      string       `json:"Password"`
	Client        *http.Client `json:"-"`
	CurrentGameID int32        `json:"-"`
	CurrentLevel  *LevelInfo   `json:"-"`
	Levels        *list.List   `json:"-"`
}

func (en *EnAPI) makeRequest(endpoint string, payload interface{}, response enResponse) error {
	var (
		enURL = fmt.Sprintf(EnAddress, LoginEndpoint)
		buf   bytes.Buffer
	)

	if err := json.NewEncoder(&buf).Encode(payload); err != nil {
		return err
	}
	log.Printf("Payload: %s", &buf)
	resp, err := en.Client.Post(enURL, "application/json", &buf)
	if err != nil {
		fmt.Print("Exit 1")
		return err
	}

	response.createFromResponse(resp)

	return nil
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
func (en *EnAPI) Login() error {
	var (
		authResponse *EnAPIAuthResponse
		err          error
	)
	authResponse = NewAuthResponse()

	err = en.makeRequest(fmt.Sprintf(EnAddress, LoginEndpoint), en, authResponse)
	log.Println("Login response: ", authResponse, err)
	if err != nil {
		log.Print(err)
		return err
	}
	if !authResponse.Ok {
		log.Printf("Failed to login to server: %s", authResponse.Description)
		return errors.New(authResponse.Description)
	}
	log.Printf("Successfully logged in on server %q as user %q", EnAddress, en.Username)
	return err
}

// GetLevelInfo returns pointer to the LevelResponse object
// with level information or empty object and the occurred error
func (en *EnAPI) GetLevelInfo() (*LevelResponse, error) {
	//gameUrl := "http://demo.en.cx/GameEngines/Encounter/Play/25733?json=1"
	var (
		gameURL = fmt.Sprintf(EnAddress, fmt.Sprintf(LevelInfoEndpoint, en.CurrentGameID))
		lvl     *LevelResponse
		err     error
	)

	resp, err := en.Client.Get(gameURL)
	if err != nil {
		log.Println("Error on GET request:", err)
		return &LevelResponse{}, err
	}

	if strings.HasPrefix(resp.Header["Content-Type"][0], "text/html") {
		log.Println("Incorrect cookies, need to re-login")
		return &LevelResponse{}, errors.New("Incorrect cookies, need to re-login")
	}

	lvl, err = parseLevelJSON(resp.Body)
	if lvl.Level == nil {
		return lvl, errors.New("No level info")
	}

	return lvl, err
}

type sendCodeResponse struct {
}

// SendCode sends post request to EN server, returns level information
// or error
func (en *EnAPI) SendCode(code string) (*LevelResponse, error) {
	var (
		codeURL  = fmt.Sprintf(EnAddress, fmt.Sprintf(SendCodeEndpoint, en.CurrentGameID))
		resp     *http.Response
		body     SendCodeRequest
		lvl      *LevelResponse
		bodyJSON []byte
		err      error
	)
	fmt.Println(codeURL)
	body = SendCodeRequest{
		codeRequest: codeRequest{
			LevelId:     en.CurrentLevel.LevelId,
			LevelNumber: en.CurrentLevel.Number},
		LevelAction: code,
	}
	bodyJSON, err = json.Marshal(body)
	if err != nil {
		log.Println("Error while serializing body:", err)
		return nil, err
	}

	resp, err = en.Client.Post(codeURL, "application/json", bytes.NewBuffer(bodyJSON))
	if err != nil {
		log.Println("Error while preforming request:", err)
		return nil, err
	}

	lvl, err = parseLevelJSON(resp.Body)
	return lvl, err
}

// SendBonusCode sends post request with bonus code to EN server,
// returns level information or error
func (en *EnAPI) SendBonusCode() {

}
