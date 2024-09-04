package main

import (
	"encoding/json"
)

const CodeSuccess = 0
const CodeUserAlreadyRegistered = 1
const CodeUsernameTaken = 2
const CodeUsernameFormatNotValid = 3
const CodeReceiverDoesNotExist = 4
const CodeEmailProfileAuthInvariantBroken = 5
const CodeCantCreateAuthUser = 6
const CodeUserNotRegistered = 7
const CodeNotAuthenticated = 8
const CodeMaximumTokensNumberReached = 9
const CodeDeviceNameTooLong = 10
const CodeInvalidArgs = 11
const Code = 12 //next code

type ApiResponse[T any] struct {
	Result T `json:"result"`
}
type ApiResponseResult struct {
	Msg  string `json:"msg"`
	Code int    `json:"code"`
}
type ApiResponseResultWithUid struct {
	Msg  string `json:"msg"`
	Code int    `json:"code"`
	Uid  string `json:"uid"`
}
type ApiResponseResultWithJson struct {
	Msg  string      `json:"msg"`
	Code int         `json:"code"`
	Obj  interface{} `json:"obj"`
}

func apiResponse(msg string, code int) string {
	resp := ApiResponse[ApiResponseResult]{
		Result: ApiResponseResult{
			Msg:  msg,
			Code: code,
		},
	}

	b, err := json.Marshal(resp)
	if err != nil {
		return "{\"\"result:{}}" //shouldn't be happening
	}

	return string(b)
	//return fmt.Sprintf("{\"result\":{\"msg\":\"%s\", code:%d}}", msg, code)
}
func apiResponseWithJson(msg string, code int, obj interface{}) string {
	//return fmt.Sprintf("{result:{msg:\"%s\", code:%d, obj:%s}}", msg, code, jsonStr)
	resp := ApiResponse[ApiResponseResultWithJson]{
		Result: ApiResponseResultWithJson{
			Msg:  msg,
			Code: code,
			Obj:  obj,
		},
	}

	b, err := json.Marshal(resp)
	if err != nil {
		return "{\"result\":{}}" //shouldn't be happening
	}

	return string(b)
}
