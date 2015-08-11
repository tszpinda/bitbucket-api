package bapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"text/template"
	//"io/ioutil"
	"fmt"
	"net/http/httputil"
)

const (
	debugOn = false
)

type BClient struct {
	ConsumerKey    string
	ConsumerSecret string
	Repo           string
	Username       string
	accessToken    *accessToken
}

func (c *BClient) Authenticate() {
	accessToken, err := createAccessToken(c)
	if err != nil {
		log.Printf("%+v", err)
		panic("\nAuthentication failure")
	}
	c.accessToken = accessToken
}

func (c *BClient) PullRequests(status PullRequestStatus) *PullRequestList {
	data := PullRequestList{}
	urlParams := c.defaultUrlParams()
	urlParams["Status"] = status.String()
	if status != PullRequestAll {
		urlParams["FilterByStatus"] = true
	}

	c.get("https://api.bitbucket.org/2.0/repositories/{{.Username}}/{{.Repo}}/pullrequests{{if .FilterByStatus}}?state={{.Status}}{{end}}", &data, urlParams)

	return &data
}

func (c *BClient) PullRequest(id int) *PullRequest {
	data := PullRequest{}
	urlParams := c.defaultUrlParams()
	urlParams["Id"] = strconv.Itoa(id)
	c.get("https://api.bitbucket.org/2.0/repositories/{{.Username}}/{{.Repo}}/pullrequests/{{.Id}}", &data, &urlParams)

	return &data
}

func (c *BClient) defaultUrlParams() (params map[string]interface{}) {

	urlParams := make(map[string]interface{})
	urlParams["Username"] = c.Username
	urlParams["Repo"] = c.Repo
	return urlParams
}

func getUrl(url string, urlData interface{}) string {
	tmpl, err := template.New("url").Parse(url)
	if err != nil {
		panic(err)
	}

	var doc bytes.Buffer
	err = tmpl.Execute(&doc, urlData)
	if err != nil {
		panic(err)
	}
	return doc.String()
}

func (c *BClient) get(url string, data, urlData interface{}) error {
	rUrl := getUrl(url, urlData)
	req, _ := http.NewRequest("GET", rUrl, nil)

	signDataRequest(req, c.accessToken, c.ConsumerKey, c.ConsumerSecret)
	if debugOn {
		debug(httputil.DumpRequestOut(req, true))
	}

	resp, err := http.DefaultClient.Do(req)
	defer resp.Body.Close()

	if debugOn {
		debug(httputil.DumpResponse(resp, true))
	}

	log.Println("response code:", resp.StatusCode)
	if resp.StatusCode != 200 {
		log.Println(resp.Status)
		return errors.New("Http Status != 200")
	}

	if err != nil {
		return err
	}
	return json.NewDecoder(resp.Body).Decode(data)
}

func debug(data []byte, err error) {
	if err == nil {
		fmt.Printf("%s\n\n", data)
	} else {
		log.Fatalf("%s\n\n", err)
	}
}
