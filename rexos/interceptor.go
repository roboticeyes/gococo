package rexos

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/radovskyb/watcher"
)

// Interceptor is a session/token handler, which can be used for local composite service
// development. This injects the local session file and adds a valid token for every request. This
// allows for testing a service locally without extra authentication.
type Interceptor struct {
	sessionFile string
	accessToken string
	tokenType   string
	httpClient  *http.Client
	mutex       sync.Mutex // used for accessing the token in parallel
}

// NewInterceptor creates a watcher for the session token information
func NewInterceptor(sessionFile string) *Interceptor {

	i := &Interceptor{
		sessionFile: sessionFile,
		httpClient:  http.DefaultClient,
	}
	i.loadToken()

	go i.watchSessionFile()
	return i
}

func (i *Interceptor) loadToken() {

	log.Println("Loading token file", i.sessionFile)
	sessionReader, err := os.Open(i.sessionFile)
	if err != nil {
		log.Error(err)
		return
	}

	buf, err := ioutil.ReadAll(sessionReader)
	if err != nil {
		log.Error(err)
		return
	}

	var session struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
	}
	err = json.Unmarshal(buf, &session)
	if err != nil {
		log.Error("Cannot unmarshal session file:", err)
	}

	i.mutex.Lock()
	defer i.mutex.Unlock()
	i.accessToken = session.AccessToken
	i.tokenType = session.TokenType
}

func (i *Interceptor) watchSessionFile() {

	watch := watcher.New()
	watch.FilterOps(watcher.Write)

	go func() {
		for {
			select {
			case <-watch.Event:
				log.Println("Session file got changed, reload token")
				i.loadToken()
			case err := <-watch.Error:
				if err == watcher.ErrWatchedFileDeleted {
					// Usually happens because the watcher looks for the file as the OS is updating it
					continue
				}
				fmt.Println("Session file cannot be loaded")
			case <-watch.Closed:
				return
			}
		}
	}()

	if err := watch.Add(i.sessionFile); err != nil {
		log.Error(err)
	}

	if err := watch.Start(time.Second * 1); err != nil {
		log.Error(err)
	}
}

// AddToken adds the token to the request
func (i *Interceptor) AddToken(c *gin.Context) {

	i.mutex.Lock()
	token := i.tokenType + " " + i.accessToken
	c.Set("authorization", token)
	i.mutex.Unlock()
}
