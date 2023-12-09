package response

type Response struct {
	IsSuccess bool        `json:"isSuccess"`
	Data      interface{} `json:"data"`
	Message   interface{} `json:"message"`
}

type SignUpLoginResponse struct {
	ID    int    `json:"id"`
	Token string `json:"token"`
}

type RecordReponse struct {
	ID int `json:"id"`
}

func New() *Response {
	return &Response{
		IsSuccess: false,
		Data:      nil,
		Message:   nil,
	}
}
