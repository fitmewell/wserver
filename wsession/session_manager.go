package wsession

import (
	"crypto/rand"
	"encoding/base64"
	"io"
	"net/http"
	"sync"
	"time"
)

type SessionManager interface {
	NewSession() Session
	DeleteSession(string) error
	GetSession(string) Session
	NewId() string
	Scan()
	Sync(http.ResponseWriter, *http.Request) Session
}

type defaultSessionManager struct {
	sessionMap map[string]Session
	cookieName string
	scanTime   time.Duration
	lifeTime   time.Duration
	lock       sync.Mutex
}

func (ds *defaultSessionManager) NewSession() Session {
	tmpSession := &defaultSession{name: ds.NewId(), properties: map[string]interface{}{}, createTime: time.Now(), expireTime: time.Now().Add(ds.lifeTime)}
	ds.sessionMap[tmpSession.Name()] = tmpSession
	return tmpSession
}
func (ds *defaultSessionManager) DeleteSession(key string) error {
	delete(ds.sessionMap, key)
	return nil
}
func (ds *defaultSessionManager) GetSession(key string) Session {
	return ds.sessionMap[key]
}
func (ds *defaultSessionManager) Scan() {
	for key, value := range ds.sessionMap {
		if value.IsExpire() {
			ds.DeleteSession(key)
		}
	}
	time.AfterFunc(ds.scanTime, ds.Scan)
}
func (ds *defaultSessionManager) NewId() string {
	b := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}
func (ds *defaultSessionManager) Sync(resp http.ResponseWriter, req *http.Request) Session {
	ds.lock.Lock()
	defer ds.lock.Unlock()
	ck, _ := req.Cookie(ds.cookieName)
	var tmpSession Session
	if ck == nil || ds.GetSession(ck.Value) == nil || ds.GetSession(ck.Value).IsExpire() {
		tmpSession = ds.NewSession()
	} else {
		tmpSession = ds.GetSession(ck.Value)
		tmpSession.SetExpireTime(time.Now().Add(ds.lifeTime))
	}
	cookie := &http.Cookie{Name: ds.cookieName, Value: tmpSession.Name(), Path: "/", Expires: tmpSession.GetExpireTime()}
	http.SetCookie(resp, cookie)
	return tmpSession
}

func NewDefaultSessionManager(name string) SessionManager {
	manager := &defaultSessionManager{sessionMap: map[string]Session{}, cookieName: name, scanTime: 1 * time.Minute, lifeTime: 15 * time.Minute}
	go manager.Scan()
	return manager
}
