package comm

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

const (
	CodeSuccess                         = 0
	CodeUserAlreadyRegistered           = 1
	CodeUsernameTaken                   = 2
	CodeUsernameFormatNotValid          = 3
	CodeReceiverDoesNotExist            = 4
	CodeEmailProfileAuthInvariantBroken = 5
	CodeCantCreateAuthUser              = 6
	CodeUserNotRegistered               = 7
	CodeNotAuthenticated                = 8
	CodeMaximumTokensNumberReached      = 9
	CodeDeviceNameTooLong               = 10
	CodeInvalidArgs                     = 11
	Code                                = 12 //next code
)

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
type ApiResponseResultWith[T any] struct {
	Msg  string `json:"msg"`
	Code int    `json:"code"`
	Obj  T      `json:"obj"`
}

type ApiResponseWith[T any] ApiResponse[ApiResponseResultWith[T]]
type ApiResponseWithJson ApiResponse[ApiResponseResultWithJson]
type ApiResponsePlain ApiResponse[ApiResponseResult]

func (resp *ApiResponse[T]) String() string {
	b, err := json.Marshal(resp)
	if err != nil {
		return "{\"\"result:{}}" //shouldn't be happening
	}

	return string(b)
}

func NewApiResponse(msg string, code int) *ApiResponsePlain {
	resp := ApiResponsePlain{
		Result: ApiResponseResult{
			Msg:  msg,
			Code: code,
		},
	}

	return &resp
}
func NewApiResponseWithJson(msg string, code int, obj interface{}) *ApiResponseWithJson {
	resp := ApiResponseWithJson{
		Result: ApiResponseResultWithJson{
			Msg:  msg,
			Code: code,
			Obj:  obj,
		},
	}

	return &resp
}

func OK(ctx *gin.Context, msg string, code int) {
	ctx.JSON(http.StatusOK, NewApiResponse(msg, code))
}
func GenericOK(ctx *gin.Context) {
	OK(ctx, "Success", CodeSuccess)
}
func OKJSON(ctx *gin.Context, msg string, code int, obj any) {
	ctx.JSON(http.StatusOK, NewApiResponseWithJson(msg, code, obj))
}
func GenericOKJSON(ctx *gin.Context, obj any) {
	OKJSON(ctx, "Success", CodeSuccess, obj)
}
func AbortBadRequest(ctx *gin.Context, msg string, code int) {
	ctx.AbortWithStatusJSON(http.StatusBadRequest, NewApiResponse(msg, code))
}
func AbortBadRequestJSON(ctx *gin.Context, msg string, code int, obj any) {
	ctx.AbortWithStatusJSON(http.StatusBadRequest, NewApiResponseWithJson(msg, code, obj))
}
func AbortGenericInvalidArgs(ctx *gin.Context) {
	ctx.AbortWithStatusJSON(http.StatusBadRequest, NewApiResponse("Invalid args", CodeInvalidArgs))
}
func AbortUnauthorized(ctx *gin.Context, msg string, code int) {
	ctx.AbortWithStatusJSON(http.StatusUnauthorized, NewApiResponse(msg, code))
}
func AbortFailedBinding(ctx *gin.Context, err error) {
	if err == nil {
		return
	}
	respMsg := "Invalid args"
	if errs, ok := err.(validator.ValidationErrors); ok {
		respMsg = fmt.Sprintf("Failed validation: \n%s", errs.Error())
	}

	AbortBadRequest(ctx, respMsg, CodeInvalidArgs)
}
