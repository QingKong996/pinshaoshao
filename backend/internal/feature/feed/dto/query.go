package feeddto

type ListFeedQuery struct {
	Page     *int `form:"page" binding:"omitempty,min=1,max=1000000" example:"1"`
	PageSize *int `form:"page_size" binding:"omitempty,min=1,max=50" example:"20"`
}
