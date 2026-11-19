package model

// DashboardStats holds the shop-wide counters shown on the console dashboard.
type DashboardStats struct {
	ProductCount        int64
	OnShelfProductCount int64
	SKUCount            int64
	CategoryCount       int64
	TotalStock          int64
	InventoryValueCents int64
	OutOfStockSKUCount  int64
}
