package wserver

import (
	"net/http"
)

type handlerTree interface {
	AddHandler(method string, path string, handler interface{}) handlerTree
	AddAspect(AspectHandler) handlerTree
	GetHandler(*http.Request) interface{}
	AspectBefore(ServletContext, http.ResponseWriter, *http.Request) bool
	AspectAfter(ServletContext, http.ResponseWriter, *http.Request) bool
}

func NewDefaultHandlerTree() handlerTree {
	defaultTree := &defaultHandlerTree{}
	defaultTree.rootNode = &defaultHandlerTreeNode{tree:defaultTree, Path:"", PathKey:"", handlers: map[string]interface{}{},childNodes: map[string]handlerTreeNode{}}
	return defaultTree
}

type defaultHandlerTree struct {
	rootNode             handlerTreeNode
	beforeAspectHandlers []AspectHandler
	afterAspectHandlers  []AspectHandler
}

func (h *defaultHandlerTree)AddHandler(method string, path string, handler interface{}) handlerTree {
	h.rootNode.addHandler(method, path, handler)
	return h
}
func (h *defaultHandlerTree)AddAspect(handler AspectHandler) handlerTree {
	if handler.BeforeOrAfter() {
		h.beforeAspectHandlers = append(h.beforeAspectHandlers, handler)
	} else {
		h.afterAspectHandlers = append(h.afterAspectHandlers, handler)
	}
	return h
}
func (h *defaultHandlerTree)GetHandler(req *http.Request) interface{} {
	node := h.rootNode.getChild(req.RequestURI)
	if node == nil {
		return nil
	}
	return node.getHandler(req)
}
func (h *defaultHandlerTree)AspectBefore(serverContext ServletContext, resp http.ResponseWriter, req *http.Request) bool {
	for _, aspect := range h.beforeAspectHandlers {
		if aspect.ShouldAppendOn(req) {
			if !aspect.Server(serverContext, resp, req) {
				return false
			}
		}
	}
	return true
}
func (h *defaultHandlerTree)AspectAfter(serverContext ServletContext, resp http.ResponseWriter, req *http.Request) bool {
	for _, aspect := range h.afterAspectHandlers {
		if aspect.ShouldAppendOn(req) {
			if !aspect.Server(serverContext, resp, req) {
				return false
			}
		}
	}
	return true
}