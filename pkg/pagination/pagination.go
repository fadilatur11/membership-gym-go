package pagination

import (
	"math"
	"strconv"

	"github.com/gin-gonic/gin"
)

const DefaultLimit = 10
const MaxLimit = 100

type Params struct {
	Page  int
	Limit int
}

type Meta struct {
	Page      int   `json:"page"`
	Limit     int   `json:"limit"`
	Total     int64 `json:"total"`
	TotalPage int   `json:"total_page"`
}

func Parse(ctx *gin.Context) Params {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}
	return Params{Page: page, Limit: limit}
}

func (p Params) Offset() int {
	return (p.Page - 1) * p.Limit
}

func NewMeta(page, limit int, total int64) Meta {
	totalPage := 0
	if limit > 0 {
		totalPage = int(math.Ceil(float64(total) / float64(limit)))
	}
	return Meta{Page: page, Limit: limit, Total: total, TotalPage: totalPage}
}
