package dial

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

type Request struct {
	*http.Request
	body                         []byte
	ForwardIP, FullPath, ReailIp string
}

func (r *Request) ReadAll() {

	if r.Header.Get("x-Orignal-Forwarded-For") != "" {
		r.ForwardIP = r.Header.Get("x-Orignal-Forwarded-For")
	} else if r.Header.Get("X-Forwarded-For") != "" {
		r.ForwardIP = r.Header.Get("X-Forwarded-For")
	} else if r.Header.Get("x-Real-Ip") != "" {
		r.ForwardIP = r.Header.Get("x-Real-Ip")
	} else {
		r.ForwardIP = r.RemoteAddr
	}

	if r.URL.RawQuery != "" {
		r.FullPath = r.URL.Path + "?" + r.URL.RawQuery
	} else {
		r.FullPath = r.URL.Path
	}

	r.ReailIp = r.Header.Get("Referer")

	r.body, _ = io.ReadAll(r.Body)
	r.Body = io.NopCloser(bytes.NewBuffer(r.body))

}

func (r *Request) Read(data any) (err error) {

	// if r.Header.Get("Content-Type") == "application/json" {
	return json.Unmarshal(r.body, data)
	// }

	// return proto.Unmarshal(r.body, data, "json")
}
