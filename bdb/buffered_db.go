package bdb

import (
	"database/sql"
	"log"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type BufferedState interface {
	//execute the sequence by combine sqlState and parameters , all the state will be stored
	ExecutePreparedSql(sqlState string, parameters ...interface{}) (sql.Result, error)

	//use #{} as symbol in sequence and get related value from interface in and return the query result
	ExecuteDbSequence(sequence string, in interface{}) (sql.Result, error)

	//use parameters to fill the symbol ? in sqlSequence ,and auto fill the interface v
	SelectInInterface(sqlSequence string, v interface{}, parameters ...interface{}) error

	//use #{} as symbol in sequence and get related value from interface in , and fill the result to interface out
	SelectInInterfaceAuto(sequence string, out interface{}, in interface{}) error
}

type BufferedDB interface {
	//extend from buffered state
	BufferedState

	//return a BufferedState with transaction
	BeginTransactional() (BufferedTransactional, error)
}

type BufferedTransactional interface {
	//extend from buffered state
	BufferedState

	//commit transaction
	Commit() error

	//rollback transaction
	Rollback() error
}

const default_time_pattern = "2006-01-02 15:04:05"

var timeType = reflect.TypeOf(time.Time{})

type defaultBdb struct {
	*sql.DB
	preparedStmtMap map[string]*sql.Stmt
}

type defaultBtx struct {
	*sql.Tx
}

func NewBufferedDb(db *sql.DB) BufferedDB {
	return &defaultBdb{db, map[string]*sql.Stmt{}}
}

func (tdb *defaultBdb) SelectInInterface(sqlSequence string, v interface{}, parameters ...interface{}) error {
	sqlStmt, err := tdb.getPreparedStatement(sqlSequence)
	if err != nil {
		return err
	}
	return selectInInterface(sqlStmt, v, parameters...)
}

func (tdb *defaultBdb) getPreparedStatement(sqlSentence string) (*sql.Stmt, error) {
	if preparedStmt, ok := tdb.preparedStmtMap[sqlSentence]; ok {
		return preparedStmt, nil
	} else {
		sqlStmt, err := tdb.Prepare(sqlSentence)
		if err != nil {
			return nil, err
		}
		if tdb.preparedStmtMap == nil {
			tdb.preparedStmtMap = make(map[string]*sql.Stmt)
		}
		tdb.preparedStmtMap[sqlSentence] = sqlStmt
		return sqlStmt, nil
	}
}

func (tdb *defaultBdb) ExecutePreparedSql(sqlState string, parameters ...interface{}) (sql.Result, error) {
	sqlStmt, err := tdb.getPreparedStatement(sqlState)
	if err != nil {
		return nil, err
	}
	return sqlStmt.Exec(parameters...)
}

func (tdb *defaultBdb) ExecuteDbSequence(sequence string, parameter interface{}) (sql.Result, error) {
	sequence, parameters := parseInput(sequence, parameter)
	return tdb.ExecutePreparedSql(sequence, parameters...)
}

func (tdb *defaultBdb) SelectInInterfaceAuto(sequence string, out interface{}, in interface{}) error {
	sequence, parameters := parseInput(sequence, in)
	return tdb.SelectInInterface(sequence, out, parameters...)
}

func (tdb *defaultBdb) BeginTransactional() (BufferedTransactional, error) {
	t, e := tdb.Begin()
	return &defaultBtx{t}, e
}

func (btx *defaultBtx) SelectInInterface(sqlSequence string, v interface{}, parameters ...interface{}) error {
	sqlStmt, err := btx.getPreparedStatement(sqlSequence)
	if err != nil {
		return err
	}
	return selectInInterface(sqlStmt, v, parameters...)
}

func (btx *defaultBtx) getPreparedStatement(sqlSentence string) (*sql.Stmt, error) {
	return btx.Prepare(sqlSentence)
}

func (btx *defaultBtx) ExecutePreparedSql(sqlState string, parameters ...interface{}) (sql.Result, error) {
	sqlStmt, err := btx.getPreparedStatement(sqlState)
	if err != nil {
		return nil, err
	}
	return sqlStmt.Exec(parameters...)
}

func (btx *defaultBtx) ExecuteDbSequence(sequence string, parameter interface{}) (sql.Result, error) {
	sequence, parameters := parseInput(sequence, parameter)
	return btx.ExecutePreparedSql(sequence, parameters...)
}

func (btx *defaultBtx) SelectInInterfaceAuto(sequence string, out interface{}, in interface{}) error {
	sequence, parameters := parseInput(sequence, in)
	return btx.SelectInInterface(sequence, out, parameters...)
}

func (btx *defaultBtx) Commit() error {
	return btx.Tx.Commit()
}

func (btx *defaultBtx) Rollback() error {
	return btx.Tx.Rollback()
}

func parseInput(sequence string, parameter interface{}) (string, []interface{}) {
	r, err := regexp.Compile("([#$])\\{([^}]+)}")
	if err != nil {
		return "", nil
	}
	tmpParameters := r.FindAllStringSubmatch(sequence, -1)
	parameters := []interface{}{}
	for _, param := range tmpParameters {
		parameters = append(parameters, getParameterValues(parameter, param[2]))
	}
	return r.ReplaceAllString(sequence, "?"), parameters
}

func getParameterValues(object interface{}, names string) interface{} {
	t := reflect.TypeOf(object)
	v := reflect.ValueOf(object)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}
	name := names
	restName := ""
	splitIndex := strings.Index(names, ".")
	if splitIndex > 0 {
		name = names[0:splitIndex]
		restName = names[splitIndex+1:]
	}
	name = strings.ToUpper(name[0:1]) + name[1:]

	tmpV := v.FieldByName(name)
	if tmpV.Kind() == reflect.Ptr {
		tmpV = tmpV.Elem()
	}
	if splitIndex > 0 {
		return getParameterValues(tmpV.Interface(), restName)
	} else {
		return tmpV.Interface()
	}
}

func selectInInterface(sqlStmt *sql.Stmt, v interface{}, parameters ...interface{}) (err error) {
	var rows *sql.Rows

	if parameters == nil || len(parameters) == 0 {
		rows, err = sqlStmt.Query()
	} else {
		rows, err = sqlStmt.Query(parameters...)
	}

	if err != nil {
		return err
	}

	rowColumns, err := rows.Columns()
	if err != nil {
		return err
	}

	typeOfBean := reflect.TypeOf(v)
	valueOfBean := reflect.ValueOf(v)
	var actualType reflect.Type
	for typeOfBean.Kind() == reflect.Ptr {
		typeOfBean = typeOfBean.Elem()
		valueOfBean = valueOfBean.Elem()
	}
	isArray := typeOfBean.Kind() == reflect.Array || typeOfBean.Kind() == reflect.Slice
	if isArray {
		actualType = typeOfBean.Elem()
	} else {
		actualType = typeOfBean
	}

	var valueCache = make([][]byte, len(rowColumns))
	var interfaceCache = make([]interface{}, len(rowColumns))
	columnCache, err := getMatchedColumns(actualType, rowColumns)
	if err != nil {
		return err
	}
	for i := range valueCache {
		interfaceCache[i] = &valueCache[i]
	}

	for rows.Next() {
		err := rows.Scan(interfaceCache...)
		if err != nil {
			return err
		}
		var valueOfInnerBean reflect.Value
		if isArray {
			valueOfInnerBean = reflect.New(actualType).Elem()
		} else {
			valueOfInnerBean = valueOfBean
		}
		for i, value := range valueCache {
			if columnIndex := columnCache[i]; columnIndex != -1 {
				columnField := valueOfInnerBean.Field(columnIndex)
				columnFieldKind := columnField.Type().Kind()
				switch columnFieldKind {
				case reflect.String:
					columnField.SetString(string(value))
				case reflect.Int:
					fallthrough
				case reflect.Int64:
					fallthrough
				case reflect.Int32:
					intValue, err := strconv.ParseInt(string(value), 10, 0)
					if err != nil {
						return err
					}
					columnField.SetInt(intValue)
				case reflect.Bool:
					boolValue, err := strconv.ParseBool(string(value))
					if err != nil {
						return err
					}
					columnField.SetBool(boolValue)
				case reflect.Float32:
					fallthrough
				case reflect.Float64:
					floatValue, err := strconv.ParseFloat(string(value), 0)
					if err != nil {
						return err
					}
					columnField.SetFloat(floatValue)
				case reflect.Struct:
					switch columnField.Type() {
					case timeType:
						//Mon Jan 2 15:04:05 -0700 MST 2006
						timePattern := actualType.Field(columnIndex).Tag.Get("pattern")
						if value == nil || len(value) == 0 {
							continue
						}
						if timePattern == "" {
							timePattern = default_time_pattern
						}
						timeValue, err := time.Parse(timePattern, string(value))
						if err != nil {
							return err
						}
						columnField.Set(reflect.ValueOf(timeValue))
					default:
						log.Print(columnField.Type())
					}
				default:
					log.Print("new Kind found not matched :" + columnFieldKind.String())
				}
			} else {
				continue
			}
		}
		if isArray == false {
			return err
		} else {
			n := valueOfBean.Len()
			if n >= valueOfBean.Cap() {
				c := 2 * n
				if c < 4 {
					c = 4
				}
				newSlice := reflect.MakeSlice(typeOfBean, n, c)
				reflect.Copy(newSlice, valueOfBean)
				valueOfBean.Set(newSlice)
			}
			valueOfBean.SetLen(n + 1)
			valueOfBean.Index(n).Set(valueOfInnerBean)
		}
	}
	return nil
}

func getMatchedColumns(actualType reflect.Type, rowColumns []string) (columnCache []int, err error) {
		columnCache = make([]int, len(rowColumns))
		for i, columnName := range rowColumns {
			for j := 0; j < actualType.NumField(); j++ {
				filedI := actualType.Field(j)
				columnIndex := filedI.Tag.Get("column")
				if columnIndex != "" {
					intValue, err := strconv.ParseInt(string(columnIndex), 10, 0)
					columnCache[int(intValue)] = j
					if err != nil {
						return columnCache, err
					}
				} else {
					filedColumnName := filedI.Name
					if filedI.Tag.Get("name") != "" {
						filedColumnName = filedI.Tag.Get("name")
					} else {
						filedColumnName = strings.ToLower(strings.Replace(filedColumnName, "_", "", -1))
					}
					columnName = strings.ToLower(strings.Replace(columnName, "_", "", -1))
					if filedColumnName == columnName {
						columnCache[i] = j
						break
					} else {
						columnCache[i] = -1
					}
				}
			}
		}
	return columnCache, err
}
