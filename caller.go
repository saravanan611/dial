package dial

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/saravanan611/log"
)

func GetClient(rawURL string) (string, string, *http.Client, error) {
	lUrl, lErr := url.Parse(rawURL)
	if lErr != nil {
		return "", "", nil, lErr
	}

	host := lUrl.Hostname() // removes port if any
	if strings.HasSuffix(host, ".localhost") {
		parts := strings.Split(host, ".")

		transport := &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return net.Dial("unix", filepath.Join(basepath, parts[0]+".sock"))
			},
		}
		return "unix", "http://unix" + lUrl.RequestURI(), &http.Client{Transport: transport}, nil
	}

	return "tcp", rawURL, &http.Client{}, nil
}

func RawCall[T any](pUrl, pMethod string, pReqMap map[string]string, pData T) (lResp *http.Response, lErr error) {

	var lInOut struct {
		ReqDateTime, RespDateTime                                              time.Time
		Type, Duration, Method, URL, Header, ReqBody, RespBody, ResponseStatus string
	}

	lInOut.ReqDateTime = time.Now()

	defer func() {
		lInOut.Type = "External api call"
		lInOut.RespDateTime = time.Now()
		lInOut.Duration = time.Since(lInOut.ReqDateTime).String()
		lInOut.Method = pMethod
		lInOut.URL = pUrl
		lInOut.Header = fmt.Sprintf("%+v", pReqMap)
		lInOut.ReqBody = fmt.Sprintf("%+v", pData)
		log.Debug("%+v", lInOut)
	}()

	_, lUrl, lClient, lErr := GetClient(pUrl)
	if lErr != nil {
		return
	}

	var lReqByte []byte

	if lData, lOk := any(pData).(string); lOk {
		lReqByte = []byte(lData)
		// } else if lType == "unix" {
		// 	lReqByte, lErr = proto.Marshal(pData, "json")
		// 	if lErr != nil {
		// 		return
		// 	}
	} else {
		lReqByte, lErr = json.Marshal(pData)
		if lErr != nil {
			return
		}
	}

	lReq, lErr := http.NewRequest(pMethod, lUrl, bytes.NewBuffer(lReqByte))
	if lErr != nil {
		return
	}
	defer lReq.Body.Close()

	for k, v := range pReqMap {
		lReq.Header.Set(k, v)
	}

	lResp, lErr = lClient.Do(lReq)
	lInOut.ResponseStatus = fmt.Sprintf("%+v", lResp.Status)
	lRespByte, _ := io.ReadAll(lResp.Body)
	lInOut.RespBody = string(lRespByte)
	lResp.Body = io.NopCloser(bytes.NewBuffer(lRespByte))

	return

}

func Call[T, R any](pUrl, pMethod string, pReqMap map[string]string, pData T) (lResp R, lErr error) {

	lRespData, lErr := RawCall(pUrl, pMethod, pReqMap, pData)
	if lErr != nil {
		return
	}

	defer lRespData.Body.Close()

	lRespBody, lErr := io.ReadAll(lRespData.Body)
	if lErr != nil {
		return
	}

	// if strings.Contains(strings.ToLower(pReqMap["Accept"]), "application/json") {
	if lErr = json.Unmarshal(lRespBody, &lResp); lErr != nil {
		return
	}
	return

	// }
	// var lProtoResp respStruct[R]

	// if lErr = proto.Unmarshal(lRespBody, &lProtoResp, "json"); lErr != nil {
	// 	return lProtoResp.Info, lErr
	// }
	// if lProtoResp.Status != successCode {
	// 	return lProtoResp.Info, lErr
	// }
	// return lProtoResp.Info, nil

}
