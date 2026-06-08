package model

import "github.com/tj330/bookapp/metadata/pkg/model"

// BookDetails consisting of the book rating and the
// corresponding book metadata.
type BookDetails struct {
	// Empty rating is allowed.
	Rating   *float64       `json:"rating,omitempty"`
	Metadata model.Metadata `json:"metadata"`
}
