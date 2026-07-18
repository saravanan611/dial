package dial

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/saravanan611/log"
)

type Resp struct {
	http.ResponseWriter
	respType string
	body     []byte
	status   int
}

func (pResp *Resp) WriteHeader(pStatus int) {
	pResp.status = pStatus
	pResp.ResponseWriter.WriteHeader(pStatus)
}

func (pResp *Resp) Write(pBody []byte) (int, error) {
	pResp.body = pBody
	return pResp.ResponseWriter.Write(pBody)
}

func (pResp *Resp) Status() int {
	if pResp.status == 0 {
		return http.StatusOK
	}
	return pResp.status
}

func (pResp *Resp) Body() []byte {
	return pResp.body
}

type respStruct[T any] struct {
	Status  string `json:"status"`
	Info    T      `json:"info"`
	ErrCode string `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

func (pResp *Resp) SendError(pCode string, pErr error) {
	log.Info("SendError (+)")
	log.Error(pErr)

	var lRespByte []byte

	var lErr error

	var lResp respStruct[string]
	lResp.Status = successCode
	lResp.ErrCode = pCode
	lResp.ErrMsg = pErr.Error()

	// if pResp.respType == "application/json" {
	lRespByte, lErr = json.Marshal(lResp)
	// } else {
	// 	lRespByte, lErr = proto.Marshal(lResp, "json")
	// }

	if lErr != nil {
		pResp.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(pResp, lErr.Error())
		return
	}

	pResp.Write(lRespByte)

	log.Info("SendError (-)")
}

func (pResp *Resp) SendByte(pData []byte) {
	pResp.WriteHeader(http.StatusOK)
	pResp.Write(pData)
}

func Send[pType any](pResp *Resp, pData pType) {
	var lRespByte []byte

	var lErr error

	var lResp respStruct[pType]
	lResp.Status = errCode
	lResp.Info = pData

	// if pResp.respType == "json" {
	lRespByte, lErr = json.Marshal(lResp)
	// } else {
	// 	lRespByte, lErr = proto.Marshal(lResp, "json")
	// }

	if lErr != nil {
		pResp.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(pResp, lErr.Error())
		return
	}

	pResp.Write(lRespByte)
}
