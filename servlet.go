package wserver

import "io"

type ServletContext interface {
	ServerContext

	//Get servlet session
	GetSession() *session

	//get data stored in servlet
	GetData() map[string]interface{}

	//set data store in servlet
	SetData(key string, value interface{})
}

/**
    Context build for per request
 */
type DefaultServletContext struct {
	ServerContext ServerContext
	Session       *session
	data          map[string]interface{}
}

func (defaultContext *DefaultServletContext)GetDb() BufferedDB {
	return defaultContext.ServerContext.GetDb()
}

func (defaultContext *DefaultServletContext)GetSelectDb(dbName string) BufferedDB {
	return defaultContext.ServerContext.GetSelectDb(dbName)
}

func (defaultContext *DefaultServletContext)GetProperty(key string) string {
	return defaultContext.ServerContext.GetProperty(key)
}

func (defaultContext *DefaultServletContext)ContainsProperty(key string) bool {
	return defaultContext.ServerContext.ContainsProperty(key)
}

func (defaultContext *DefaultServletContext)GetData() map[string]interface{} {
	return defaultContext.data
}

func (defaultContext *DefaultServletContext)SetData(key string, value interface{}) {
	defaultContext.data[key] = value
}

func (defaultContext *DefaultServletContext)ExecuteTemplate(wr io.Writer, name string, data interface{}) error {
	return defaultContext.ServerContext.ExecuteTemplate(wr, name, data)
}

func (defaultContext *DefaultServletContext)GetSession() *session {
	return defaultContext.Session
}