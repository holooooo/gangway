package settings

import "runtime"

var (
	ContextArch string = runtime.GOOS
	ContextType string
)

const (
	ContextTypeClient     = "client"
	ContextTypeController = "controller"
)
