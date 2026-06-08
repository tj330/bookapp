package model

// The raw metadata definition.
type Metadata struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Author      string `json:"author"`
	ISBN        string `json:"isbn"`
}
