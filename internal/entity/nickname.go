package entity

type Nickname struct {
	ID       int64  `json:"id"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

func (t *Nickname) TableName() string {
	return "nickname"
}
