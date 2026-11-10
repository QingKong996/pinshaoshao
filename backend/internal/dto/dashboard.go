package dto

type DashboardResponse struct {
	ProductCount        int64 `json:"product_count" example:"12"`
	OnShelfProductCount int64 `json:"on_shelf_product_count" example:"8"`
	SKUCount            int64 `json:"sku_count" example:"34"`
	CategoryCount       int64 `json:"category_count" example:"4"`
	TotalStock          int64 `json:"total_stock" example:"500"`
	InventoryValueCents int64 `json:"inventory_value_cents" example:"123456"`
	OutOfStockSKUCount  int64 `json:"out_of_stock_sku_count" example:"1"`
}
