package accrual

// Order is the order status response payload from the accrual system.
type Order struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual"`
}
