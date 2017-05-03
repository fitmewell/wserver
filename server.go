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

func (ws *WServer)Start() {
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
	ws.listen()
}
func (ws *WServer) listen() {

	var err error
	config := ws.config
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
		s := &http.Server{Addr: ":" + config.SSLConfig.SSLPort, Handler:ws.handler}
		err = s.ListenAndServeTLS(config.SSLConfig.CertFile, config.SSLConfig.KeyFile)
	} else {
		s := &http.Server{Addr: ":" + config.Port, Handler:ws.handler}
		err = s.ListenAndServe()
	}
	if err != nil {
		log.Print(err)
	}
}
/**
'method' support method , use * to support all method
'path' path
'e' handler method , the server will auto handle the return value , the method parameter support *http.Request ,http.ResponseWriter, custom struct wserver context
 */
func (ws *WServer)AddHandler(method string, path string, e interface{}) *WServer {
	ws.handler.addHandler(method, path, e)
	return ws
}

func (ws *WServer)AddAspectHandler(handler AspectHandler) *WServer {
	ws.handler.addAspect(handler)
	return ws
}