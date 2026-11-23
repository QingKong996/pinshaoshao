package model

type FeedItem struct {
	ID            int
	ShopID        int
	ShopName      string
	Name          string
	Description   string
	CategoryName  *string
	MinPriceCents int64
	MaxPriceCents int64
}

type FeedResult struct {
	Items    []FeedItem
	Page     int
	PageSize int
	Total    int64
}

type ProductSKU struct {
	ID             int
	Specs          map[string]string
	PriceCents     int64
	IsEnabled      bool
	AvailableStock int64
}

func (sku ProductSKU) InStock() bool {
	return sku.AvailableStock > 0
}

type ProductDetail struct {
	ID           int
	ShopID       int
	ShopName     string
	Name         string
	Description  string
	CategoryName *string
	SKUs         []ProductSKU
}
