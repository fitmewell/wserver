package wserver

import (
	"database/sql"
	"github.com/fitmewell/wserver/bdb"
	. "github.com/fitmewell/wserver/log"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

type ServerContext interface {
	//return default db
	GetDb() bdb.BufferedDB

	//return select db
	GetSelectDb(string) bdb.BufferedDB

	//write to write by template name
	ExecuteTemplate(io.Writer, string, interface{}) error

	//get system properties
	GetProperty(string) string

	//judge if properties exists
	ContainsProperty(string) bool

	//init
	Init()
}

type DefaultServerContext struct {
	defaultDb  bdb.BufferedDB
	dbs        map[string]bdb.BufferedDB
	template   *template.Template
	properties map[string]string
	config     *ServerConfig
}

func (defaultContext *DefaultServerContext) GetDb() bdb.BufferedDB {
	return defaultContext.defaultDb
}

func (defaultContext *DefaultServerContext) GetSelectDb(dbName string) bdb.BufferedDB {
	return defaultContext.dbs[dbName]
}

func (defaultContext *DefaultServerContext) GetProperty(key string) string {
	return defaultContext.properties[key]
}

func (defaultContext *DefaultServerContext) ContainsProperty(key string) bool {
	_, ok := defaultContext.properties[key]
	return ok
}

func (defaultContext *DefaultServerContext) ExecuteTemplate(wr io.Writer, name string, data interface{}) error {
	return defaultContext.template.ExecuteTemplate(wr, name, data)
}

func (defaultContext *DefaultServerContext) Init() {
	temp := template.New("default")
	for _, templateConfig := range defaultContext.config.Templates {
		dir := templateConfig.Dir
		temp.ParseGlob(dir + "/*")
	}
	dbs := map[string]bdb.BufferedDB{}
	var defaultDb bdb.BufferedDB = nil
	for _, dbConfig := range defaultContext.config.Databases {
		db, err := sql.Open(dbConfig.DriverName, dbConfig.GenerateUrl())
		if err != nil {
			Fatal("db {} connection failed ", dbConfig.DbName, err)
		}
		db.SetMaxOpenConns(dbConfig.MaxConnections)
		bufferedDb := bdb.NewBufferedDb(db)
		dbs[dbConfig.DbName] = bufferedDb
		if dbConfig.IsDefault {
			defaultDb = bufferedDb
		}
		if defaultDb == nil {
			defaultDb = bufferedDb
		}
	}
	defaultContext.defaultDb = defaultDb
	defaultContext.dbs = dbs
	defaultContext.template = temp
}

func NewContextFrom(config *ServerConfig) *DefaultServerContext {

	properties := map[string]string{}
	for key, value := range config.PropertiesConfig.Properties {
		properties[key] = value
		DebugF("SystemProperties:'%s'='%s'", key, value)
	}
	for _, pf := range config.PropertiesConfig.PropertiesFiles {
		locate := pf.Locate
		tmp, err := os.Open(locate)
		if err != nil {
			Fatal(err)
		}
		stat, err := tmp.Stat()
		if err != nil {
			Fatal(err)
		}
		if stat.IsDir() {
			files, err := ioutil.ReadDir(pf.Locate)
			if err != nil {
				Fatal(err)
			}

			for _, file := range files {
				name := file.Name()
				if strings.HasSuffix(strings.ToUpper(name), ".PROPERTIES") {
					parsePropertiesFile(locate+name, properties)
				}
			}
		} else {
			properties = parsePropertiesFile(locate, properties)
		}
	}

	return &DefaultServerContext{properties: properties, config: config}
}

func parsePropertiesFile(locate string, properties map[string]string) map[string]string {
	Debug("loading file:" + locate)
	content, err := ioutil.ReadFile(locate)
	if err != nil {
		Fatal(err)
	}
	lines := strings.Split(string(content), "\n")
	for i, line := range lines {
		if len(strings.Trim(line, "")) == 0 {
			continue
		}
		sepIndex := strings.Index(line, "=")
		if sepIndex == -1 {
			DebugF("properties file parse failed :[%s]:%d:%s\n", locate, i, line)
		} else {
			value := line[sepIndex+1:]
			key := line[0:sepIndex]
			if old, ok := properties[key]; ok {
				DebugF("duplicate properties found [%s]{'%s'->'%s'} :[%s]:%d:%s\n", key, old, value, locate, i, line)
			}
			DebugF("SystemProperties:'%s'='%s'", key, value)
			properties[key] = value
		}
	}
	return properties
}
