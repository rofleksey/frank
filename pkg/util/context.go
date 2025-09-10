package util

type ContextKey string

func (c ContextKey) String() string {
	return "frank_" + string(c)
}

var UsernameContextKey ContextKey = "username"
var IpContextKey ContextKey = "ip"
