package wserver

import (
	"reflect"
	"regexp"
	"strconv"
)

func CheckDto(dto interface{}, path string) []string {
	v := reflect.ValueOf(dto)
	t := reflect.TypeOf(dto)
	path += "."
	if t.Kind() == reflect.Ptr {
		v = v.Elem()
		t = t.Elem()
	}
	es := []string{}
	for i := 0; i < v.NumField(); i++ {
		childV := v.Field(i)
		childT := t.Field(i)
		if childV.Kind() != reflect.String && childV.Kind() != reflect.Struct {
			if p := childT.Tag.Get("notNull"); p == "true" && childV.IsNil() {
				es = append(es, path+childT.Name+" shouldn't be null")
			}
		}
		if childT.PkgPath != "" {
			continue
		}
		switch childV.Kind() {
		case reflect.Func:
			continue
		case reflect.Interface:
			fallthrough
		case reflect.Struct:
			if childV.IsValid() {
				es = append(CheckDto(childV.Interface(), path+childT.Name), es...)
			}
		case reflect.Array:
			fallthrough
		case reflect.Slice:
			if childT.Type.Elem().Kind() == reflect.Struct || childT.Type.Elem().Kind() == reflect.Interface {
				for j := 0; j < childV.Cap(); j++ {
					es = append(CheckDto(childV.Index(j).Interface(), path+childT.Name+"["+strconv.Itoa(j)+"]"), es...)
				}
			}
		case reflect.String:
			if p := childT.Tag.Get("pattern"); p != "" {
				match, err := regexp.MatchString(p, childV.String())
				if err != nil || !match {
					es = append(es, path+childT.Name+" shouldn match pattern:"+p)
				}
			}
			if p := childT.Tag.Get("notEmpty"); p == "true" && childV.String() == "" {
				es = append(es, path+childT.Name+" shouldn't be empty")
			}
		case reflect.Int:
			fallthrough
		case reflect.Int8:
			fallthrough
		case reflect.Int16:
			fallthrough
		case reflect.Int32:
			fallthrough
		case reflect.Int64:
			//todo
		}
	}
	return es
}
