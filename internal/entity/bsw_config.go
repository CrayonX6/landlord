package entity

type BswConfigEntity struct {
    Key             string `json:"key"`               // 配置键
    Type            int8   `json:"type"`              // 值类型
    Value           string `json:"value"`             // 配置值
    AllowClientPull int8   `json:"allow_client_pull"` // 是否允许下发给客户端
    Remark          string `json:"remark"`            // 备注
    MySQLTable
}

func (t *BswConfigEntity) TableName() string {
    return "bsw_config"
}

// Cache key for 项目配置
func CK4AppConfig() string {
    return "cw:app_config"
}
