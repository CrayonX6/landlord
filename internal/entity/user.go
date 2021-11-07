package entity

type User struct {
    OpenId    string `json:"open_id"`
    HeadImg   string `json:"head_img"`
    Nickname  string `json:"nickname"`
    Gender    int8   `json:"gender"`
    Bean      int    `json:"bean"`
    PlayTimes int    `json:"play_times"`
    WinTimes  int    `json:"win_times"`
    MySQLTable
}

func (t *User) TableName() string {
    return "user"
}
