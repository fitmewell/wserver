package wserver

import (
	"strings"
	"net/http"
	"reflect"
	"encoding/json"
	"io/ioutil"
	"encoding/xml"
	"log"
)

type handler struct {
	wServer     *WServer
	handlerTree handlerTree
}

func NewDefaultHandler(wServer *WServer) (h *handler) {
	h = &handler{wServer:wServer, handlerTree:NewDefaultHandlerTree()}
	for _, resource := range wServer.config.StaticResources {
		path := resource.Path
		if strings.HasSuffix(path, "**") {
			path = path[0:len(path) - 2]
		}
		if strings.HasSuffix(path, "*") {
			path = path[0:len(path) - 1]
		}
		t := http.StripPrefix(path, http.FileServer(http.Dir(resource.FileLocate)))
		h.addHandler("GET", resource.Path, func(context ServletContext, resp http.ResponseWriter, req *http.Request) error {
			t.ServeHTTP(resp, req)
			return nil
		})
	}
	return
}

func (h *handler)ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	tmp_session := h.wServer.sessionManager.Sync(resp, req)
	servletContext := &DefaultServletContext{ServerContext:h.wServer.context, Session:tmp_session, data:map[string]interface{}{}}
	if !h.handlerTree.AspectBefore(servletContext, resp, req) {
		return
	}
	var err error
	debug("METHOD:" + req.Method + "\tPATH:" + req.RequestURI)
	if ha := h.handlerTree.GetHandler(req); ha != nil {
		//err = ha(servletContext, resp, req)
		err = h.handle(servletContext, resp, req, ha)
		if err != nil {
			if e, ok := err.(*StatusError); ok {
				http.Error(resp, e.statusMessage, e.statusCode)
			} else {
				log.Print(err)
			}
			//wServer.handlerTree.HandlerError TODO  add common error handler here
		}
	} else {
		accept := req.Header.Get("Accept")
		if strings.Contains(accept, "text/html") {
			err := h.wServer.context.ExecuteTemplate(resp, "index.html", nil)
			if err != nil {
				log.Print(err)
			}
		} else {
			http.Error(resp, STATUS_NOT_FOUND.statusMessage, STATUS_NOT_FOUND.statusCode)
		}
	}
	if !h.handlerTree.AspectAfter(servletContext, resp, req) {
		return
	}
}

func (h *handler)addHandler(method string, path string, e interface{}) *handler {
	h.handlerTree.AddHandler(method, path, e)
	return h
}

func (h *handler)addAspect(a AspectHandler) *handler {
	h.handlerTree.AddAspect(a)
	return h
}

var (
	respKind reflect.Type = reflect.TypeOf((*http.ResponseWriter)(nil)).Elem()
	reqKind reflect.Type = reflect.TypeOf((*http.Request)(nil)).Elem()
	contextKind reflect.Type = reflect.TypeOf((*ServletContext)(nil)).Elem()
	errorKind reflect.Type = reflect.TypeOf((error)(nil))
)

func (h *handler)handle(context ServletContext, resp http.ResponseWriter, req *http.Request, m interface{}) (err error) {
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
			case respKind:
				inputs[i] = reflect.ValueOf(resp)
			case contextKind:
				inputs[i] = reflect.ValueOf(context)
			}
		case reflect.Struct:
			switch at {
			case reqKind:
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
							json.Unmarshal(b, t.Interface())
						case "TEXT/XML":
							fallthrough
						case "APPLICATION/XML":
							xml.Unmarshal(b, t.Interface())
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
		ot := out.Kind()
		switch ot {
		case reflect.String:
			path := out.Interface().(string)
			if e := context.ExecuteTemplate(resp, path, context.GetData()); e != nil {
				http.Redirect(resp, req, path, http.StatusFound)
			}
		case reflect.Interface:
			if out.Type() == errorKind {
				err = out.Interface().(error)
			}
		//todo
		case reflect.Struct:
			resp.Header().Set("Content-Type", "application/json")
			tb, err := json.Marshal(out.Interface())
			if err != nil {
				log.Print(err)
				resp.Write([]byte(err.Error()))
				break
			}
			resp.Write(tb)
		case reflect.Slice:
			switch out.Type().Elem().Kind(){
			case reflect.Uint8:
				resp.Write(out.Interface().([]byte))
			default:
				tb, err := json.Marshal(out.Interface())
				if err != nil {
					log.Print(err)
					resp.Write([]byte(err.Error()))
					break
				}
				resp.Write(tb)
			}
		}
	}
	return err
}