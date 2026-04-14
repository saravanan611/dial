package dial

import (
	"strings"

	"github.com/gorilla/mux"
	"github.com/saravanan611/log"
)

/* set ref func */
var (
	lOrginEnable func(pOrgin string) bool
	orgin        = "*"
	credflag     = false
	basepath     = "temp"
)

const (
	errCode     = "E"
	successCode = "S"
)

/*
it will change global variable,
use it in main.go file
*/

func SetOrginCheckFunc(pFunc func(pOrgin string) bool) error {
	if pFunc == nil {
		return log.Error("can't read the function")
	}
	lOrginEnable = pFunc
	return nil
}

type FTRouter struct {
	*mux.Router
	lRouterMap map[string]func(*Resp, *Request)
}

type FTRoute struct {
	*mux.Route
	fTHeaders, fTMethods []string
}

func SetOrgin(pOrgins ...string) {
	if len(pOrgins) != 0 {
		orgin = strings.Join(pOrgins, ",")
	}
}

func EnableCred() {
	credflag = true
}
