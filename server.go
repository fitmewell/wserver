package wserver

import (
	"errors"
	"github.com/fitmewell/wserver/wsession"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"
)

func New(filePath string) (wServer *Server, err error) {
	debug("loading config file: " + filePath)
	config, err := NewConfig(filePath)
	if err != nil {
		return nil, err
	}
	wServer = &Server{
		config:         config,
		sessionManager: wsession.NewDefaultSessionManager(config.Session.CookieName),
		aftermaths:     map[string]func(){},
	}
	wServer.handler = NewDefaultHandler(wServer)
	return
}

func NewPortServer(port string) (wServer *Server) {
	config := &ServerConfig{Port: port}
	wServer = &Server{
		config:         config,
		sessionManager: wsession.NewDefaultSessionManager(config.Session.CookieName),
		aftermaths:     map[string]func(){},
	}
	wServer.handler = NewDefaultHandler(wServer)
	return
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

	ws.handler.init()

	ws.lock.Lock()
	defer func() {
		ws.started = false
		debug("end")
		ws.lock.Unlock()
	}()

	if ws.started {
		log.Fatal("Server already started")
	}
	ws.started = true
	ws.context = NewContextFrom(ws.config)
	debug("started")
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
				log.Fatalf("ListenAndServe error: %v", err)
			}
		}()
		s := &http.Server{Addr: ":" + config.SSLConfig.SSLPort, Handler: ws.handler}
		err = s.ListenAndServeTLS(config.SSLConfig.CertFile, config.SSLConfig.KeyFile)
	} else {
		s := &http.Server{Addr: ":" + config.Port, Handler: ws.handler}
		err = s.ListenAndServe()
	}
	if err != nil {
		log.Print(err)
	}
}

//'method' support method , use * to support all method
//'path' path
//'e' handler method , the server will auto handle the return value , the method parameter support *http.Request ,http.ResponseWriter, custom struct server context
func (ws *Server) AddHandler(method string, path string, e interface{}) *Server {
	ws.handler.addHandler(method, path, e)
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
		cs := <-s
		debug("caught system signal:" + cs.String())
		switch cs {
		case os.Interrupt:
			fallthrough
		case os.Kill:
			debug("closing")
			ws.started = false

			go func() {
				time.Sleep(5 * time.Second)
				debug("end")
				ws.lock.Unlock()
				os.Exit(1)
			}()
			c := make(chan string, len(ws.aftermaths))
			for k, v := range ws.aftermaths {
				go func(name string, method func()) {
					debug("[aftermath][" + name + "]start")
					method()
					c <- name
				}(k, v)
			}

			amount := 0
			for {
				k := <-c
				debug("[aftermath][" + k + "]end")
				amount++
				if amount == len(ws.aftermaths) {
					debug("end")
					ws.lock.Unlock()
					os.Exit(1)
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
