package acme

import (
    tool "beginning/pkg/acme"
    "errors"
    "fmt"
    "github.com/dgrijalva/jwt-go"
    "github.com/golang-module/carbon"
    "github.com/labstack/echo/v4"
    "github.com/spf13/viper"
    "strings"
)

type UserInfo struct {
    Acme
}

type UserInfoApi struct {
    UID      int64  `json:"id"`
    UserMark string `json:"user_mark"`
    OpenID   string `json:"open_id"`
    Nickname string `json:"nickname"`
    Avatar   string `json:"avatar"`
    jwt.StandardClaims
}

// 生成用户 Token
func (u *UserInfo) TokenCreator(user map[string]interface{}) string {
    if _, ok := user["id"]; !ok {
        return ""
    }
    if tool.ToInt(user["id"]) == 0 {
        return ""
    }

    claims := &UserInfoApi{}
    tool.AlignStructAndMap(user, &claims)
    hour := viper.GetInt("cnf.token_expire_hours")
    claims.StandardClaims.ExpiresAt = carbon.Now().AddHours(hour).ToTimestampWithSecond()

    // Create token with claims
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    t, err := token.SignedString([]byte(viper.GetString("jwt.secret")))
    if err != nil {
        panic(err)
    }

    return fmt.Sprintf("Bearer %s", t)
}

// 解析用户 Token
func (u *UserInfo) ParseToken(toke string) (uia UserInfoApi, err error) {
    toke = strings.Replace(toke, "Bearer ", "", 1)
    result, err := jwt.Parse(toke, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return viper.Get("jwt.secret"), nil
    })
    if result == nil {
        return
    }

    claims := result.Claims.(jwt.MapClaims)
    surplus := claims["exp"].(float64) - float64(carbon.Now().ToTimestampWithSecond())
    if surplus <= 0 {
        err = errors.New("jwt token expired")
        return
    }

    tool.AlignStructAndMap(claims, &uia)
    err = nil
    return
}

// 获取用户结构体
func (u *UserInfo) ParseUserInfo(c echo.Context) *UserInfoApi {
    token := c.Get("user")
    if token == nil {
        return &UserInfoApi{}
    }
    user := c.Get("user").(*jwt.Token)
    claims := user.Claims.(*UserInfoApi)

    return claims
}
