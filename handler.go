package wserver

import (
	"encoding/json"
	"encoding/xml"
	. "github.com/fitmewell/wserver/log"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
)

type wHandler struct {
	wServer     *Server
	handlerTree handlerTree
}

func newDefaultHandler(wServer *Server) (h *wHandler) {
	h = &wHandler{wServer: wServer, handlerTree: newDefaultHandlerTree()}
	return
}

func (h *wHandler) init() {
	for _, resource := range h.wServer.config.StaticResources {
		path := resource.Path
		if strings.HasSuffix(path, "**") {
			path = path[0 : len(path)-2]
		}
		if strings.HasSuffix(path, "*") {
			path = path[0 : len(path)-1]
		}
		t := http.StripPrefix(path, http.FileServer(http.Dir(resource.FileLocate)))
		DebugF("Severing static files: %s %s", path, resource.FileLocate)
		h.addHandler("GET", resource.Path, func(context ServletContext, resp http.ResponseWriter, req *http.Request) error {
			t.ServeHTTP(resp, req)
			return nil
		})
	}
}

func (h *wHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	tmp_session := h.wServer.sessionManager.Sync(resp, req)
	servletContext := &DefaultServletContext{ServerContext: h.wServer.context, Session: tmp_session, data: map[string]interface{}{}}
	if !h.handlerTree.AspectBefore(servletContext, resp, req) {
		return
	}
	var err error
	Debug("METHOD:" + req.Method + "\tPATH:" + req.RequestURI)
	if ha := h.handlerTree.GetHandler(req); ha != nil {
		//err = ha(servletContext, resp, req)
		err = h.handle(servletContext, resp, req, ha)
		if err != nil {
			if e, ok := err.(*StatusError); ok {
				switch e.statusCode {
				case STATUS_UNAUTHORIZED.statusCode:
					http.Redirect(resp, req, "/", 301)
				default:
					http.Error(resp, e.statusMessage, e.statusCode)
				}
			} else {
				Debug(err)
			}
			//wServer.handlerTree.HandlerError TODO  add common error handler here
		}
	} else {
		accept := req.Header.Get("Accept")
		if strings.Contains(accept, "text/html") {
			http.Error(resp, "Page not found", STATUS_NOT_FOUND.statusCode)
		} else {
			http.Error(resp, STATUS_NOT_FOUND.statusMessage, STATUS_NOT_FOUND.statusCode)
		}
	}
	if !h.handlerTree.AspectAfter(servletContext, resp, req) {
		return
	}
}

func (h *wHandler) addHandler(method string, path string, e interface{}) *wHandler {
	h.handlerTree.AddHandler(method, path, e)
	return h
}

func (h *wHandler) addAspect(a AspectHandler) *wHandler {
	h.handlerTree.AddAspect(a)
	return h
}

var (
	respType    reflect.Type = reflect.TypeOf((*http.ResponseWriter)(nil)).Elem()
	reqType     reflect.Type = reflect.TypeOf((*http.Request)(nil)).Elem()
	contextType reflect.Type = reflect.TypeOf((*ServletContext)(nil)).Elem()
	errorType   reflect.Type = reflect.TypeOf((*error)(nil)).Elem()
)

/*

 */
func (h *wHandler) handle(context ServletContext, resp http.ResponseWriter, req *http.Request, m interface{}) (err error) {
	t := reflect.TypeOf(m)
	v := reflect.ValueOf(m)
	n := t.NumIn()
	inputs := make([]reflect.Value, n)
	for i := 0; i < n; i++ {
		at := t.In(i)
		for at.Kind() == reflect.Ptr {
			at = at.Elem()
		}
		switch at.Kind() {
		case reflect.Interface:
			switch at {
			case respType:
				inputs[i] = reflect.ValueOf(resp)
			case contextType:
				inputs[i] = reflect.ValueOf(context)
			}
		case reflect.Struct:
			switch at {
			case reqType:
				inputs[i] = reflect.ValueOf(req)
			default:
				//todo add custom struct parse
				t := reflect.New(at)
				switch req.Method {
				case "POST":
					contentType := strings.ToUpper(req.Header.Get("Content-Type"))
					if contentType != "" {
						typeDetail := strings.Split(contentType, ";")
						b, err := ioutil.ReadAll(req.Body)
						if err != nil {
							return err
						}
						switch typeDetail[0] {
						case "TEXT/JSON":
							fallthrough
						case "APPLICATION/JSON":
							if err := json.Unmarshal(b, t.Interface()); err != nil {
								return err
							}
						case "TEXT/XML":
							fallthrough
						case "APPLICATION/XML":
							if err := xml.Unmarshal(b, t.Interface()); err != nil {
								return err
							}
						}
					}
				}
				for t.Kind() == reflect.Ptr {
					t = t.Elem()
				}
				inputs[i] = t
			}

		}
	}
	outs := v.Call(inputs)
	for _, out := range outs {
		for out.Kind() == reflect.Ptr {
			out = out.Elem()
		}
		ot := out.Kind()
		switch ot {
		case reflect.String:
			path := out.Interface().(string)
			if e := context.ExecuteTemplate(resp, path, context.GetData()); e != nil {
				Debug(e.Error())
				http.Redirect(resp, req, path, http.StatusFound)
			}
		case reflect.Interface:
			if out.IsNil() {
				continue
			}
			if out.Type().Implements(errorType) {
				err = out.Interface().(error)
			}
			//todo
		case reflect.Struct:
			resp.Header().Set("Content-Type", "application/json")
			tb, err := json.Marshal(out.Interface())
			if err != nil {
				Debug(err)
				resp.Write([]byte(err.Error()))
				break
			}
			resp.Write(tb)
		case reflect.Slice:
			switch out.Type().Elem().Kind() {
			case reflect.Uint8:
				resp.Write(out.Interface().([]byte))
			default:
				tb, err := json.Marshal(out.Interface())
				if err != nil {
					Debug(err)
					resp.Write([]byte(err.Error()))
					break
				}
				resp.Write(tb)
			}
		}
	}
	return err
}
