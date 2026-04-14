package dial

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/saravanan611/proto"
)

func GetClient(rawURL string) (string, string, *http.Client, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", "", nil, err
	}

	host := u.Hostname() // removes port if any
	if strings.HasSuffix(host, ".localhost") {
		parts := strings.Split(host, ".")

		transport := &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return net.Dial("unix", filepath.Join(basepath, parts[0]+".sock"))
			},
		}

		lUrl := "http://unix" + u.RequestURI()
		return "unix", lUrl, &http.Client{Transport: transport}, nil
	}

	return "tcp", rawURL, &http.Client{}, nil
}

func Call[T any, R any](pUrl, pMethod string, pReqMap map[string]string, pData T) (lResp R, lErr error) {

	lType, lUrl, lClient, lErr := GetClient(pUrl)
	if lErr != nil {
		return lResp, lErr
	}

	var lReqByte []byte

	if lData, lOk := any(pData).(string); lOk {
		lReqByte = []byte(lData)
	} else if lType == "unix" {
		lReqByte, lErr = proto.Marshal(pData, "json")
		if lErr != nil {
			return lResp, lErr
		}
	} else {
		lReqByte, lErr = json.Marshal(pData)
		if lErr != nil {
			return lResp, lErr
		}
	}

	lReq, lErr := http.NewRequest(pMethod, lUrl, bytes.NewBuffer(lReqByte))
	if lErr != nil {
		return lResp, lErr
	}
	defer lReq.Body.Close()

	for k, v := range pReqMap {
		lReq.Header.Set(k, v)
	}

	lRespByte, lErr := lClient.Do(lReq)
	if lErr != nil {
		return lResp, lErr
	}
	defer lRespByte.Body.Close()

	lRespBody, lErr := io.ReadAll(lRespByte.Body)
	if lErr != nil {
		return lResp, lErr
	}

	if strings.Contains(strings.ToLower(pReqMap["Accept"]), "application/json") {
		if lErr = json.Unmarshal(lRespBody, &lResp); lErr != nil {
			return lResp, lErr
		}
		return lResp, nil

	}
	var lProtoResp respStruct[R]

	if lErr = proto.Unmarshal(lRespBody, &lProtoResp, "json"); lErr != nil {
		return lProtoResp.Info, lErr
	}
	if lProtoResp.Status != successCode {
		return lProtoResp.Info, lErr
	}
	return lProtoResp.Info, nil

}
