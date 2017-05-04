# wserver
A golang http framework 

To use this module you the sample below
`sample.go`
```go
package main

import (
	"github.com/fitmewell/wserver"
	"log"
)

func main() {
	s, e := wserver.New("config.json")
	if e != nil {
		log.Fatal(e)
	}
	s.AddHandler("*", "/", func() []byte {
		return []byte("helloworld")
	}).Start()
}
```
`config.json`
```json
{
  "Port": "8080",
  "UseSSL": false,
  "SSLConfig": { //necessary when using https 
    "SSLPort": "443",
    "CertFile": "~/key/fullchain.pem",
    "KeyFile": "~/key/privkey.pem"
  },
  "Databases": [
    {
      "IsDefault": true,
      "Name": "RW",
      "Address": "127.0.0.1",
      "Port": "3306",
      "Protocol": "tcp",
      "Username": "test",
      "Password": "test",
      "Charset": "utf8",
      "DBName": "test",
      "MaxConnections": 30,
      "DriverName": "mysql"//remember to import db driver
    }
  ],
  "StaticResources": [
    {
      "Name": "HttpsAutoUpdate",
      "Path": "/.well-known/**",
      "FileLocate": "./.well-known/"
    },
    {
      "Name": "StaticFile",
      "Path": "/static/**",
      "FileLocate": "./static/"
    }
  ],
  "Templates": [
    {
      "Name": "default",
      "Dir": "./template"
    }
  ],
  "DefaultPage": {
    "HomePage": "index.html"
  },
  "PropertiesConfig": {
    "Properties": {
      "Hello": "test"
    }
  }
}
```
