package dial

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/saravanan611/log"
	"github.com/saravanan611/proto"
)

type Request struct {
	*http.Request
}

func (r *Request) Read(data any) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return log.Error(err)
	}

	if r.Header.Get("Content-Type") == "application/json" {
		return json.Unmarshal(body, data)
	}

	return proto.Unmarshal(body, data, "json")
}
