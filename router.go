package dial

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/saravanan611/log"
)

/* ================== INIT ================== */

func NewRouter() *FTRouter {
	r := &FTRouter{Router: mux.NewRouter()}

	// Method not allowed
	r.MethodNotAllowedHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "Method <%s> not allowed for <%s>", req.Method, req.URL.Path)
	})

	// Global middleware
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

			log.SetRequestID(strings.ReplaceAll(uuid.NewString(), "-", ""))
			defer log.ClearRequestID()

			// CORS
			if lOrginEnable != nil && lOrginEnable(req.RemoteAddr) {
				w.Header().Set("Access-Control-Allow-Origin", req.RemoteAddr)
			}

			if w.Header().Get("Access-Control-Allow-Origin") == "" {
				w.Header().Set("Access-Control-Allow-Origin", orgin)
			}

			w.Header().Set("Access-Control-Allow-Credentials", fmt.Sprint(credflag))

			log.Info("Middleware (+) %s %s", req.Method, req.URL.Path)

			next.ServeHTTP(w, req)

			log.Info("Middleware (-)")
		})
	})

	return r
}

func (r *FTRouter) Start(pType, pName string) error {

	if r.Router == nil {
		return log.Error("router is nil")
	}

	if pName == "" {
		return log.Error("port number required || socket name is required")
	}

	switch pType {
	case "unix":
		pName = filepath.Join(basepath, pName+".sock")
		_ = os.Remove(pName)
		defer os.Remove(pName)

	case "tcp":
		if pName == "" {
			return log.Error("port number required")
		}
	default:
		return log.Error("invalid type")
	}

	listener, err := net.Listen(pType, pName)
	if err != nil {
		return err
	}
	defer listener.Close()

	log.Debug(" < %s > Server start :%s .....", pType, pName)

	return http.Serve(listener, r.Router)
}

func (r *FTRouter) HandleFunc(path string, f func(*Resp, *Request)) *FTRoute {
	route := &FTRoute{}

	muxRoute := r.Router.HandleFunc(path, func(w http.ResponseWriter, req *http.Request) {

		// Apply per-route headers
		if len(route.fTHeaders) == 0 {
			route.SetHrdKey()
		}

		if len(route.fTMethods) == 0 {
			route.Methods()
		}
		w.Header().Set("Access-Control-Allow-Headers", strings.Join(route.fTHeaders, ","))
		w.Header().Set("Access-Control-Allow-Methods", strings.Join(route.fTMethods, ","))

		if req.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		resp := &Resp{ResponseWriter: w, respType: req.Header.Get("Accept")}
		request := &Request{Request: req}
		request.ReadAll()

		// handle panic
		defer func() {
			if r := recover(); r != nil {
				resp.SendError("500", fmt.Errorf("panic: %v", r))
			}
		}()

		var lInOut struct {
			ReqDateTime, RespDateTime                                                                                                time.Time
			Type, Duration, RealIP, ForwardedIP, Method, Path, Host, RemoteAddr, Header, Endpoint, ReqBody, RespBody, ResponseStatus string
		}

		lInOut.ReqDateTime = time.Now()

		defer func() {
			if req.Method != http.MethodOptions {
				w.Header().Set(lReqIdKey, log.GetRequestID())
				lInOut.Type = "In and Out"
				lInOut.RespDateTime = time.Now()
				lInOut.Duration = time.Since(lInOut.ReqDateTime).String()

				lInOut.Header = fmt.Sprintf("%+v", req.Header)
				lInOut.Endpoint = req.URL.Path
				lInOut.Host = req.Host
				lInOut.RemoteAddr = req.RemoteAddr
				lInOut.Method = req.Method
				lInOut.ReqBody = string(request.body)
				lInOut.RespBody = string(resp.body)
				lInOut.ResponseStatus = fmt.Sprint(resp.Status())
				lInOut.ForwardedIP = request.ForwardIP
				lInOut.RealIP = request.ReailIp
				lInOut.Path = request.FullPath

				if !strings.Contains(strings.ToLower(http.DetectContentType([]byte(lInOut.ReqBody))), "text/plain") {
					// if the content type is not multipart/form-data
					if !strings.Contains(strings.ToLower(req.Header.Get("Content-Type")), "multipart/form-data") {
						// if the content encoding is gzip
						if req.Header.Get("Content-Encoding") == "gzip" {
							// unzip the request body
							lInOut.ReqBody = UnGzipResp([]byte(lInOut.ReqBody))
						} else {
							// set the request body to non plain text or file
							lInOut.ReqBody = "request body contains non plain text or file"
						}
					} else {
						lInOut.ReqBody = "request body contains non plain text or file"
					}
				}

				// if the response body contains non plain text or file
				if !strings.Contains(strings.ToLower(http.DetectContentType([]byte(lInOut.RespBody))), "text/plain") {
					// if the content encoding is gzip
					if w.Header().Get("Accept-Encoding") == "gzip" {
						// unzip the response body
						lInOut.RespBody = UnGzipResp([]byte(lInOut.RespBody))
					} else {
						lInOut.RespBody = "response body contains non plain text or file"
					}
				}

				log.Debug("%+v", lInOut)
			}

		}()

		f(resp, request)
	})

	route.Route = muxRoute
	return route
}
