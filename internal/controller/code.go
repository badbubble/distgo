package controller

type ReturnCode int64

const (
	CodeSuccess ReturnCode = 1000 + iota

	CodeServerBusy
)

var codeToMessage = map[ReturnCode]string{
	CodeSuccess:    "successful",
	CodeServerBusy: "The server is busy, please try again later.",
}

func (r ReturnCode) Message() string {
	msg, ok := codeToMessage[r]
	if !ok {
		msg = codeToMessage[CodeServerBusy]
	}
	return msg
}
