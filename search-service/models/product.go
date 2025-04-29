package models

type Product struct {
	ID 		string `json:"id"`
	Name 	string `json:"name"`
	Price 	float64 `json:"price"`
	Category string `json:"category"`
	Description string `json:"description"`
	ImageURL string `json:"image_url"`
}