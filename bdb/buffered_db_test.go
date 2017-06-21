package bdb

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"testing"
)

var db BufferedDB

func init() {
	d, err := sql.Open("mysql", "")
	if err != nil {
		log.Fatal(err)
	}
	db = NewBufferedDb(d)
}

func TestMap(t *testing.T) {
	var map1 map[interface{}]interface{}
	var map2 **map[string]string
	var map3 *map[string]interface{}
	var err error
	err = db.SelectInInterface("SHOW COLUMNS FROM policy_table", &map1)
	if err != nil {
		t.Error(err)
	}
	t.Log(map1)
	err = db.SelectInInterface("SHOW COLUMNS FROM policy_table", &map2)
	if err != nil {
		t.Error(err)
	}
	t.Log(*map2)
	err = db.SelectInInterface("SHOW COLUMNS FROM policy_table", &map3)
	if err != nil {
		t.Error(err)
	}
	t.Log(map3)
}

func TestList(t *testing.T) {
	var list1 []interface{}
	var list2 []map[string]interface{}
	var list3 *[]map[string]string
	var err error
	err = db.SelectInInterface("SHOW COLUMNS FROM policy_table", &list1)
	if err != nil {
		t.Error(err)
	}
	t.Log(list1)
	err = db.SelectInInterface("SHOW COLUMNS FROM policy_table", &list2)
	if err != nil {
		t.Error(err)
	}
	t.Log(list2)
	err = db.SelectInInterface("SHOW COLUMNS FROM policy_table", &list3)
	if err != nil {
		t.Error(err)
	}
	t.Log(list3)
}

func TestInterface(t *testing.T) {
	var v1 interface{}
	var v2 *interface{}
	var v3 **interface{}
	var err error
	err = db.SelectInInterface("SHOW COLUMNS FROM policy_table", &v1)
	if err != nil {
		t.Error(err)
	}
	t.Log(v1)
	err = db.SelectInInterface("SHOW COLUMNS FROM policy_table", &v2)
	if err != nil {
		t.Error(err)
	}
	t.Log(*v2)
	err = db.SelectInInterface("SHOW COLUMNS FROM policy_table", &v3)
	if err != nil {
		t.Error(err)
	}
	t.Log(**v3)
}
