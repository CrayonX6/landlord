package handler

import (
    "github.com/golang-module/carbon"
    "github.com/labstack/echo/v4"
    "net/http"
)

var ResponseHandler Response

type Response struct {
}

type ResponseStruct struct {
    Context  echo.Context `json:"-"`
    HttpCode int          `json:"-"`
    Code     int          `json:"code"`
    Message  string       `json:"message"`
    Data     interface{}  `json:"data"`
    Time     int64        `json:"time"`
}

func (r *Response) RS(c echo.Context) *ResponseStruct {
    return &ResponseStruct{
        Context:  c,
        HttpCode: http.StatusOK,
        Data:     echo.Map{},
        Time:     carbon.Now().ToTimestampWithMillisecond(),
    }
}

// error interface
func (rs *ResponseStruct) Error() string {
    return rs.Message
}

// 设置响应错误码
func (rs *ResponseStruct) SetHttpCode(code int) *ResponseStruct {
    rs.HttpCode = code
    return rs
}

// 设置逻辑错误码
func (rs *ResponseStruct) SetCode(code int) *ResponseStruct {
    rs.Code = code
    return rs
}

// 设置响应消息
func (rs *ResponseStruct) SetMessage(message string) *ResponseStruct {
    rs.Message = message
    return rs
}

// 设置响应数据
func (rs *ResponseStruct) SetData(data interface{}) *ResponseStruct {
    rs.Data = data
    return rs
}

// 追加响应数据
func (rs *ResponseStruct) AppendData(field string, value interface{}) *ResponseStruct {
    if rs.Data == nil {
        rs.Data = echo.Map{field: value}
    } else {
        rd := rs.Data.(echo.Map)
        rd[field] = value
        rs.Data = rd
    }
    return rs
}

// 向响应结构中追加错误详情
func (rs *ResponseStruct) ErrorDetail(err interface{}) *ResponseStruct {
    switch msg := err.(type) {
    case string:
        _ = rs.AppendData("err_detail", msg)
    case error:
        _ = rs.AppendData("err_detail", msg.Error())
    case echo.Map:
    case map[string]interface{}:
        _ = rs.AppendData("err_detail", msg)
    }
    return rs
}

// 获取响应结构
func (rs *ResponseStruct) GetStruct() error {
    return rs.Context.JSON(rs.HttpCode, rs)
}

// 响应错误
func (rs *ResponseStruct) ShowError(message string, err interface{}) error {
    return rs.SetMessage(message).ErrorDetail(err).GetStruct()
}

// 响应提示
func (rs *ResponseStruct) ShowMessage(message string) error {
    return rs.SetMessage(message).GetStruct()
}

// 响应数据
func (rs *ResponseStruct) ShowOkay(data interface{}) error {
    return rs.SetData(data).GetStruct()
}
