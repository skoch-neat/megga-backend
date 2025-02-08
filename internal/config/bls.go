package config

import "os"

// BLS API Key (Loaded from .env)
var BLS_API_KEY = os.Getenv("BLS_API_KEY")

// BLS Series IDs
var BLS_SERIES_IDS = []string{
	"APU0000708111", // Eggs, grade A, large, per doz.
	"APU0000702111", // Bread, white, pan, per lb.
	"APU0000709213", // Milk, fresh, low fat, per gal.
	"APU0000FF1101", // Chicken breast, boneless, per lb.
	"APU0000704111", // Bacon, sliced, per lb.
	"APU0000711111", // Apples, Red Delicious, per lb.
	"APU0000711311", // Oranges, Navel, per lb.
	"APU00007471A",  // Gasoline, all types, per gal.
	"LEU0252881600", // Median usual weekly earnings (CPI-U Adjusted)
}
