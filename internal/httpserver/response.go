package httpserver

type Response struct {
	Code int `json:"code"` //业务状态 0 成功 -1 失败
	Msg string `json:"msg"`//提示信息
	Data any `json:"data,omitempty"`//返回数据
}
const(
	codeOK = 0
	codeFail = -1
)
func OK(data any) Response{
	return Response{
		Code: codeOK,
		Msg: "ok",
		Data: data,
	}
}

func Fail(msg string) Response{
	return Response{
		Code: codeFail,
		Msg: msg,
	}
}