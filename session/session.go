package session

import (
	"time"
)

type Session interface {
	Name() string
	Get(key string) interface{}
	Set(key string, value interface{})
	Del(key string)
	SetCreateTime(time.Time) error
	SetExpireTime(time.Time) error
	GetCreateTime() time.Time
	GetExpireTime() time.Time
	IsExpire() bool
}

type defaultSession struct {
	name       string
	properties map[string]interface{}
	createTime time.Time
	expireTime time.Time
}

func (ds *defaultSession) Name() string {
	return ds.name
}

func (ds *defaultSession) Get(key string) interface{} {
	return ds.properties[key]
}
func (ds *defaultSession) Set(key string, value interface{}) {
	ds.properties[key] = value
}

func (ds *defaultSession) SetCreateTime(t time.Time) error {
	ds.createTime = t
	return nil
}
func (ds *defaultSession) SetExpireTime(t time.Time) error {
	ds.expireTime = t
	return nil
}
func (ds *defaultSession) Del(key string) {
	delete(ds.properties, key)
}
func (ds *defaultSession) GetCreateTime() time.Time {
	return ds.createTime
}
func (ds *defaultSession) GetExpireTime() time.Time {
	return ds.expireTime
}
func (ds *defaultSession) IsExpire() bool {
	return time.Now().After(ds.GetExpireTime())
}
