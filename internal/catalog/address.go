package catalog

import "strings"

// Address is the postal address of a seller.
type Address struct {
	City       string `json:"city"`
	PostalCode string `json:"postal_code"`
	FirstLine  string `json:"first_line"`
	SecondLine string `json:"second_line,omitempty"`
}

// SameCity reports whether the address is in city. The match ignores case
// and surrounding spaces.
func (a Address) SameCity(city string) bool {
	return strings.EqualFold(strings.TrimSpace(a.City), strings.TrimSpace(city))
}
