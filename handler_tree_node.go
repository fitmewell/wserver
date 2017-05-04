package wserver

import (
	"errors"
	"net/http"
	"strings"
)

type handlerTreeNode interface {
	hasChild(path string) bool
	getChild(path string) handlerTreeNode
	newChild(key string) handlerTreeNode
	addChild(handler handlerTreeNode) (handlerTreeNode, error)
	addHandler(method string, path string, handler interface{}) (handlerTreeNode, error)
	getHandler(req *http.Request) interface{}
}

type defaultHandlerTreeNode struct {
	tree       handlerTree
	Path       string
	PathKey    string
	handlers   map[string]interface{}
	childNodes map[string]handlerTreeNode
}

func (t *defaultHandlerTreeNode) getChild(path string) handlerTreeNode {
	if len(path) != 0 && path[0] == '/' {
		path = path[1:]
	}
	if markI := strings.Index(path, "?"); markI > -1 {
		path = path[0:markI]
	}
	if path == "" || path == "/" {
		return t
	} else {
		key := path
		restPath := ""
		if i := strings.Index(key, "/"); i > -1 {
			key = path[0:i]
			restPath = path[i+1:]
		}
		if child, ok := t.childNodes[key]; ok {
			return child.getChild(restPath)
		} else if child, ok := t.childNodes["*"]; ok && !strings.Contains(restPath, "/") {
			return child
		} else if child, ok := t.childNodes["**"]; ok {
			return child
		} else {
			return nil
		}
	}
}

func (t *defaultHandlerTreeNode) getHandler(req *http.Request) interface{} {
	if executor, ok := t.handlers[req.Method]; ok {
		return executor
	}
	if executor, ok := t.handlers["*"]; ok {
		return executor
	}
	return nil
}

func (t *defaultHandlerTreeNode) hasChild(path string) bool {
	_, ok := t.childNodes[path]
	return ok
}

func (t *defaultHandlerTreeNode) newChild(key string) handlerTreeNode {
	if strings.HasPrefix(key, "/") {
		key = key[1:]
	}
	a := &defaultHandlerTreeNode{Path: t.Path + "/" + key, PathKey: key, handlers: map[string]interface{}{}, tree: t.tree, childNodes: map[string]handlerTreeNode{}}
	t.childNodes[key] = a
	return a
}

//todo match
func (t *defaultHandlerTreeNode) addChild(child handlerTreeNode) (handlerTreeNode, error) {
	return child, nil
}

func (t *defaultHandlerTreeNode) addHandler(method string, handlerPath string, handler interface{}) (handlerTreeNode, error) {
	var nextPath string
	if !strings.HasPrefix(handlerPath, "/") {
		handlerPath = "/" + handlerPath
	}
	if strings.Index(handlerPath, t.Path) != 0 {
		return nil, errors.New("Path root match failed")
	} else {
		nextPath = handlerPath[len(t.Path):]
	}
	if nextPath == "" || nextPath == "/" {
		if _, ok := t.handlers[method]; ok {
			return nil, errors.New("Duplicate executor found")
		}
		t.handlers[method] = handler
	} else {
		pathParts := strings.Split(nextPath, "/")
		var child handlerTreeNode = t
		for _, pathPart := range pathParts {
			tmp := child.getChild(pathPart)
			if tmp == nil {
				tmp = child.newChild(pathPart)
			}
			child = tmp
		}
		_, err := child.addHandler(method, handlerPath, handler)
		if err != nil {
			return t, err
		}
	}
	return t, nil
}
