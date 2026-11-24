package catalog

import (
	"errors"
	"testing"
)

func TestSameCity(t *testing.T) {
	a := Address{City: "Gotham"}
	for _, city := range []string{"Gotham", "gotham", "  GOTHAM "} {
		if !a.SameCity(city) {
			t.Errorf("SameCity(%q) = false, want true", city)
		}
	}
	if a.SameCity("Metropolis") {
		t.Error("SameCity(Metropolis) = true, want false")
	}
}

func TestPromote(t *testing.T) {
	s := Seller{ID: "s1", Name: "Gotham Goods", IsVerified: true}
	d, err := Promote(s)
	if err != nil {
		t.Fatalf("Promote: %v", err)
	}
	if d.ID != "s1" || d.MaxListings != defaultMaxListings {
		t.Errorf("Promote = %+v", d)
	}

	s.IsVerified = false
	if _, err := Promote(s); !errors.Is(err, ErrUnverifiedSeller) {
		t.Errorf("Promote(unverified) error = %v, want ErrUnverifiedSeller", err)
	}
}

func TestSellersIn(t *testing.T) {
	sellers := []Seller{
		{ID: "a", Address: Address{City: "Gotham"}},
		{ID: "b", Address: Address{City: "Metropolis"}},
		{ID: "c", Address: Address{City: "gotham"}},
	}
	got := SellersIn(sellers, "Gotham")
	if len(got) != 2 || got[0].ID != "a" || got[1].ID != "c" {
		t.Errorf("SellersIn = %+v, want a and c", got)
	}
	if got := SellersIn(sellers, "Star City"); len(got) != 0 {
		t.Errorf("SellersIn(Star City) = %+v, want none", got)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name string
		p    Product
		want error
	}{
		{"ok", Product{Name: "Mug", PriceCents: 1499, Stock: 1}, nil},
		{"free", Product{Name: "Sticker"}, nil},
		{"blank name", Product{Name: "  ", PriceCents: 1}, ErrEmptyName},
		{"negative price", Product{Name: "Mug", PriceCents: -1}, ErrNegativePrice},
		{"negative stock", Product{Name: "Mug", Stock: -1}, ErrNegativeStock},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.p.Validate(); !errors.Is(err, tt.want) {
				t.Errorf("Validate() = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestApply(t *testing.T) {
	p := Product{ID: "p1", Name: "Mug", PriceCents: 1499, Description: "Blue", Stock: 3}
	price := int64(999)
	stock := 0
	got := p.Apply(ProductUpdate{PriceCents: &price, Stock: &stock})

	want := Product{ID: "p1", Name: "Mug", PriceCents: 999, Description: "Blue", Stock: 0}
	if got != want {
		t.Errorf("Apply = %+v, want %+v", got, want)
	}
	if p.PriceCents != 1499 {
		t.Error("Apply changed the receiver")
	}
}

func TestInStock(t *testing.T) {
	p := Product{Stock: 2}
	if !p.InStock(2) {
		t.Error("InStock(2) = false, want true")
	}
	if p.InStock(3) {
		t.Error("InStock(3) = true, want false")
	}
	if p.InStock(0) {
		t.Error("InStock(0) = true, want false")
	}
}
