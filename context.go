package wserver

import (
	"io"
	"html/template"
	"database/sql"
	"log"
)

type ServerContext interface {
	GetDb() BufferedDB
	GetSelectDb(string) BufferedDB
	ExecuteTemplate(io.Writer, string, interface{}) error
	GetProperty(string) string
	ContainsProperty(string) bool
}

type DefaultServerContext struct {
	defaultDb  BufferedDB
	dbs        map[string]BufferedDB
	template   *template.Template
	properties map[string]string
}

func (defaultContext *DefaultServerContext)GetDb() BufferedDB {
	return defaultContext.defaultDb
}

func (defaultContext *DefaultServerContext)GetSelectDb(dbName string) BufferedDB {
	return defaultContext.dbs[dbName]
}

func (defaultContext *DefaultServerContext)GetProperty(key string) string {
	return defaultContext.properties[key]
}

func (defaultContext *DefaultServerContext)ContainsProperty(key string) bool {
	_, ok := defaultContext.properties[key]
	return ok
}

func (defaultContext *DefaultServerContext)ExecuteTemplate(wr io.Writer, name string, data interface{}) error {
	return defaultContext.template.ExecuteTemplate(wr, name, data)
}

func NewContextFrom(config *ServerConfig) *DefaultServerContext {
	temp := template.New("default")
	for _, templateConfig := range config.Templates {
		dir := templateConfig.Dir
		temp.ParseGlob(dir + "/*")
	}
	dbs := map[string]BufferedDB{}
	var defaultDb BufferedDB = nil
	for _, dbConfig := range config.Databases {
		db, err := sql.Open("mysql", dbConfig.GenerateUrl())
		if err != nil {
			log.Fatal("db {} connection failed ", dbConfig.DbName, err)
		}
		db.SetMaxOpenConns(dbConfig.MaxConnections)
		bufferedDb := NewBufferedDb(db)
		dbs[dbConfig.DbName] = bufferedDb
		if dbConfig.IsDefault {
			defaultDb = bufferedDb
		}
		if defaultDb == nil {
			defaultDb = bufferedDb
		}
	}

	properties := map[string]string{}
	for key, value := range config.PropertiesConfig.Properties {
		properties[key] = value
	}

	return &DefaultServerContext{defaultDb:defaultDb, dbs:dbs, template:temp, properties:properties}
}