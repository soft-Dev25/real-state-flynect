package domain

type Listing struct {
	ID       string  `json:"id,omitempty"`
	Source   string  `json:"source"`
	Title    string  `json:"title"`
	PriceMXN float64 `json:"price_mxn"`
	Location string  `json:"location"`
	Link     string  `json:"link"`
}
