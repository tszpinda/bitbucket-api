package bapi

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

type requestToken struct {
	token             string
	secret            string
	callbackConfirmed bool
}

type accessToken struct {
	OAuthToken       string
	OAuthTokenSecret string
}

func createAccessToken(client *BClient) (*accessToken, error) {
	accessToken := readCachedAccessToken()
	if accessToken != nil {
		return accessToken, nil
	}

	accessToken, err := auth(client.ConsumerKey, client.ConsumerSecret)
	if err != nil {
		return nil, err
	}
	cacheAccessToken(accessToken)
	return accessToken, nil
}

func cacheAccessToken(token *accessToken) {
	tokenByte, _ := json.MarshalIndent(token, "", "  ")
	err := ioutil.WriteFile("/tmp/token", tokenByte, 0644)
	if err != nil {
		panic(err)
	}
}
func readCachedAccessToken() *accessToken {
	tokenBytes, err := ioutil.ReadFile("/tmp/token")
	if err != nil {
		return nil
	}
	token := accessToken{}
	json.Unmarshal(tokenBytes, &token)
	return &token
}

func signDataRequest(req *http.Request, t *accessToken, consumerKey, consumerSecret string) {
	params := map[string]string{}
	params["oauth_consumer_key"] = consumerKey
	params["oauth_nonce"] = nonce()
	params["oauth_signature_method"] = "HMAC-SHA1"
	params["oauth_timestamp"] = timestamp()
	params["oauth_version"] = "1.0"
	params["oauth_token"] = t.OAuthToken
	
	key := escape(consumerSecret) + "&" + escape(t.OAuthTokenSecret)
	base := requestString(req.Method, req.URL.String(), params)

	params["oauth_signature"] = sign(base, key)
	
	req.Header.Add("Authorization", authorizationString(params))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
}

func auth(consumerKey, consumerSecret string) (*accessToken, error) {

	urlToken := "https://bitbucket.org/api/1.0/oauth/request_token/"

	token, err := createRequestToken(urlToken, consumerKey, consumerSecret)
	if err != nil {
		return nil, err
	}

	redirectUrl, err := authorizationRedirect(token)
	if err != nil {
		return nil, err
	}
	fmt.Println(redirectUrl)

	//read token returned on the page
	fmt.Println("Enter code: ")
	var verifier string
	fmt.Scanln(&verifier)

	accessToken, err := authorizeToken(token, consumerKey, consumerSecret, verifier)
	if err != nil {
		return nil, err
	}

	return accessToken, nil
}

func authorizeToken(t *requestToken, consumerKey, consumerSecret, verifier string) (*accessToken, error) {
	accessTokenUrl, _ := url.Parse("https://bitbucket.org/api/1.0/oauth/access_token/")
	req := http.Request{
		URL:        accessTokenUrl,
		Method:     "POST",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Close:      true,
	}
	req.Header = http.Header{}

	signAuthorizeToken(&req, t, consumerKey, consumerSecret, verifier)

	resp, err := http.DefaultClient.Do(&req)
	if err != nil {
		return nil, err
	}
	accessToken, err := parseAccessToken(resp.Body)
	if err != nil {
		return nil, err
	}

	return accessToken, nil
}

func parseAccessToken(reader io.ReadCloser) (*accessToken, error) {
	body, err := ioutil.ReadAll(reader)
	reader.Close()
	if err != nil {
		return nil, err
	}
	str := string(body)
	parts, err := url.ParseQuery(str)
	if err != nil {
		return nil, err
	}

	oauthToken := parts.Get("oauth_token")
	oauthTokenSecret := parts.Get("oauth_token_secret")
	token := accessToken{OAuthToken: oauthToken, OAuthTokenSecret: oauthTokenSecret}
	return &token, nil
}

func signAuthorizeToken(req *http.Request, t *requestToken, consumerKey, consumerSecret, verifier string) {
	params := map[string]string{}
	params["oauth_verifier"] = verifier
	params["oauth_consumer_key"] = consumerKey
	params["oauth_nonce"] = nonce()
	params["oauth_signature_method"] = "HMAC-SHA1"
	params["oauth_timestamp"] = timestamp()
	params["oauth_version"] = "1.0"
	params["oauth_callback"] = "oob"
	params["oauth_token"] = t.token
	
	key := escape(consumerSecret) + "&" + escape(t.secret)
	base := requestString(req.Method, req.URL.String(), params)
	
	params["oauth_signature"] = sign(base, key)
	
	req.Header.Add("Authorization", fmt.Sprintf("OAuth %s", authorizationString(params)))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
}

func authorizationRedirect(t *requestToken) (string, error) {
	authorizationUrl := "https://bitbucket.org/!api/1.0/oauth/authenticate"
	redirectUrl, _ := url.Parse(authorizationUrl)
	params := make(url.Values)
	params.Add("oauth_token", t.token)
	redirectUrl.RawQuery = params.Encode()

	u := redirectUrl.String()
	if strings.HasPrefix(u, "https://bitbucket.org/%21api/") {
		u = strings.Replace(u, "/%21api/", "/!api/", -1)
	}

	return u, nil
}

func createRequestToken(tokenUrl, consumerKey, consumerSecret string) (*requestToken, error) {
	requestTokenUrl, _ := url.Parse(tokenUrl)
	req := http.Request{
		URL:        requestTokenUrl,
		Method:     "POST",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Close:      true,
	}
	req.Header = http.Header{}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	signRequest(&req, consumerKey, consumerSecret)

	resp, err := http.DefaultClient.Do(&req)
	if err != nil {
		panic(err)
	}

	requestToken, err := parserequestToken(resp.Body)
	if err != nil {
		return nil, err
	}
	return requestToken, nil
}

func signRequest(req *http.Request, consumerKey, consumerSecret string) {
	params := map[string]string{}

	params["oauth_consumer_key"] = consumerKey
	params["oauth_nonce"] = nonce()
	params["oauth_signature_method"] = "HMAC-SHA1"
	params["oauth_timestamp"] = timestamp()
	params["oauth_version"] = "1.0"
	params["oauth_callback"] = "oob" //"http://localhost:8080/auth/bitbucket"
	
	//we'll need to sign any form values
	if req.Form != nil {
		for k, _ := range req.Form {
			params[k] = req.Form.Get(k)
		}
	}

	queryParams := req.URL.Query()
	for k, _ := range queryParams {
		params[k] = queryParams.Get(k)
	}

	key := escape(consumerSecret) + "&" + escape("")
	base := requestString(req.Method, req.URL.String(), params)
	params["oauth_signature"] = sign(base, key)
	
	req.Header.Add("Authorization", fmt.Sprintf("OAuth %s", authorizationString(params)))
}

func parserequestToken(reader io.ReadCloser) (*requestToken, error) {
	body, err := ioutil.ReadAll(reader)
	reader.Close()
	if err != nil {
		return nil, err
	}

	requestData := string(body)
	parts, err := url.ParseQuery(requestData)
	if err != nil {
		return nil, err
	}

	token := requestToken{}
	token.token = parts.Get("oauth_token")
	token.secret = parts.Get("oauth_token_secret")
	token.callbackConfirmed = parts.Get("oauth_callback_confirmed") == "true"

	//some error checking ...
	switch {
	case len(token.token) == 0:
		return nil, errors.New(requestData)
	case len(token.secret) == 0:
		return nil, errors.New(requestData)
	}

	return &token, nil
}

// Generates an HMAC Signature for an OAuth1.0a request.
func sign(message, key string) string {
	hashfun := hmac.New(sha1.New, []byte(key))
	hashfun.Write([]byte(message))
	rawsignature := hashfun.Sum(nil)
	base64signature := make([]byte, base64.StdEncoding.EncodedLen(len(rawsignature)))
	base64.StdEncoding.Encode(base64signature, rawsignature)

	return string(base64signature)
}

func authorizationString(params map[string]string) string {

	// loop through params, add keys to map
	var keys []string
	for key, _ := range params {
		keys = append(keys, key)
	}

	// sort the array of header keys
	sort.StringSlice(keys).Sort()

	// create the signed string
	var str string
	var cnt = 0

	// loop through sorted params and append to the string
	for _, key := range keys {

		// we previously encoded all params (url params, form data & oauth params)
		// but for the authorization string we should only encode the oauth params
		fmt.Println(key)
		if !strings.HasPrefix(key, "oauth_") && key != "realm" {
			continue
		}

		if cnt > 0 {
			str += ","
		}
		
		if key == "oauth_signature" || key == "realm" {
			str += fmt.Sprintf("%s=%q", key, params[key])
		} else {
			str += fmt.Sprintf("%s=%q", key, escape(params[key]))
		}
		cnt++
	}

	return str
}

// Nonce generator, seeded with current time
var nonceGenerator = rand.New(rand.NewSource(time.Now().Unix()))

// Nonce generates a random string. Nonce's are uniquely generated
// for each request.
func nonce() string {
	return strconv.FormatInt(nonceGenerator.Int63(), 10)
}

// Timestamp generates a timestamp, expressed in the number of seconds
// since January 1, 1970 00:00:00 GMT.
func timestamp() string {
	return strconv.FormatInt(time.Now().Unix(), 10)
}
func requestString(method string, uri string, params map[string]string) string {

	// loop through params, add keys to map
	var keys []string
	for key, _ := range params {
		keys = append(keys, key)
	}

	// sort the array of header keys
	sort.StringSlice(keys).Sort()

	// create the signed string
	result := method + "&" + escape(uri)

	// loop through sorted params and append to the string
	for pos, key := range keys {
		if pos == 0 {
			result += "&"
		} else {
			result += escape("&")
		}

		result += escape(fmt.Sprintf("%s=%s", key, escape(params[key])))
	}

	return result
}

func escape(s string) string {
	t := make([]byte, 0, 3*len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if isEscapable(c) {
			t = append(t, '%')
			t = append(t, "0123456789ABCDEF"[c>>4])
			t = append(t, "0123456789ABCDEF"[c&15])
		} else {
			t = append(t, s[i])
		}
	}
	return string(t)
}

func isEscapable(b byte) bool {
	return !('A' <= b && b <= 'Z' || 'a' <= b && b <= 'z' || '0' <= b && b <= '9' || b == '-' || b == '.' || b == '_' || b == '~')
}
