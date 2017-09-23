package bdb

type SqlBuilder interface {
	// special table
	Table(name string) SqlBuilder
	// set value for update
	Set(key, value string) SqlBuilder
	// add eq statement
	Eq(name, value string) SqlBuilder
	// add a object for insert
	Append(value interface{}) SqlBuilder
	// add not eq statement
	NotEq(name, value string) SqlBuilder
	// add eq statement
	Like(name, value string) SqlBuilder
	// add not eq statement
	NotLike(name, value string) SqlBuilder
	// add in statement
	In(name string, values ...string) SqlBuilder
	// add not in statement
	NotIn(name string, values ...string) SqlBuilder
	// add between statement
	Between(name, start, end string) SqlBuilder
	// add nil judge statement
	IsNil(name string) SqlBuilder
	// add nil judge statement
	NotNil(name string) SqlBuilder
	// add order by sort
	OrderBy(name string, asc bool) SqlBuilder
}

type conditionType int

type stateType int

const (
	EQUAL    conditionType = iota // =
	NOTEQUAL                      // !=
	LIKE                          // LIKE
	NOT_LIKE                      // NOT LIKE
	IN                            // IN
	NOT_IN                        // NOT IN
	NIL                           // IS NULL
	NON_NIL                       // IS NOT NULL
	BETWEEN                       // BETWEEN ... AND ...
)

const (
	INSERT stateType = iota
	DELETE
	UPDATE
	SELECT
)

type condition struct {
	_type  conditionType
	name   string
	values []string
}

type WState struct {
	table      string
	columns    []string
	conditions []condition
	_type      stateType
	values     []interface{}
	value      map[string]string
}

func (w *WState) Append(value interface{}) SqlBuilder {
	w.values = append(w.values, value)
	return w
}

func (w *WState) Table(name string) SqlBuilder {
	w.table = name
	return w
}

func (w *WState) Set(key, value string) SqlBuilder {
	w.value[key] = value
	return w
}

func (w *WState) Like(name, value string) SqlBuilder {
	w.conditions = append(w.conditions, condition{
		_type:  LIKE,
		name:   name,
		values: []string{value},
	})
	return w
}

func (w *WState) NotLike(name, value string) SqlBuilder {
	w.conditions = append(w.conditions, condition{
		_type:  NOT_LIKE,
		name:   name,
		values: []string{value},
	})
	return w
}

func (w *WState) Eq(name, value string) SqlBuilder {
	w.conditions = append(w.conditions, condition{
		_type:  EQUAL,
		name:   name,
		values: []string{value},
	})
	return w
}

func (w *WState) NotEq(name, value string) SqlBuilder {
	w.conditions = append(w.conditions, condition{
		_type:  NOTEQUAL,
		name:   name,
		values: []string{value},
	})
	return w
}

func (w *WState) In(name string, values ...string) SqlBuilder {
	w.conditions = append(w.conditions, condition{
		_type:  IN,
		name:   name,
		values: values,
	})
	return w
}

func (w *WState) NotIn(name string, values ...string) SqlBuilder {
	w.conditions = append(w.conditions, condition{
		_type:  NOT_IN,
		name:   name,
		values: values,
	})
	return w
}

func (w *WState) Between(name, start, end string) SqlBuilder {
	w.conditions = append(w.conditions, condition{
		_type:  BETWEEN,
		name:   name,
		values: []string{start, end},
	})
	return w
}

func (w *WState) IsNil(name string) SqlBuilder {
	w.conditions = append(w.conditions, condition{
		_type: NIL,
		name:  name,
	})
	return w
}

func (w *WState) NotNil(name string) SqlBuilder {
	w.conditions = append(w.conditions, condition{
		_type: NON_NIL,
		name:  name,
	})
	return w
}

func (w *WState) OrderBy(name string, asc bool) SqlBuilder {
	return w
}

func Select(table string, columns ...string) SqlBuilder {
	return &WState{
		table:   table,
		columns: columns,
		_type:   SELECT,
	}
}

func Update(table string) SqlBuilder {
	return &WState{
		table: table,
		_type: UPDATE,
	}
}

func Delete(table string) SqlBuilder {
	return &WState{
		table: table,
		_type: DELETE,
	}
}

func Insert(table string, v interface{}) SqlBuilder {
	return &WState{
		table:  table,
		_type:  INSERT,
		values: []interface{}{v},
	}
}
