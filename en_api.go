package main

import (
	"net/http"
	"container/list"
	"fmt"
	"encoding/json"
	"log"
	"bytes"
	"io"
	"io/ioutil"
	"net/url"
	"strings"
	"errors"
)

const (
	//EnAddress = "http://quest.ua/%s"
	EnAddress = "http://demo.en.cx/%s"
	LoginEndpoint = "login/signin?json=1"
	LevelInfoEndpoint = "GameEngines/Encounter/Play/%d?json=1"
	SendCodeEndpoint
	SendBonusCodeEndpoint
)

type EnAPIAuthResponse struct {
	Ok          bool
	Cookies     []*http.Cookie
	Result      json.RawMessage
	StatusCode  int
	Description string
}

func (apiResp *EnAPIAuthResponse) CreateFromResponse(resp *http.Response) error {
	var (
		bytes []byte
		err error
		respBody map[string]interface{}
	)
	bytes, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	err = json.Unmarshal(bytes, &respBody)
	if err != nil {
		return err
	}

	apiResp.Ok = respBody["Error"].(float64) == 0
	if !apiResp.Ok {
		apiResp.Description = respBody["Description"].(string)
	} else {
		apiResp.Description = ""
	}
	apiResp.Result = bytes
	apiResp.Cookies = resp.Cookies()
	apiResp.StatusCode = resp.StatusCode

	return nil
}

type EnAPI struct {
	login         string       `json:"Login"`
	password      string       `json:"Password"`
	Client        *http.Client `json:"-"`
	CurrentGameId int32        `json:"-"`
	CurrentLevel  *LevelInfo   `json:"-"`
	Levels        *list.List   `json:"-"`
}

func (en *EnAPI) MakeRequest(endpoint string, params url.Values) (EnAPIAuthResponse, error) {
	var enUrl string = fmt.Sprintf(EnAddress, LoginEndpoint)

	resp, err := en.Client.PostForm(enUrl, params)
	if err != nil {
		fmt.Print("Exit 1")
		return EnAPIAuthResponse{}, err
	}

	var apiResp EnAPIAuthResponse
	apiResp.CreateFromResponse(resp)

	return apiResp, nil
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

func (en *EnAPI) Login() (EnAPIAuthResponse, error) {
	var (
		authResponse EnAPIAuthResponse
		err error
		params url.Values
	)
	params = make(url.Values)
	params.Set("Login", en.login)
	params.Set("Password", en.password)

	authResponse, err = en.MakeRequest(fmt.Sprintf(EnAddress, LoginEndpoint), params)
	if err != nil {
		log.Print(err)
		return EnAPIAuthResponse{}, nil
	}
	log.Printf("Successfully logged in on server %q as user %q", EnAddress,en.login)
	return authResponse, err
}

func (en *EnAPI) GetLevelInfo() (*LevelResponse, error) {
	//gameUrl := "http://demo.en.cx/GameEngines/Encounter/Play/25733?json=1"
	var (
		gameUrl string = fmt.Sprintf(EnAddress, fmt.Sprintf(LevelInfoEndpoint, en.CurrentGameId))
		lvl *LevelResponse
		err error
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

	return lvl, err
}

type sendCodeResponse struct {

}

func (en *EnAPI) SendCode(gameId int32, code string) (*LevelResponse, error) {
	var (
		codeUrl string = fmt.Sprintf(EnAddress, fmt.Sprintf(SendCodeEndpoint, gameId))
		resp *http.Response
		body SendCodeRequest
		lvl *LevelResponse
		bodyJson []byte
		err error
	)
	body = SendCodeRequest{
		codeRequest: codeRequest{
			LevelId: 249435,
			LevelNumber: 3},
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

