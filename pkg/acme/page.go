package acme

import (
    "github.com/labstack/echo/v4"
    "gorm.io/gorm"
    "strconv"
)

type Paginate struct {
    Page  int
    Limit int
    count int64
}

// 进行分页
func NewPaginate() *Paginate {
    return &Paginate{}
}

// 设置页码
func (p *Paginate) SetPage(page int) *Paginate {
    if page <= 0 {
        page = 1
    }
    p.Page = page
    return p
}

// 设置每页条数
func (p *Paginate) SetLimit(limit int) *Paginate {
    if limit <= 0 {
        limit = 1
    } else if limit > 100 {
        limit = 100
    }
    p.Limit = limit
    return p
}

// 暗示总条数
func (p *Paginate) HintCount(count int64) *Paginate {
    p.count = count
    return p
}

// 进行分页
func (p *Paginate) Paginate(c echo.Context, query *gorm.DB, repo interface{}) error {
    query = query.WithContext(c.Request().Context())
    if p.Page == 0 {
        p.Page, _ = strconv.Atoi(c.QueryParam("page"))
    }
    if p.Limit == 0 {
        p.Limit, _ = strconv.Atoi(c.QueryParam("page_size"))
    }

    if p.Page == 0 {
        p.Page = 1
    }
    if p.Limit == 0 {
        p.Limit = 20
    }

    offset := (p.Page - 1) * p.Limit
    result := query.Offset(offset).Limit(p.Limit).Find(repo)

    count := p.Count(query)
    c.Response().Header().Set("X-Total-Count", strconv.Itoa(int(count)))
    return result.Error
}

// 统计总行数
func (p *Paginate) Count(query *gorm.DB) int64 {
    if p.count > 0 {
        return p.count
    }
    var count int64
    query.Count(&count)
    return count
}
