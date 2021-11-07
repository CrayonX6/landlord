package user

import (
    "beginning/internal/entity"
    "beginning/internal/pkg/acme"
    tool "beginning/pkg/acme"
    "beginning/pkg/db"
    "beginning/pkg/handler"
    "github.com/labstack/echo/v4"
    "github.com/spf13/viper"
)

var ActionUserVar ActionUser

type ActionUser struct {
    handler.Response
    acme.UserInfo
    acme.Acme
}

type LoginArgs struct {
    Code string `json:"code"`
}

func (cc *ActionUser) getToken(u entity.User) interface{} {
    var UMap map[string]interface{}
    tool.AlignStructAndMap(u, &UMap)
    token := cc.TokenCreator(UMap)
    return token
}

func (cc *ActionUser) ShowUserMap(u entity.User) map[string]interface{} {
    return echo.Map{
        "uid":        u.ID,
        "open_id":    u.OpenId,
        "head_img":   u.HeadImg,
        "nickname":   u.Nickname,
        "gender":     u.Gender,
        "bean":       u.Bean,
        "play_times": u.PlayTimes,
        "win_times":  u.WinTimes,
        "token":      cc.getToken(u),
        "cnf":        viper.GetStringMap("cnf"),
    }
}

func (cc *ActionUser) PostLogin(c echo.Context) error {
    args := LoginArgs{}
    _ = c.Bind(&args)

    r, err := cc.JsCode2Session(args.Code)
    if err != nil {
        return cc.RS(c).SetCode(422).ShowError("登录失败", err)
    }
    if _, ok := r["errcode"]; ok {
        return cc.RS(c).SetCode(422).ShowError("登录失败", r["errmsg"])
    }

    openId := r["openid"]
    var u entity.User
    result := db.Orm.Where("open_id = ?", openId).First(&u)
    if result == nil {
        return cc.RS(c).SetCode(500).ShowMessage("读取用户失败")
    }

    if u.ID > 0 {
        if u.State == 0 {
            return cc.RS(c).ShowMessage("用户被冻结")
        }
        return cc.RS(c).SetMessage("登录成功").ShowOkay(cc.ShowUserMap(u))
    }

    u.OpenId = openId.(string)
    u.Bean = 2000
    result = db.Orm.Create(&u)
    if result.Error != nil {
        return cc.RS(c).SetCode(500).ShowError("注册失败", result.Error)
    }
    return cc.RS(c).SetMessage("注册成功").ShowOkay(cc.ShowUserMap(u))
}

type PostDebugLoginArgs struct {
    Uid int `json:"uid"`
}

func (cc *ActionUser) PostDebugLogin(c echo.Context) error {
    var u entity.User
    args := PostDebugLoginArgs{}
    _ = c.Bind(&args)

    db.Orm.Model(u).Where("id=?", args.Uid).First(&u)
    return cc.RS(c).ShowOkay(cc.ShowUserMap(u))
}

func (cc *ActionUser) GetRefreshUserInfo(c echo.Context) error {
    usr := cc.ParseUserInfo(c)
    var u entity.User
    db.Orm.Model(u).Where("id = ?", usr.UID).First(&u)
    return cc.RS(c).ShowOkay(cc.ShowUserMap(u))
}

type PostUserInfoArgs struct {
    HeadImg  string `json:"head_img"`
    Nickname string `json:"nickname"`
    Gender   int8   `json:"gender"`
}

func (cc *ActionUser) PostUserInfo(c echo.Context) error {
    args := PostUserInfoArgs{}
    _ = c.Bind(&args)

    usr := cc.ParseUserInfo(c)
    var u entity.User
    result := db.Orm.Model(u).Where("id = ?", usr.UID).Updates(map[string]interface{}{
        "head_img": args.HeadImg,
        "nickname": args.Nickname,
        "gender":   args.Gender,
    })

    db.Orm.Model(u).Where("id=?", usr.UID).First(&u)
    if result.Error != nil {
        return cc.RS(c).SetCode(500).ShowError("更新用户信息失败", result.Error)
    }
    return cc.RS(c).SetMessage("更新用户信息成功").ShowOkay(cc.ShowUserMap(u))
}
