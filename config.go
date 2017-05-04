package wserver

import (
	"encoding/json"
	"io/ioutil"
)

func NewConfig(path string) (*ServerConfig, error) {
	config := &ServerConfig{}
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(file, &config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

type Database struct {
	IsDefault      bool
	Name           string
	Address        string
	Port           string
	Protocol       string
	Username       string
	Password       string
	Charset        string
	DbName         string
	MaxConnections int
	DriverName     string
}

type SSLConfig struct {
	SSLPort  string
	CertFile string
	KeyFile  string
}

type StaticResource struct {
	Name       string
	Path       string
	FileLocate string
}

type Template struct {
	Name string
	Dir  string
}

func (database *Database) GenerateUrl() string {
	return database.Username + ":" + database.Password + "@" + database.Protocol + "(" + database.Address + ":" + database.Port + ")/" + database.DbName + "?charset=" + database.Charset
}

type PropertiesConfig struct {
	Properties      map[string]string
	PropertiesFiles []PropertiesFile
}

type PropertiesFile struct {
	Locate string
}
type ServerConfig struct {
	Port             string
	Token            string
	UseSSL           bool
	SSLConfig        SSLConfig
	Databases        []Database
	StaticResources  []StaticResource
	Templates        []Template
	PropertiesConfig PropertiesConfig
}
