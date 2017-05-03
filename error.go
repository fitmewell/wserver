package wserver

import "reflect"

type StatusError struct {
	statusCode    int
	statusMessage string
}

var statusErrorType = reflect.TypeOf(&StatusError{})

func (s *StatusError)Error() string {
	return s.statusMessage
}

var (
	/* not error status , just record here
	STATUS_CONTINUE = &StatusError{statusCode:100, statusMessage:"Continue"}
	STATUS_SWITCHING_PROTOCOLS = &StatusError{statusCode:101, statusMessage:"Switching Protocols"}

	STATUS_OK = &StatusError{statusCode:200, statusMessage:"OK"}
	STATUS_CREATED = &StatusError{statusCode:201, statusMessage:"Created"}
	STATUS_ACCEPTED = &StatusError{statusCode:202, statusMessage:"Accepted"}
	STATUS_NON_AUTHORITATIVE_INFORMATION = &StatusError{statusCode:203, statusMessage:"Non-Authoritative Information"}
	STATUS_NO_CONTENT = &StatusError{statusCode:204, statusMessage:"No Content"}
	STATUS_RESET_CONTENT = &StatusError{statusCode:205, statusMessage:"Reset Content"}
	STATUS_PARTIAL_CONTENT = &StatusError{statusCode:206, statusMessage:"Partial Content"}

	STATUS_MULTIPLE_CHOICES = &StatusError{statusCode:300, statusMessage:"Multiple Choices"}
	STATUS_MOVED_PERMANENTLY = &StatusError{statusCode:301, statusMessage:"Moved Permanently"}
	STATUS_FOUND = &StatusError{statusCode:302, statusMessage:"Found"}
	STATUS_SEE_OTHER = &StatusError{statusCode:303, statusMessage:"See Other"}
	STATUS_NOT_MODIFIED = &StatusError{statusCode:304, statusMessage:"Not Modified"}
	STATUS_USE_PROXY = &StatusError{statusCode:305, statusMessage:"Use Proxy"}
	STATUS_UNUSED = &StatusError{statusCode:306, statusMessage:"(Unused)"}
	STATUS_TEMPORARY_REDIRECT = &StatusError{statusCode:307, statusMessage:"Temporary Redirect"}
	*/

	STATUS_BAD_REQUEST = &StatusError{statusCode:400, statusMessage:"Bad Request"}
	STATUS_UNAUTHORIZED = &StatusError{statusCode:401, statusMessage:"Unauthorized"}
	STATUS_PAYMENT_REQUIRED = &StatusError{statusCode:402, statusMessage:"Payment Required"}
	STATUS_FORBIDDEN = &StatusError{statusCode:403, statusMessage:"Forbidden"}
	STATUS_NOT_FOUND = &StatusError{statusCode:404, statusMessage:"Not Found"}
	STATUS_METHOD_NOT_ALLOWED = &StatusError{statusCode:405, statusMessage:"Method Not Allowed"}
	STATUS_NOT_ACCEPTABLE = &StatusError{statusCode:406, statusMessage:"Not Acceptable"}
	STATUS_PROXY_AUTHENTICATION_REQUIRED = &StatusError{statusCode:407, statusMessage:"Proxy Authentication Required"}
	STATUS_REQUEST_TIMEOUT = &StatusError{statusCode:408, statusMessage:"Request Timeout"}
	STATUS_CONFLICT = &StatusError{statusCode:409, statusMessage:"Conflict"}
	STATUS_GONE = &StatusError{statusCode:410, statusMessage:"Gone"}
	STATUS_LENGTH_REQUIRED = &StatusError{statusCode:411, statusMessage:"Length Required"}
	STATUS_PRECONDITION_FAILED = &StatusError{statusCode:412, statusMessage:"Precondition Failed"}
	STATUS_REQUEST_ENTITY_TOO_LARGE = &StatusError{statusCode:413, statusMessage:"Request Entity Too Large"}
	STATUS_REQUEST_URI_TOO_LONG = &StatusError{statusCode:414, statusMessage:"Request-URI Too Long"}
	STATUS_UNSUPPORTED_MEDIA_TYPE = &StatusError{statusCode:415, statusMessage:"Unsupported Media Type"}
	STATUS_REQUESTED_RANGE_NOT_SATISFIABLE = &StatusError{statusCode:416, statusMessage:"Requested Range Not Satisfiable"}
	STATUS_EXPECTATION_FAILED = &StatusError{statusCode:417, statusMessage:"Expectation Failed"}

	STATUS_INTERNAL_SERVER_ERROR = &StatusError{statusCode:500, statusMessage:"Internal Server Error"}
	STATUS_NOT_IMPLEMENTED = &StatusError{statusCode:501, statusMessage:"Not Implemented"}
	STATUS_BAD_GATEWAY = &StatusError{statusCode:502, statusMessage:"Bad Gateway"}
	STATUS_SERVICE_UNAVAILABLE = &StatusError{statusCode:503, statusMessage:"Service Unavailable"}
	STATUS_GATEWAY_TIMEOUT = &StatusError{statusCode:504, statusMessage:"Gateway Timeout"}
	STATUS_HTTP_VERSION_NOT_SUPPORTED = &StatusError{statusCode:505, statusMessage:"HTTP Version Not Supported"}
)