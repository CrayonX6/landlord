package mixed

import (
    "beginning/internal/pkg/acme"
    "beginning/internal/pkg/config"
    "beginning/pkg/handler"
    "github.com/labstack/echo/v4"
)

var CConfig Config

type Config struct {
    handler.Response
    acme.UserInfo
}

// 列表配置
func (cc *Config) GetConfigure(c echo.Context) error {
    _ = cc.ParseUserInfo(c)
    var list = make(map[string]config.Cnf)
    for key, cnf := range config.VBswConfig.ListConfig() {
        if !cnf.AllowClientPull {
            continue
        }
        list[key] = cnf
    }
    return cc.RS(c).ShowOkay(list)
}
