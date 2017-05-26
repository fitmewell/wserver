package wlog

import "log"

func Debug(v ...interface{}) {
	v = append([]interface{}{"[DEBUG] "}, v...)
	log.Print(v...)
}

func DebugF(format string, v ...interface{}) {
	log.Printf("[DEBUG] "+format, v...)
}

func Fatal(v ...interface{}) {
	log.Fatal(v...)
}

func FatalF(format string, v ...interface{}) {
	log.Fatalf(format, v...)
}
