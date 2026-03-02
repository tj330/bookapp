package model

import "github.com/tj330/bookapp/metadata/pkg/model"

type BookDetails struct {
	Rating   *float64       `json:"rating,omitEmpty"`
	Metadata model.Metadata `json:"metadata"`
}
