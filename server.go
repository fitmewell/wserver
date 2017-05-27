package wserver

import (
	"errors"
	. "github.com/fitmewell/wserver/log"
	"github.com/fitmewell/wserver/wsession"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func New(filePath string) (wServer *Server, err error) {
	Debug("loading config file: " + filePath)
	config, err := NewConfig(filePath)
	if err != nil {
		return nil, err
	}
	return NewServer(config), nil
}

func NewPortServer(port string) (wServer *Server) {
	config := &ServerConfig{Port: port}
	return NewServer(config)
}
func NewServer(config *ServerConfig) *Server {
	server := &Server{
		config:         config,
		sessionManager: wsession.NewDefaultSessionManager(config.Session.CookieName),
		aftermaths:     map[string]func(){},
	}
	server.handler = newDefaultHandler(server)
	server.context = NewContextFrom(server.config)
	return server
}

type Server struct {
	config         *ServerConfig
	handler        *wHandler
	context        ServerContext
	lock           sync.Mutex
	started        bool
	sessionManager wsession.SessionManager
	aftermaths     map[string]func()
}

func (ws *Server) Start() {

	ws.lock.Lock()
	defer func() {
		ws.started = false
		Debug("end")
		ws.lock.Unlock()
	}()

	if ws.started {
		Fatal("Server already started")
	}
	ws.started = true
	ws.context.Init()
	ws.handler.init()
	Debug("started")
	ws.aftermath()
	ws.listen()
}

func (ws *Server) listen() {

	var err error
	config := ws.config
	if config.UseSSL {
		go func() {
			httpServerMux := http.NewServeMux()
			httpServerMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, "https:"+r.Host+":443"+r.RequestURI, http.StatusMovedPermanently)
			})
			s := &http.Server{Addr: ":" + config.Port, Handler: httpServerMux}
			if err := s.ListenAndServe(); err != nil {
				FatalF("ListenAndServe error: %v", err)
			}
		}()
		s := &http.Server{Addr: ":" + config.SSLConfig.SSLPort, Handler: ws.handler}
		err = s.ListenAndServeTLS(config.SSLConfig.CertFile, config.SSLConfig.KeyFile)
	} else {
		s := &http.Server{Addr: ":" + config.Port, Handler: ws.handler}
		err = s.ListenAndServe()
	}
	if err != nil {
		Debug(err)
	}
}

//'method' support method , use * to support all method
//'path' path
//'e' handler method , the server will auto handle the return value , the method parameter support *http.Request ,http.ResponseWriter, custom struct server context
func (ws *Server) AddHandler(method string, path string, e interface{}) *Server {
	ws.handler.addHandler(method, path, e)
	DebugF("Listening:[%s]\t[%s]\t{%v}", path, method, e)
	return ws
}

func (ws *Server) AddStaticSource(path, fileLocate string) *Server {
	ws.config.StaticResources = append(ws.config.StaticResources, StaticResource{Path: path, FileLocate: fileLocate})
	return ws
}

func (ws *Server) AddTemplate(name, dir string) *Server {
	ws.config.Templates = append(ws.config.Templates, Template{Name: name, Dir: dir})
	return ws
}

func (ws *Server) AddAspectHandler(handler AspectHandler) *Server {
	ws.handler.addAspect(handler)
	return ws
}

func (ws *Server) aftermath() {
	s := make(chan os.Signal, 2)
	signal.Notify(s)
	go func() {
		for {
			cs := <-s
			Debug("caught system signal:" + cs.String())
			switch cs {
			case os.Interrupt:
				fallthrough
			case syscall.SIGTERM:
				fallthrough
			case os.Kill:
				Debug("closing")
				ws.started = false

				go func() {
					time.Sleep(5 * time.Second)
					Debug("end")
					ws.lock.Unlock()
					os.Exit(1)
				}()
				c := make(chan string, len(ws.aftermaths))
				for k, v := range ws.aftermaths {
					go func(name string, method func()) {
						Debug("[aftermath][" + name + "]start")
						method()
						c <- name
					}(k, v)
				}

				amount := 0
				for {
					k := <-c
					Debug("[aftermath][" + k + "]end")
					amount++
					if amount == len(ws.aftermaths) {
						Debug("end")
						ws.lock.Unlock()
						os.Exit(1)
					}
				}
			}
		}
	}()
}

func (ws *Server) AddAftermath(name string, method func()) error {
	if _, ok := ws.aftermaths[name]; ok {
		return errors.New("duplicate aftermatch found")
	}
	ws.aftermaths[name] = method
	return nil
}

func (ws *Server) GetProperties(i string) string {
	return ws.context.GetProperty(i)
}
