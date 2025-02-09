package config

import "os"

// BLS API Key (Loaded from .env)
var BLS_API_KEY = os.Getenv("BLS_API_KEY")

// BLS Series IDs
var BLS_SERIES_INFO = map[string]struct {
	Name string
	Unit string
}{
	"APU0000708111": {"Eggs, Grade A, Large", "per dozen"},
	"APU0000702111": {"Bread, White, Pan", "per lb."},
	"APU0000709213": {"Milk, Fresh, Low Fat", "per gallon"},
	"APU0000FF1101": {"Chicken Breast, Boneless", "per lb."},
	"APU0000704111": {"Bacon, Sliced", "per lb."},
	"APU0000711111": {"Apples, Red Delicious", "per lb."},
	"APU0000711311": {"Oranges, Navel", "per lb."},
	"APU00007471A": {"Gasoline, All Types", "per gal."},
	"LEU0252881600": {"Median Usual Weekly Earnings", "constant 1982-1984 dollars"},
}
