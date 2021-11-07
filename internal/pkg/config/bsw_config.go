package config

import (
    "beginning/internal/entity"
    "beginning/internal/pkg/acme"
    "beginning/internal/server/variables"
    tool "beginning/pkg/acme"
    "beginning/pkg/cache"
    "beginning/pkg/db"
    "errors"
    "strconv"
)

var VBswConfig RepoBswConfig

type RepoBswConfig struct {
    acme.Acme
}

type Cnf struct {
    Type            string      `json:"type"`
    Value           interface{} `json:"value"`
    OriginValue     interface{} `json:"origin_value"`
    AllowClientPull bool        `json:"allow_client_pull"`
    Comment         string      `json:"comment"`
}

// 列表配置
func (bc *RepoBswConfig) ListConfig() map[string]Cnf {
    kvp := make(map[string]Cnf)
    key := entity.CK4AppConfig()
    hit, _ := cache.Cached.Fetch(key)
    if len(hit) != 0 {
        tool.JsonToInterface(hit, &kvp)
    } else {
        var bce []entity.BswConfigEntity
        var item Cnf
        db.Orm.Where("state = ?", 1).Find(&bce)
        for _, record := range bce {
            switch record.Type {
            case 2: // 整数
                item.Value = tool.Str2Int64(record.Value)
                item.Type = "int64"
                break
            case 3: // 浮点数
                item.Value = tool.Str2Float64(record.Value)
                item.Type = "float64"
                break
            case 4: // 布尔值
                item.Value = tool.Str2Bool(record.Value)
                item.Type = "boolean"
                break
            default: // 字符串
                item.Value = record.Value
                item.Type = "string"
                break
            }
            item.OriginValue = record.Value
            item.Comment = record.Remark
            item.AllowClientPull = record.AllowClientPull > 0
            kvp[record.Key] = item
        }
        json := tool.InterfaceToJson(kvp)
        _ = cache.Cached.Save(key, json, 0) // 写入缓存
    }
    return kvp
}

// 获取项目配置
func (bc *RepoBswConfig) GetConfig(field string) (Cnf, error) {
    kvp := bc.ListConfig()
    if val, ok := kvp[field]; ok {
        return val, nil
    }
    return Cnf{}, errors.New("配置不存在")
}

// 获取项目配置(字符串)
func (bc *RepoBswConfig) GetString(field string, def string) string {
    cnf, err := bc.GetConfig(field)
    if err != nil {
        return def
    }
    return cnf.OriginValue.(string)
}

// 获取项目配置(整数)
func (bc *RepoBswConfig) GetInt(field string, def int) int {
    cnf, err := bc.GetConfig(field)
    if err != nil {
        return def
    }
    val, _ := strconv.Atoi(cnf.OriginValue.(string))
    return val
}

// 获取项目配置(整数)
func (bc *RepoBswConfig) GetInt64(field string, def int64) int64 {
    cnf, err := bc.GetConfig(field)
    if err != nil {
        return def
    }
    val, _ := strconv.ParseInt(cnf.OriginValue.(string), 10, 64)
    return val
}

// 获取项目配置(浮点数)
func (bc *RepoBswConfig) GetFloat64(field string, def float64) float64 {
    cnf, err := bc.GetConfig(field)
    if err != nil {
        return def
    }
    val, _ := strconv.ParseFloat(cnf.OriginValue.(string), 64)
    return val
}

// 获取项目配置(布尔值)
func (bc *RepoBswConfig) GetBool(field string, def bool) bool {
    cnf, err := bc.GetConfig(field)
    if err != nil {
        return def
    }
    return tool.ToBool(cnf.OriginValue)
}

// 修改配置
func (bc *RepoBswConfig) ModifyConfig(key string, value interface{}) error {
    result := db.Orm.Model(&entity.BswConfigEntity{}).Where("`key` = ?", key).Update("value", value)
    if result.Error != nil {
        key := entity.CK4AppConfig()
        variables.Rds.Del(key)
    }
    return result.Error
}
