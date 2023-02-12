package vo

import (
	"fmt"
)

type RespBase[T any] struct {
	Code       int    `json:"code"`
	Msg        string `json:"msg"`
	ResultCode int    `json:"resultcode"`
	ResultMsg  string `json:"resultmsg"`
	Data       T      `json:"data"`
}

// SuccessRespBase 业务成功
func SuccessRespBase[T any](t T) RespBase[T] {
	return RespBase[T]{
		Code:       200,
		Msg:        "success",
		ResultCode: 1,
		ResultMsg:  "业务成功",
		Data:       t,
	}
}

// ParamInvalidRespBase 参数无效
func ParamInvalidRespBase(paramName string) RespBase[string] {
	return RespBase[string]{
		Code:       302,
		Msg:        fmt.Sprintf("参数错误: %s", paramName),
		ResultCode: 0,
		ResultMsg:  "业务失败",
	}
}

// BusinessFailedRespBase 业务失败
func BusinessFailedRespBase(resultMsg string) RespBase[string] {
	return RespBase[string]{
		Code:       200,
		Msg:        "success",
		ResultCode: 0,
		ResultMsg:  resultMsg,
	}
}
