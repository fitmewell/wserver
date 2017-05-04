package wserver

import (
	"crypto/rand"
	"encoding/base64"
	"io"
	"net/http"
	"time"
)

type session struct {
	Name       string
	Properties map[string]interface{}
	CreateTime time.Time
	ExpireTime time.Time
}

func (ds *session) Get(key string) interface{} {
	return ds.Properties[key]
}
func (ds *session) Set(key string, value interface{}) {
	ds.Properties[key] = value
}
func (ds *session) Del(key string) {
	delete(ds.Properties, key)
}
func (ds *session) GetCreateTime() time.Time {
	return ds.CreateTime
}
func (ds *session) GetExpireTime() time.Time {
	return ds.ExpireTime
}
func (ds *session) IsExpire() bool {
	return time.Now().After(ds.GetExpireTime())
}

type SessionManager interface {
	NewSession() *session
	DeleteSession(string) error
	GetSession(string) *session
	NewId() string
	Scan()
	Sync(http.ResponseWriter, *http.Request) *session
}

type defaultSessionManager struct {
	sessionMap map[string]*session
	cookieName string
	scanTime   time.Duration
	lifeTime   time.Duration
}

func (ds *defaultSessionManager) NewSession() *session {
	tmpSession := &session{Name: ds.NewId(), Properties: map[string]interface{}{}, CreateTime: time.Now(), ExpireTime: time.Now().Add(ds.lifeTime)}
	ds.sessionMap[tmpSession.Name] = tmpSession
	return tmpSession
}
func (ds *defaultSessionManager) DeleteSession(key string) error {
	delete(ds.sessionMap, key)
	return nil
}
func (ds *defaultSessionManager) GetSession(key string) *session {
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
func (ds *defaultSessionManager) Sync(resp http.ResponseWriter, req *http.Request) *session {
	ck, _ := req.Cookie(ds.cookieName)
	var tmpSession *session
	if ck == nil || ds.GetSession(ck.Value) == nil || ds.GetSession(ck.Value).IsExpire() {
		tmpSession = ds.NewSession()
	} else {
		tmpSession = ds.GetSession(ck.Value)
		tmpSession.ExpireTime = time.Now().Add(ds.lifeTime)
	}
	cookie := &http.Cookie{Name: ds.cookieName, Value: tmpSession.Name, Path: "/", Expires: tmpSession.ExpireTime}
	http.SetCookie(resp, cookie)
	return tmpSession
}

func NewDefaultSessionManager(config *ServerConfig) SessionManager {
	manager := &defaultSessionManager{sessionMap: map[string]*session{}, cookieName: "hlsSession", scanTime: 1 * time.Minute, lifeTime: 15 * time.Minute}
	go manager.Scan()
	return manager
}
