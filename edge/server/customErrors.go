package server

import (
	"edge/utils"
	"encoding/json"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CustomError struct {
	ErrorCode    int32
	ErrorMessage string
}

func NewCustomError(errorCode int32, msg string) error {
	return status.Error(codes.Unknown, buildJsonMessage(errorCode, msg))
}

func buildJsonMessage(errorCode int32, msg string) string {
	byteEncoding, _ := json.Marshal(
		CustomError{
			ErrorCode:    errorCode,
			ErrorMessage: msg,
		})
	stringEncoding := string(byteEncoding)
	utils.PrintEvent("CONVERSION_RESULT", stringEncoding)
	return stringEncoding
}
