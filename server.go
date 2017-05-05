package wserver

import (
	"errors"
	"github.com/fitmewell/wserver/session"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"
)

func New(filePath string) (wServer *WServer, err error) {
	debug("loading config file: " + filePath)
	config, err := NewConfig(filePath)
	if err != nil {
		return nil, err
	}
	wServer = &WServer{
		config:         config,
		sessionManager: session.NewDefaultSessionManager(config.Session.CookieName),
		aftermaths:     map[string]func(){},
	}
	wServer.handler = NewDefaultHandler(wServer)
	return
}

type WServer struct {
	config         *ServerConfig
	handler        *wHandler
	context        ServerContext
	lock           sync.Mutex
	started        bool
	sessionManager session.SessionManager
	aftermaths     map[string]func()
}

func (ws *WServer) Start() {
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

func (ws *WServer) listen() {

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
//'e' handler method , the server will auto handle the return value , the method parameter support *http.Request ,http.ResponseWriter, custom struct wserver context
func (ws *WServer) AddHandler(method string, path string, e interface{}) *WServer {
	ws.handler.addHandler(method, path, e)
	return ws
}

func (ws *WServer) AddAspectHandler(handler AspectHandler) *WServer {
	ws.handler.addAspect(handler)
	return ws
}

func (ws *WServer) aftermath() {
	s := make(chan os.Signal, 2)
	signal.Notify(s)
	go func() {
		switch <-s {
		case os.Interrupt:
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

func (ws *WServer) AddAftermath(name string, method func()) error {
	if _, ok := ws.aftermaths[name]; ok {
		return errors.New("duplicate aftermatch found")
	}
	ws.aftermaths[name] = method
	return nil
}
