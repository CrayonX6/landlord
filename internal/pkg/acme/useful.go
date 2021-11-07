package acme

import (
    tool "beginning/pkg/acme"
    "github.com/spf13/viper"
    "github.com/wumansgy/goEncrypt"
)

type Acme struct {
}

// AES 加密
func (a *Acme) AesEncode(plaintext interface{}) (cipherText string) {
    plaintextByte := []byte(tool.ToStr(plaintext))
    cipherTextByte, _ := goEncrypt.AesCbcEncrypt(plaintextByte, []byte(viper.GetString("aes.key")), []byte(viper.GetString("aes.iv"))...)
    cipherText, _ = tool.SafeBase64Encode(string(cipherTextByte))
    return
}

// AES 解密
func (a *Acme) AesDecode(cipherText string) (plaintext string) {
    cipherText, _ = tool.SafeBase64Decode(cipherText)
    plaintextByte, _ := goEncrypt.AesCbcDecrypt([]byte(cipherText), []byte(viper.GetString("aes.key")), []byte(viper.GetString("aes.iv"))...)
    plaintext = string(plaintextByte)
    return
}

// ID => Mark
func (a *Acme) IdToMark(id int64) string {
    return a.AesEncode(id)
}

// Mark => ID
func (a *Acme) MarkToId(mark string) int64 {
    if len(mark) == 0 {
        return 0
    }
    return tool.Str2Int64(a.AesDecode(mark))
}

// 调用微信接口
func (a *Acme) JsCode2Session(code string) (result map[string]interface{}, err error) {
    // 微信接口
    _, A := tool.GetCost(0)
    result, err = tool.NewCURL().SendGet("https://api.weixin.qq.com/sns/jscode2session", map[string]string{
        "appid":      viper.GetString("wx-applet.app_id"),
        "secret":     viper.GetString("wx-applet.app_secret"),
        "js_code":    code,
        "grant_type": "authorization_code",
    })
    _ = tool.Cost("WX_JSON2SESSION", "小程序登录调用接口 WX_JSON2SESSION 耗时", A)
    tool.PrintVar("小程序登录调用接口 WX_JSON2SESSION 返回结果", result)
    return
}

// code 是否有效
func (a *Acme) CodeIsValid(code string) bool {
    if len(code) == 0 {
        return false
    }
    result, err := a.JsCode2Session(code)
    if err != nil {
        return false
    }
    if _, ok := result["errcode"]; ok && result["errcode"] != 0 {
        return false
    }

    return true
}
