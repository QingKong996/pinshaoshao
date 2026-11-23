package dto

type ListFeedQuery struct {
	Page     *int `form:"page" binding:"omitempty,min=1,max=1000000" example:"1"`
	PageSize *int `form:"page_size" binding:"omitempty,min=1,max=50" example:"20"`
}

type FeedItemResponse struct {
	ID            int     `json:"id" example:"1"`
	ShopID        int     `json:"shop_id" example:"2"`
	ShopName      string  `json:"shop_name" example:"山野杂货铺"`
	Name          string  `json:"name" example:"手冲咖啡豆"`
	Description   string  `json:"description" example:"中度烘焙，坚果与焦糖风味"`
	CategoryName  *string `json:"category_name" example:"食品"`
	MinPriceCents int64   `json:"min_price_cents" example:"3999"`
	MaxPriceCents int64   `json:"max_price_cents" example:"4999"`
}

type FeedListResponse struct {
	Items      []FeedItemResponse `json:"items"`
	Page       int                `json:"page" example:"1"`
	PageSize   int                `json:"page_size" example:"20"`
	Total      int64              `json:"total" example:"42"`
	TotalPages int                `json:"total_pages" example:"3"`
}

type ProductDetailSKUResponse struct {
	ID         int               `json:"id" example:"5"`
	Specs      map[string]string `json:"specs"`
	PriceCents int64             `json:"price_cents" example:"1999"`
	IsEnabled  bool              `json:"is_enabled" example:"true"`
	InStock    bool              `json:"in_stock" example:"true"`
}

type ProductDetailResponse struct {
	ID           int                        `json:"id" example:"10"`
	ShopID       int                        `json:"shop_id" example:"2"`
	ShopName     string                     `json:"shop_name" example:"山野杂货铺"`
	Name         string                     `json:"name" example:"手冲咖啡豆"`
	Description  string                     `json:"description" example:"中度烘焙，坚果与焦糖风味"`
	CategoryName *string                    `json:"category_name" example:"食品"`
	SKUs         []ProductDetailSKUResponse `json:"skus"`
}
