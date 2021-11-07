package hook

import (
    "bytes"
    tool "beginning/pkg/acme"
    "beginning/pkg/log"
    "fmt"
    "github.com/golang-module/carbon"
    "io/ioutil"
    "math"
    "net/http"
    "strings"

    "github.com/labstack/echo/v4"
    "github.com/labstack/echo/v4/middleware"
)

var DefaultSignatureConfig = SignatureConfig{
    Skipper:           middleware.DefaultSkipper,
    SignKey:           "Signature",
    Salt:              "echo@signature#x",
    TimeKey:           "_time",
    TimeMaxDiffSecond: 30,
}

type SignatureConfig struct {
    Skipper           middleware.Skipper
    SignKey           string
    PassRoute         []string
    Salt              string
    TimeKey           string
    TimeMaxDiffSecond int64
}

func SignatureMiddleware() echo.MiddlewareFunc {
    return SignatureWithConfig(DefaultSignatureConfig)
}

// SignatureWithConfig 按配置签名
func SignatureWithConfig(config SignatureConfig) echo.MiddlewareFunc {
    if config.Skipper == nil {
        config.Skipper = DefaultSignatureConfig.Skipper
    }
    if len(config.SignKey) == 0 {
        config.SignKey = DefaultSignatureConfig.SignKey
    }
    if len(config.Salt) == 0 {
        config.Salt = DefaultSignatureConfig.Salt
    }
    if len(config.TimeKey) == 0 {
        config.TimeKey = DefaultSignatureConfig.TimeKey
    }
    if config.TimeMaxDiffSecond == 0 {
        config.TimeMaxDiffSecond = DefaultSignatureConfig.TimeMaxDiffSecond
    }

    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(ctx echo.Context) (err error) {
            if config.Skipper(ctx) {
                return next(ctx)
            }

            req := ctx.Request()
            part := fmt.Sprintf("%s:%s", strings.ToUpper(req.Method), ctx.Path())
            if exists, _ := tool.InArray(part, config.PassRoute); exists {
                return next(ctx)
            }

            var args map[string]interface{}
            var signature = req.Header.Get(config.SignKey)

            if req.Method == http.MethodGet {
                args = tool.ParseUrlQuery(req.RequestURI)
            } else {
                bodyBytes, _ := ioutil.ReadAll(req.Body)
                req.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
                tool.JsonToInterface(string(bodyBytes), &args)
            }

            args, _signature := tool.DigMapField(args, config.SignKey)
            if len(tool.ToStr(_signature)) > 0 {
                signature = tool.ToStr(_signature)
            }

            if len(signature) == 0 {
                message := fmt.Sprintf("校验签名参数有误 (%s:%s, Client:%s, Args:%s)", req.Method, req.RequestURI, signature, tool.InterfaceToJson(args))
                log.Logger.Error(message)
                return &echo.HTTPError{
                    Code:     http.StatusUnprocessableEntity,
                    Message:  "校验签名参数为空",
                    Internal: err,
                }
            }

            var signTime int64 = 0
            if _, ok := args[config.TimeKey]; ok {
                signTime = int64(tool.ToInt(args[config.TimeKey]))
            }
            timeDiff := carbon.Now().ToTimestampWithSecond() - signTime
            if math.Abs(float64(timeDiff)) > float64(config.TimeMaxDiffSecond) {
                return &echo.HTTPError{
                    Code:     http.StatusUnprocessableEntity,
                    Message:  fmt.Sprintf("客户端时间与服务端时间差异大于%d秒", config.TimeMaxDiffSecond),
                    Internal: err,
                }
            }

            signStr, signMd5 := tool.MapToSignature(args, config.Salt, config.TimeKey, " is ", " and ", " and salt is ")
            if strings.Compare(signMd5, signature) != 0 {
                message := fmt.Sprintf("校验签名失败 (%s:%s, Client:%s, Server:%s, SignatureStr:%s, Args:%s)", req.Method, req.RequestURI, signature, signMd5, signStr, tool.InterfaceToJson(args))
                log.Logger.Error(message)
                return &echo.HTTPError{
                    Code:     http.StatusForbidden,
                    Message:  "校验签名失败,访问终止",
                    Internal: err,
                }
            }
            return next(ctx)
        }
    }
}
