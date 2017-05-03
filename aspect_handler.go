package wserver

import (
	"net/http"
	"strings"
)

/**
   A handler for aop
 */
type AspectHandler interface {
	ShouldAppendOn(req *http.Request) bool
	Server(ServletContext, http.ResponseWriter, *http.Request) bool
	BeforeOrAfter() bool
}

type DefaultAspectHandler struct {
	CustomCheck func(req *http.Request) bool
	Execute     func(ServletContext, http.ResponseWriter, *http.Request) bool
	MatchPath   string
	ExcludePath []string
	StrictMatch bool
	PositionFlg bool
}

func (defaultAspectHandler *DefaultAspectHandler)ShouldAppendOn(req *http.Request) bool {
	if defaultAspectHandler.CustomCheck != nil {
		return defaultAspectHandler.CustomCheck(req)
	}
	path := strings.Split(req.RequestURI, "?")[0]
	if defaultAspectHandler.StrictMatch {
		return defaultAspectHandler.MatchPath == path
	} else {
		if len(defaultAspectHandler.ExcludePath) != 0 {
			for _, excludePath := range defaultAspectHandler.ExcludePath {
				if strings.HasPrefix(path, excludePath) {
					return false
				}
			}
		}
		if strings.HasPrefix(path, defaultAspectHandler.MatchPath) {
			return true
		}
	}
	return false
}

func (defaultAspectHandler *DefaultAspectHandler)BeforeOrAfter() bool {
	return defaultAspectHandler.PositionFlg
}
func (defaultAspectHandler *DefaultAspectHandler)Server(serverContext ServletContext, resp http.ResponseWriter, req *http.Request) bool {
	return defaultAspectHandler.Execute(serverContext, resp, req)
}