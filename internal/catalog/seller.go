package catalog

import "errors"

// ErrUnverifiedSeller is returned when an unverified seller is promoted.
var ErrUnverifiedSeller = errors.New("catalog: seller is not verified")

const defaultMaxListings = 500

// Seller lists products on the marketplace.
type Seller struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Address    Address `json:"address"`
	IsVerified bool    `json:"is_verified"`
}

// Dealer is a verified seller that can list products in bulk.
type Dealer struct {
	Seller
	MaxListings int `json:"max_listings"`
}

// DeliversTo reports whether the seller delivers to city. Sellers only
// deliver inside their own city for now.
func (s Seller) DeliversTo(city string) bool {
	return s.Address.SameCity(city)
}

// Promote turns a verified seller into a dealer.
func Promote(s Seller) (Dealer, error) {
	if !s.IsVerified {
		return Dealer{}, ErrUnverifiedSeller
	}
	return Dealer{Seller: s, MaxListings: defaultMaxListings}, nil
}

// SellersIn returns the sellers that deliver to city, in input order.
func SellersIn(sellers []Seller, city string) []Seller {
	var out []Seller
	for _, s := range sellers {
		if s.DeliversTo(city) {
			out = append(out, s)
		}
	}
	return out
}
