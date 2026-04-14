package dial

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/saravanan611/log"
)

/* ================== INIT ================== */

func NewRouter() *FTRouter {
	r := &FTRouter{Router: mux.NewRouter()}

	// Method not allowed
	r.MethodNotAllowedHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
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

			if w.Header().Get("Access-Control-Allow-Headers") == "" {
				w.Header().Set("Access-Control-Allow-Headers",
					"Accept,Content-Type,Content-Length,Accept-Encoding,X-CSRF-Token,Authorization")
			}

			w.Header().Set("Access-Control-Allow-Credentials", fmt.Sprint(credflag))

			if req.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}

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

	log.Debug("Server start %s %s .....", pType, pName)

	return http.Serve(listener, r.Router)
}

func (r *FTRouter) HandleFunc(path string, f func(*Resp, *Request)) *FTRoute {
	route := &FTRoute{}

	muxRoute := r.Router.HandleFunc(path, func(w http.ResponseWriter, req *http.Request) {

		// Apply per-route headers
		if len(route.fTHeaders) > 0 {
			w.Header().Set("Access-Control-Allow-Headers", strings.Join(route.fTHeaders, ","))
		}
		if len(route.fTMethods) > 0 {
			w.Header().Set("Access-Control-Allow-Methods", strings.Join(route.fTMethods, ","))
		}

		resp := &Resp{ResponseWriter: w, respType: req.Header.Get("Accept")}
		request := &Request{Request: req}

		f(resp, request)
	})

	route.Route = muxRoute
	return route
}

func (r *FTRoute) Methods(methods ...string) *FTRoute {
	r.fTMethods = append(methods, http.MethodOptions)

	if r.Route != nil {
		r.Route.Methods(methods...)
	}
	return r
}

func (r *FTRoute) SetHrdKey(keys ...string) *FTRoute {
	r.fTHeaders = append(keys,
		"Accept",
		"Content-Type",
		"Content-Length",
		"Accept-Encoding",
		"X-CSRF-Token",
		"Authorization",
	)
	return r
}
