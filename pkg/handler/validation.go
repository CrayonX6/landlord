package handler

import (
    "github.com/go-playground/locales/en"
    "github.com/go-playground/locales/zh"
    "github.com/go-playground/universal-translator"
    "github.com/go-playground/validator/v10"
    zhI18n "github.com/go-playground/validator/v10/translations/zh"
    "sync"
)

var uni *ut.UniversalTranslator

type CustomValidator struct {
    lock      sync.Mutex
    validator *validator.Validate
}

func NewCustomValidator() *CustomValidator {
    return &CustomValidator{validator: validator.New()}
}

func (cv *CustomValidator) Validate(i interface{}) error {
    cv.lock.Lock()
    defer cv.lock.Unlock()
    
    zhTrans := zh.New()
    enTrans := en.New()
    uni = ut.New(zhTrans, zhTrans, enTrans)
    trans, _ := uni.GetTranslator("zh")
    _ = zhI18n.RegisterDefaultTranslations(cv.validator, trans)
    err := cv.validator.Struct(i)
    return err
}
