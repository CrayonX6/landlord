package acme

import (
    "beginning/pkg/db"
    "fmt"
    "gorm.io/gorm"
)

//
// 针对类型字段进行自动修改条件
//
// 规范
// 当类型值 < 0 时，该字段不加任何条件
// 当类型值 = 0 时，该字段等于 0
// 当类型值 > 0 时，该字段等于指定值或 0
//
func (a *Acme) AutoTypeCondition(query *gorm.DB, field string, value int) *gorm.DB {
    if value < 0 {
        return query
    }
    field = fmt.Sprintf("%s = ?", field)
    if value == 0 {
        return query.Where(field, value)
    }
    return query.Where(db.Orm.Where(field, 0).Or(field, value))
}

//
// 针对状态字段进行自动修改条件
//
// 规范
// 当类型值 < 0 时，该字段不加任何条件
// 当类型值 >= 0 时，该字段等于指定值
//
func (a *Acme) AutoStateCondition(query *gorm.DB, field string, value int) *gorm.DB {
    if value < 0 {
        return query
    }
    return query.Where(fmt.Sprintf("%s = ?", field), value)
}
