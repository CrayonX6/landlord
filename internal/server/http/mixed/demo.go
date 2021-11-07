package mixed

import (
    "beginning/internal/pkg/acme"
    "beginning/pkg/handler"
    "github.com/labstack/echo/v4"
)

var CDemo Demo

type Demo struct {
    handler.Response
    acme.UserInfo
    acme.Acme
}

// Demo
func (cc *Demo) GetDemo(c echo.Context) error {
    _ = cc.ParseUserInfo(c)
    return cc.RS(c).ShowMessage("DONE")
}
