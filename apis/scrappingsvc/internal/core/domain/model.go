package domain

type Listing struct {
	Source   string
	Title    string
	Link     string
	PriceMXN float64
	Location string
}

func (l Listing) IsValid() bool {
	return l.PriceMXN > 0 && l.Title != ""
}
