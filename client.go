package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"text/template"
)

type BClient struct {
	ConsumerKey    string
	ConsumerSecret string
	Repo           string
	Username       string
	accessToken    *accessToken
}

func main() {
	client := BClient{
		ConsumerKey:    "MSxtPGXknnpg9BEkpG",
		ConsumerSecret: "wth45aqgcEVD3JgTxHCrJqucwUF9KXEL",
		Repo:           "funny",
		Username:       "tszpinda"}
	client.Authenticate()

	list := client.PullRequests()

	for _, v := range list.PullRequests {
		fmt.Printf("\n%+v\n", v)
	}
}

func (c *BClient) Authenticate() {
	accessToken, err := createAccessToken(c)
	if err != nil {
		log.Printf("%+v", err)
		panic("\nAuthentication failure")
	}
	c.accessToken = accessToken
}

func (c *BClient) PullRequests() *PullRequests {
	data := PullRequests{}
	c.get("https://api.bitbucket.org/2.0/repositories/{{.Username}}/{{.Repo}}/pullrequests", &data)
	return &data
}

func (c *BClient) get(url string, data interface{}) error {
	tmpl, err := template.New("url").Parse(url)
	if err != nil {
		panic(err)
	}

	var doc bytes.Buffer
	err = tmpl.Execute(&doc, c)
	if err != nil {
		panic(err)
	}

	rUrl := doc.String()
	req, _ := http.NewRequest("GET", rUrl, nil)

	signDataRequest(req, c.accessToken, c.ConsumerKey, c.ConsumerSecret)

	// make the request
	resp, err := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	if err != nil {
		log.Fatal(err)
	}

	err = json.NewDecoder(resp.Body).Decode(data)
	return err
}

type PullRequests struct {
	Pagelen      int           `json:"pagelen"`
	Page         int           `json:"page"`
	Size         int           `json:"size"`
	PullRequests []PullRequest `json:"values"`
}

type PullRequest struct {
	Description       string      `json:"description"`
	Title             string      `json:"title"`
	Links             Links       `json:"links"`
	CloseSourceBranch bool        `json:"close_source_branch"`
	MergeCommit       string      `json:"merge_commit"`
	Reason            string      `json:"reason"`
	ClosedBy          string      `json:"closed_by"`
	Source            Source      `json:"source"`
	State             string      `json:"state"`
	Author            Author      `json:"author"`
	CreatedOn         string      `json:"created_on"`
	UpdatedOn         string      `json:"updated_on"`
	Type              string      `json:"type"`
	Id                string      `json:"id"`
	Destination       Destination `json:"destination"`
}

type Links struct {
	Decline  Link `json:"decline"`
	Commits  Link `json:"commits"`
	Self     Link `json:"self"`
	Comments Link `json:"comments"`
	Merge    Link `json:"merge"`
	Html     Link `json:"html"`
	Activity Link `json:"activity"`
	Diff     Link `json:"diff"`
	Approve  Link `json:"approve"`
}

type Link struct {
	Href string `json:"href"`
}
type Author struct {
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Type        string `json:"type"`
	UUID        string `json:"uuid"`
}

type Destination struct {
}

type Source struct {
}
