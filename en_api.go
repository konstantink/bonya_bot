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
	EnAddress         = "http://demo.en.cx/%s"
	LoginEndpoint     = "login/signin?json=1"
	LevelInfoEndpoint = "GameEngines/Encounter/Play/%d?json=1"
	SendCodeEndpoint
	SendBonusCodeEndpoint
)

var ServerError = map[int32]string{
	1: "Captcha input is required",
	2: "Incorrect login/password",
	3: "Incorrect user",
	4: "IP is blaclisted",
	5: "Server fault",
	9: "Brootforce?",
}

type EnResponse interface {
	CreateFromResponse(resp *http.Response) error
}


type EnAPIAuthResponse struct {
	Ok          bool
	Cookies     []*http.Cookie
	Result      json.RawMessage
	StatusCode  int
	Description string
}

func (apiResp *EnAPIAuthResponse) CreateFromResponse(resp *http.Response) error {
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

func NewAuthResponse() EnAPIAuthResponse {
	return EnAPIAuthResponse{}
}

type EnAPI struct {
	login         string       `json:"Login"`
	password      string       `json:"Password"`
	Client        *http.Client `json:"-"`
	CurrentGameId int32        `json:"-"`
	CurrentLevel  *LevelInfo   `json:"-"`
	Levels        *list.List   `json:"-"`
}

func (en *EnAPI) MakeRequest(endpoint string, payload interface{}, response EnResponse)  error {
	var (
		enUrl string = fmt.Sprintf(EnAddress, LoginEndpoint)
		buf bytes.Buffer
	)

	if err := json.NewEncoder(&buf).Encode(payload); err != nil {
		return err
	}
	log.Printf("Payload: %s", &buf)
	resp, err := en.Client.Post(enUrl, "application/json", &buf)
	if err != nil {
		fmt.Print("Exit 1")
		return err
	}

	response.CreateFromResponse(resp)

	return nil
}

func parseLevelJson(body io.ReadCloser) (*LevelResponse, error) {
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

func (en *EnAPI) Login()  error {
	var (
		authResponse EnAPIAuthResponse
		err          error
		params       map[string]string
	)
	authResponse = NewAuthResponse()
	params = make(map[string]string)
	params["Login"] = en.login
	params["Password"] = en.password

	err = en.MakeRequest(fmt.Sprintf(EnAddress, LoginEndpoint), params, &authResponse)
	log.Println("Login response: ", authResponse, err)
	if err != nil {
		log.Print(err)
		return err
	}
	if !authResponse.Ok {
		log.Printf("Failed to login to server: %s", authResponse.Description)
		return errors.New(authResponse.Description)
	}
	log.Printf("Successfully logged in on server %q as user %q", EnAddress, en.login)
	return err
}

func (en *EnAPI) GetLevelInfo() (*LevelResponse, error) {
	//gameUrl := "http://demo.en.cx/GameEngines/Encounter/Play/25733?json=1"
	var (
		gameUrl string = fmt.Sprintf(EnAddress, fmt.Sprintf(LevelInfoEndpoint, en.CurrentGameId))
		lvl     *LevelResponse
		err     error
	)

	resp, err := en.Client.Get(gameUrl)
	if err != nil {
		log.Println("Error on GET request:", err)
		return &LevelResponse{}, err
	}

	if strings.HasPrefix(resp.Header["Content-Type"][0], "text/html") {
		log.Println("Incorrect cookies, need to re-login")
		return &LevelResponse{}, errors.New("Incorrect cookies, need to re-login")
	}

	lvl, err = parseLevelJson(resp.Body)
	if lvl.Level == nil {
		return lvl, errors.New("No level info")
	}

	return lvl, err
}

type sendCodeResponse struct {
}

func (en *EnAPI) SendCode(code string) (*LevelResponse, error) {
	var (
		codeUrl  string = fmt.Sprintf(EnAddress, fmt.Sprintf(SendCodeEndpoint, en.CurrentGameId))
		resp     *http.Response
		body     SendCodeRequest
		lvl      *LevelResponse
		bodyJson []byte
		err      error
	)
	fmt.Println(codeUrl)
	body = SendCodeRequest{
		codeRequest: codeRequest{
			LevelId:     en.CurrentLevel.LevelId,
			LevelNumber: en.CurrentLevel.Number},
		LevelAction: code,
	}
	bodyJson, err = json.Marshal(body)
	if err != nil {
		log.Println("Error while serializing body:", err)
		return nil, err
	}

	resp, err = en.Client.Post(codeUrl, "application/json", bytes.NewBuffer(bodyJson))
	if err != nil {
		log.Println("Error while preforming request:", err)
		return nil, err
	}

	lvl, err = parseLevelJson(resp.Body)
	return lvl, err
}

func (en *EnAPI) SendBonusCode() {

}
