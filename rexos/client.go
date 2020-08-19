package rexos

import (
	"bytes"
	"context"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/roboticeyes/gococo/cron"
	"github.com/roboticeyes/gococo/event"
)

// AccessTokenType is used for the key of the authorization token for the context
type AccessTokenType string

// UserIDType is used for the key of the user_id in the context
type UserIDType string

const (
	// AccessTokenKey is the key for the context information. The context needs to store the
	// full access token with "bearer <token>"
	AccessTokenKey AccessTokenType = "authorization"
	// UserIDKey is the key for the context information. The context needs to store the
	// REXos user id
	UserIDKey UserIDType = "UserID"

	// MaxTrials defines the maximum trials for sensitive requests to recover any errors
	MaxTrials = 3
)

// JwtToken is the token which is returned from REXos
type JwtToken struct {
	AccessToken     string      `json:"access_token"`
	TokenType       string      `json:"token_type"`
	ExpiresIn       uint64      `json:"expires_in"`
	Scope           string      `json:"scope"`
	UserID          interface{} `json:"user_id"`
	UserName        interface{} `json:"user_name"`
	UserDisplayName interface{} `json:"user_display_name"`
	Jti             string      `json:"jti"`
}

// Client is the client which is used to send requests to the REXos. The client
// should be created once and shared among all services.
type Client struct {
	httpClient   *http.Client
	config       Config
	serviceToken JwtToken   // this is the service user token which gets updated using a cron job
	mutex        sync.Mutex // used for accessing the token in parallel
}

// NewClient create a new REXos HTTP client
func NewClient(cfg Config) *Client {
	client := &Client{
		httpClient: http.DefaultClient,
		config:     cfg,
	}

	if !client.config.NotApplyServiceUser {
		go client.scheduleTokenRefreshHandler()
	}
	return client
}

func (c *Client) refreshToken() {

	log.Info("Refreshing service user token ...")

	payload := c.config.ClientID + ":" + c.config.ClientSecret
	encodedToken := b64.StdEncoding.EncodeToString([]byte(payload))
	req, _ := http.NewRequest("POST", c.config.AccessTokenURL, bytes.NewReader([]byte(`grant_type=client_credentials`)))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Basic "+encodedToken)

	resp, err := c.httpClient.Do(req)

	if err != nil {
		log.Error("Service user authentication: internal POST request error -", err)
		return
	}

	// this is required to properly empty the buffer for the next call
	defer func() {
		io.Copy(ioutil.Discard, resp.Body)
	}()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Error("Service user authentication: cannot get body for authentication -", err)
		return
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()
	err = json.Unmarshal(body, &c.serviceToken)
	if err != nil {
		log.Error("Service user authentication: cannot decode JWT token -", err)
		return
	}
}

// scheduleTokenRefreshHandler starts a cron job which makes sure that the service user always has
// a valid token.
func (c *Client) scheduleTokenRefreshHandler() {

	c.refreshToken()

	var interval uint64 = 600
	if (c.serviceToken.ExpiresIn - 30) > 60 {
		interval = c.serviceToken.ExpiresIn - 30
	}
	// Take expiration attribute and make sure to early update the token (30 seconds before)
	log.Info("Starting cron job to refresh service user token with interval " + strconv.FormatUint(interval, 10) + "s")
	cron.Every(interval).Seconds().Do(c.refreshToken)
	<-cron.Start()
}

// GetWithServiceUser performs the GET request with the credentials of the service user
func (c *Client) GetWithServiceUser(ctx context.Context, query string, authenticate bool) (string, []byte, int, error) {
	if c.config.NotApplyServiceUser {
		return "", nil, http.StatusForbidden, fmt.Errorf("No service user initialized")
	}

	xf, err := GetXForwarded(ctx)
	if err != nil {
		fmt.Println("host error service user")
		fmt.Println(err)
		return "", nil, http.StatusForbidden, fmt.Errorf("Cannot get host")
	}
	c.mutex.Lock()
	token := "Bearer " + c.serviceToken.AccessToken
	c.mutex.Unlock()
	return c.get(token, xf, query, authenticate)
}

// Get performs the GET request with the credentials of the client user (stored in the token)
func (c *Client) Get(ctx context.Context, query string, authenticate bool) (string, []byte, int, error) {

	token, err := GetAccessTokenFromContext(ctx)
	if err != nil {
		return "", nil, http.StatusForbidden, fmt.Errorf("Missing token in context")
	}

	xf, err := GetXForwarded(ctx)
	if err != nil {
		fmt.Println("host error")
		fmt.Println(err)
		return "", nil, http.StatusForbidden, fmt.Errorf("Cannot get host")
	}
	return c.get(token, xf, query, authenticate)
}

// Get performs a GET request to the given query and returns the body response which is of type JSON.
// The return values also contain the http status code and a potential error which has occured.
// The request will be setup as JSON request and also takes out the authentication information from
// the given context.
func (c *Client) get(token string, xf XForwarded, query string, authenticate bool) (string, []byte, int, error) {

	req, _ := http.NewRequest("GET", query, nil)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("X-Requested-With", "XMLHttpRequest")
	req.Header.Add("X-Forwarded-Host", xf.Host)
	req.Header.Add("X-Forwarded-Port", xf.Port)
	req.Header.Add("X-Forwarded-Proto", xf.Proto)
	req.Header.Add("X-Forwarded-Prefix", c.config.BasePathExtern)

	if authenticate {
		req.Header.Add("Authorization", token)
	}

	var fileName string
	trials := 0
	for ; trials < MaxTrials; trials++ {
		if trials > 0 {
			log.Debugf("Internal GET %s: trial %d\n", query, trials)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			log.WithFields(event.Fields{
				"query":        query,
				"errorMessage": err.Error(),
			}).Debug("Internal GET request error")
		}

		// Check for content-disposition to extract optional fileName
		contentDisposition := resp.Header.Get("Content-Disposition")
		if contentDisposition != "" {
			_, params, err := mime.ParseMediaType(contentDisposition)
			if err == nil {
				fileName = params["filename"]
			}
		}

		if err != nil {
			log.Error("Internal GET request error: ", err)
			return fileName, nil, resp.StatusCode, err
		}
		// this is required to properly empty the buffer for the next call
		defer func() {
			io.Copy(ioutil.Discard, resp.Body)
		}()

		body, err := ioutil.ReadAll(resp.Body)

		if resp.StatusCode == http.StatusRequestTimeout {
			// GET request timed out. Retrying..
			time.Sleep(time.Millisecond * 100)
			continue
		}

		// Other error means outside the 2xx range
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			log.WithFields(event.Fields{
				"body": string(body),
			}).Debugf("Internal GET request did not return 2xx as expected but returned %d", resp.StatusCode)
			return fileName, body, resp.StatusCode, fmt.Errorf("Internal GET request failed after %d trials", trials+1)
		}

		// success
		return fileName, body, resp.StatusCode, err
	}

	return fileName, nil, http.StatusRequestTimeout, fmt.Errorf("Internal GET request failed after %d trials", trials+1)
}

// PostWithServiceUser performs the POST request with the credentials of the service user
func (c *Client) PostWithServiceUser(ctx context.Context, query string, payload io.Reader, contentType string) ([]byte, int, error) {
	if c.config.NotApplyServiceUser {
		return nil, http.StatusForbidden, fmt.Errorf("No service user initialized")
	}
	c.mutex.Lock()
	token := "Bearer " + c.serviceToken.AccessToken
	c.mutex.Unlock()
	return c.post(token, query, payload, contentType)
}

// Post performs the GET request with the credentials of the client user (stored in the token)
func (c *Client) Post(ctx context.Context, query string, payload io.Reader, contentType string) ([]byte, int, error) {

	token, err := GetAccessTokenFromContext(ctx)
	if err != nil {
		return nil, http.StatusForbidden, fmt.Errorf("Missing token in context")
	}

	return c.post(token, query, payload, contentType)
}

// Post performs a POST request to the given query, using the given payload as data, and the provided
// content-type. The content-type is typically 'application/json', but can also be of formdata in case of
// binary data upload.
// WARNING: Do NOT implement retries here. POST is not considered a safe nor idempotent HTTP method!
func (c *Client) post(token string, query string, payload io.Reader, contentType string) ([]byte, int, error) {

	req, _ := http.NewRequest("POST", query, payload)
	req.Header.Add("Content-Type", contentType)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("X-Requested-With", "XMLHttpRequest")
	req.Header.Add("Authorization", token)

	resp, err := c.httpClient.Do(req)

	if err != nil {
		log.WithFields(event.Fields{
			"query":        query,
			"contentType":  contentType,
			"errorMessage": err.Error(),
		}).Debug("Internal POST request error")
		return []byte{}, http.StatusInternalServerError, nil
	}

	// this is required to properly empty the buffer for the next call
	defer func() {
		io.Copy(ioutil.Discard, resp.Body)
	}()

	body, err := ioutil.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusConflict {
		// Convention: Do not try to query and return existing resource here.
		log.WithFields(event.Fields{
			"query":       query,
			"contentType": contentType,
		}).Debug("Resource already exists")
		return body, resp.StatusCode, nil
	}

	if resp.StatusCode == http.StatusRequestTimeout {
		// POST request timed out.
		return []byte{}, http.StatusRequestTimeout, fmt.Errorf("Internal POST request timed out")
	}

	// Other error means outside the 2xx range
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.WithFields(event.Fields{
			"body": string(body),
		}).Debugf("Internal POST request did not return 2xx as expected but returned %d", resp.StatusCode)
		return body, resp.StatusCode, fmt.Errorf("Internal POST request failed")
	}

	// success
	return body, resp.StatusCode, err

}

// PatchWithServiceUser performs the PATCH request with the credentials of the service user
func (c *Client) PatchWithServiceUser(ctx context.Context, query string, payload io.Reader, contentType string) ([]byte, int, error) {
	if c.config.NotApplyServiceUser {
		return nil, http.StatusForbidden, fmt.Errorf("No service user initialized")
	}

	c.mutex.Lock()
	token := "Bearer " + c.serviceToken.AccessToken
	c.mutex.Unlock()
	return c.patch(token, query, payload, contentType)
}

// Patch performs the PATCH request with the credentials of the client user (stored in the token)
func (c *Client) Patch(ctx context.Context, query string, payload io.Reader, contentType string) ([]byte, int, error) {

	token, err := GetAccessTokenFromContext(ctx)
	if err != nil {
		return nil, http.StatusForbidden, fmt.Errorf("Missing token in context")
	}

	return c.patch(token, query, payload, contentType)
}

// Patch performs a PATCH request to the given query, using the given payload as data, and the provided
// content-type. The content-type is typically 'application/json', but can also be of formdata in case of
// binary data upload.
// WARNING: Do NOT implement retries here. PATCH is not considered a safe nor idempotent HTTP method!
func (c *Client) patch(token, query string, payload io.Reader, contentType string) ([]byte, int, error) {

	req, _ := http.NewRequest("PATCH", query, payload)
	req.Header.Add("Content-Type", contentType)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("X-Requested-With", "XMLHttpRequest")
	req.Header.Add("Authorization", token)

	resp, err := c.httpClient.Do(req)

	if err != nil {
		log.Error("Internal PATCH request error: ", err)
	}

	// this is required to properly empty the buffer for the next call
	defer func() {
		io.Copy(ioutil.Discard, resp.Body)
	}()

	body, err := ioutil.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusRequestTimeout {
		// PATCH request timed out.
		return []byte{}, http.StatusRequestTimeout, fmt.Errorf("Internal PATCH request timed out")
	}

	// Other error means outside the 2xx range
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.WithFields(event.Fields{
			"body": string(body),
		}).Debugf("Internal PATCH request did not return 2xx as expected but returned %d", resp.StatusCode)
		return body, resp.StatusCode, fmt.Errorf("Internal PATCH request failed")
	}

	// success
	return body, resp.StatusCode, err
}

// DeleteWithServiceUser performs the DELETE request with the credentials of the service user
func (c *Client) DeleteWithServiceUser(ctx context.Context, link string) ([]byte, int, error) {
	if c.config.NotApplyServiceUser {
		return nil, http.StatusForbidden, fmt.Errorf("No service user initialized")
	}

	c.mutex.Lock()
	token := "Bearer " + c.serviceToken.AccessToken
	c.mutex.Unlock()
	return c.delete(token, link)
}

// Delete performs the DELETE request with the credentials of the client user (stored in the token)
func (c *Client) Delete(ctx context.Context, link string) ([]byte, int, error) {

	token, err := GetAccessTokenFromContext(ctx)
	if err != nil {
		return nil, http.StatusForbidden, fmt.Errorf("Missing token in context")
	}

	return c.delete(token, link)
}

// Delete sends a DELETE request to the given link.
func (c *Client) delete(token, link string) ([]byte, int, error) {

	req, _ := http.NewRequest("DELETE", link, nil)
	req.Header.Add("X-Requested-With", "XMLHttpRequest")
	req.Header.Add("authorization", token)

	trials := 0
	for ; trials < MaxTrials; trials++ {
		if trials > 0 {
			log.Debugf("DELETE %s: trial %d\n", link, trials)
		}

		resp, err := c.httpClient.Do(req)

		if err != nil {
			log.Error("Internal DELETE request error: ", err)
		}

		// this is required to properly empty the buffer for the next call
		defer func() {
			io.Copy(ioutil.Discard, resp.Body)
		}()

		body, err := ioutil.ReadAll(resp.Body)

		if resp.StatusCode == http.StatusRequestTimeout {
			// DELETE request timed out. Retrying..
			time.Sleep(time.Millisecond * 100)
			continue
		}

		// Other error means outside the 2xx range
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			log.WithFields(event.Fields{
				"body": string(body),
			}).Debugf("Internal DELETE request did not return 2xx as expected but returned %d", resp.StatusCode)
			return body, resp.StatusCode, fmt.Errorf("Internal DELETE request failed after %d trials", trials+1)
		}

		// success
		return []byte{}, resp.StatusCode, err
	}

	return []byte{}, http.StatusRequestTimeout, fmt.Errorf("Internal DELETE request failed after %d trials", trials+1)
}

// GetFileWithServiceUser performs the GET request with the credentials of the service user
func (c *Client) GetFileWithServiceUser(ctx context.Context, context *gin.Context, query string, authenticate bool) (int, error) {
	if c.config.NotApplyServiceUser {
		return http.StatusForbidden, fmt.Errorf("No service user initialized")
	}

	xf, err := GetXForwarded(ctx)
	if err != nil {
		return http.StatusForbidden, fmt.Errorf("Cannot get host")
	}
	c.mutex.Lock()
	token := "Bearer " + c.serviceToken.AccessToken
	c.mutex.Unlock()
	return c.getFile(context, token, xf, query, authenticate)
}

// getFile performs a GET request to the given query and forwards the file from the given url
// The return values also contain the http status code and a potential error which has occured.
// The request takes out the authentication information from the given context.
func (c *Client) getFile(context *gin.Context, token string, xf XForwarded, query string, authenticate bool) (int, error) {

	req, _ := http.NewRequest("GET", query, nil)
	req.Header.Add("Content-Type", "application/octet-stream")
	req.Header.Add("Accept", "application/octet-stream")
	req.Header.Add("X-Requested-With", "XMLHttpRequest")
	req.Header.Add("X-Forwarded-Host", xf.Host)
	req.Header.Add("X-Forwarded-Port", xf.Port)
	req.Header.Add("X-Forwarded-Proto", xf.Proto)
	req.Header.Add("X-Forwarded-Prefix", c.config.BasePathExtern)

	if authenticate {
		req.Header.Add("Authorization", token)
	}

	response, err := c.httpClient.Do(req)
	if err != nil {
		log.WithFields(event.Fields{
			"query":        query,
			"errorMessage": err.Error(),
		}).Debug("Internal GET request error")
		return http.StatusServiceUnavailable, err
	}
	if response.StatusCode != http.StatusOK {
		log.WithFields(event.Fields{
			"query": query,
		}).Debug("Internal GET request error")
		return response.StatusCode, err
	}

	// Check for content-disposition to extract optional fileName
	fileName := ""
	contentDisposition := response.Header.Get("Content-Disposition")
	if contentDisposition != "" {
		_, params, err := mime.ParseMediaType(contentDisposition)
		if err == nil {
			fileName = params["filename"]
		}
	}

	reader := response.Body
	contentLength := response.ContentLength
	contentType := response.Header.Get("Content-Type")

	extraHeaders := map[string]string{
		"Content-Disposition": `attachment; filename=` + fileName,
	}
	context.DataFromReader(http.StatusOK, contentLength, contentType, reader, extraHeaders)

	return response.StatusCode, nil
}
