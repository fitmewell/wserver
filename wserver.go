package wserver

import (
	"errors"
	"github.com/fitmewell/wserver/bdb"
	"github.com/fitmewell/wserver/wsession"
	"sync"
)

//Default sever context defined for simple running , use customer Server if you want a custom config
var DefaultSever *Server

var configLock sync.Mutex

//set config file path for default server context
func SetConfigPath(filePath string) {
	config, err := NewConfig(filePath)
	if err != nil {
		panic(err)
	}
	SetConfig(config)
}

//set config for default server context
func SetConfig(config *ServerConfig) {
	configLock.Lock()
	defer configLock.Unlock()

	DefaultSever = &Server{
		config:         config,
		sessionManager: wsession.NewDefaultSessionManager(config.Session.CookieName),
		aftermaths:     map[string]func(){},
	}

	DefaultSever.handler = newDefaultHandler(DefaultSever)
	DefaultSever.context = NewContextFrom(DefaultSever.config)
}

//set port for default server context
func SetPort(port string) {
	if DefaultSever.config != nil {
		DefaultSever.config.Port = port
	} else {
		config := &ServerConfig{Port: port}
		SetConfig(config)
	}
}

//add handler to default server
func AddHandler(method string, path string, e interface{}) *Server {
	DefaultSever.handler.addHandler(method, path, e)
	return DefaultSever
}

//add static source path handler to default server
func AddStaticSource(path, fileLocate string) *Server {
	DefaultSever.config.StaticResources = append(DefaultSever.config.StaticResources, StaticResource{Path: path, FileLocate: fileLocate})
	return DefaultSever
}

//add template path to default server
func AddTemplate(name, dir string) *Server {
	DefaultSever.config.Templates = append(DefaultSever.config.Templates, Template{Name: name, Dir: dir})
	return DefaultSever
}

//add aspect handler to default server
func AddAspectHandler(handler AspectHandler) *Server {
	DefaultSever.handler.addAspect(handler)
	return DefaultSever
}

//add handler before server close ,you have 10 second before server close to finish your work to default server
func AddAftermath(name string, method func()) error {
	if _, ok := DefaultSever.aftermaths[name]; ok {
		return errors.New("duplicate aftermatch found")
	}
	DefaultSever.aftermaths[name] = method
	return nil
}

//get default server's context properties
func GetProperty(i string) string {
	return DefaultSever.context.GetProperty(i)
}

//get default server's default DB
func GetDb() bdb.BufferedDB {
	return DefaultSever.context.GetDb()
}

//get default server's select DB
func GetSelectDb(name string) bdb.BufferedDB {
	return DefaultSever.context.GetSelectDb(name)
}

//start default server
func Start() {
	DefaultSever.Start()
}
