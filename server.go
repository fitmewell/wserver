package wserver

import (
	"log"
	"net/http"
	"sync"
)

func New(filePath string) (wServer *WServer, err error) {
	config, err := NewConfig(filePath)
	if err != nil {
		return nil, err
	}
	wServer = &WServer{
		config:config,
		sessionManager:NewDefaultSessionManager(config),
	}
	wServer.handler = NewDefaultHandler(wServer)
	return
}

type WServer struct {
	config         *ServerConfig
	handler        *handler
	context        ServerContext
	lock           sync.Mutex
	started        bool
	sessionManager SessionManager
}

func (wServer *WServer)Start() {
	wServer.lock.Lock()
	defer func() {
		wServer.started = false
		debug("end")
		wServer.lock.Unlock()
	}()

	if wServer.started {
		log.Fatal("Server already started")
	}
	wServer.started = true
	wServer.context = NewContextFrom(wServer.config)
	debug("started")
	wServer.Listen()
}
func (wServer *WServer) Listen() {

	var err error
	config := wServer.config
	if config.UseSSL {
		go func() {
			httpServerMux := http.NewServeMux()
			httpServerMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, "https:" + r.Host + ":443" + r.RequestURI, http.StatusMovedPermanently)
			})
			s := &http.Server{Addr: ":" + config.Port, Handler: httpServerMux}
			if err := s.ListenAndServe(); err != nil {
				log.Fatalf("ListenAndServe error: %v", err)
			}
		}()
		s := &http.Server{Addr: ":" + config.SSLConfig.SSLPort, Handler:wServer.handler}
		err = s.ListenAndServeTLS(config.SSLConfig.CertFile, config.SSLConfig.KeyFile)
	} else {
		s := &http.Server{Addr: ":" + config.Port, Handler:wServer.handler}
		err = s.ListenAndServe()
	}
	if err != nil {
		log.Print(err)
	}
}

func (wServer *WServer)AddHandler(method string, path string, e interface{}) *WServer {
	wServer.handler.addHandler(method, path, e)
	return wServer
}

func (wServer *WServer)AddAspectHandler(handler AspectHandler) *WServer {
	wServer.handler.addAspect(handler)
	return wServer
}