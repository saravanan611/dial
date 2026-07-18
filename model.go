package dial

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
	"github.com/saravanan611/log"
)

/* set ref func */
var (
	lOrginEnable func(pOrgin string) bool
	orgin        = "*"
	credflag     = false
	basepath     = os.TempDir()
)

const (
	lReqIdKey   = "ReqIDKey"
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
func (r *FTRoute) Methods(methods ...string) *FTRoute {
	r.fTMethods = append(methods, http.MethodOptions)

	if r.Route != nil {
		r.Route.Methods(methods...)
	}
	return r
}

func (r *FTRoute) SetHrdKey(keys ...string) *FTRoute {
	r.fTHeaders = append(keys, lReqIdKey,
		"Accept",
		"Content-Type",
		"Content-Length",
		"Accept-Encoding",
		"X-CSRF-Token",
		"Authorization",
	)
	return r
}

/*
=======================================================================
Name            : UnGzipResp
Purpose         : This function is used to unzip the response body.
---------------------------------------------------------
Inputs          : pRespBody []byte

---------------------------------------------------------
Output          :

	string: response body after unzipping and check non text/plain

---------------------------------------------------------

Author          : Saravanan selvam
Created Date    : 24/06/2026
=======================================================================
*/
func UnGzipResp(pRespBody []byte) string {
	// Buffer to store unzipped response
	var lBuffer bytes.Buffer
	// Create a new gzip reader
	lReadData, lErr := gzip.NewReader(bytes.NewBuffer(pRespBody))
	if lErr != nil {
		return fmt.Sprintf("error on process unzipping %v", lErr)
	}
	// Close the gzip reader
	defer lReadData.Close()
	// Copy the unzipped data to the buffer
	_, lErr = lBuffer.ReadFrom(lReadData)
	if lErr != nil {
		return fmt.Sprintf("error on process unzipping %v", lErr)
	}
	// Get the unzipped data
	lData := lBuffer.Bytes()
	// Check if the unzipped data is not plain text
	if !strings.Contains(strings.ToLower(http.DetectContentType(lData)), "text/plain") {
		return "response body contains non plain text or file after unzipping"
	}
	// Return the unzipped data as a string
	return string(lData)
}
