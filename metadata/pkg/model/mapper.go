package model

import "github.com/tj330/bookapp/gen"

func MetadataToProto(m *Metadata) *gen.Metadata {
	return &gen.Metadata{
		Id:          m.ID,
		Title:       m.Title,
		Description: m.Description,
		Author:      m.Author,
		Isbn:        m.ISBN,
	}
}

func MetadataFromProto(m *gen.Metadata) *Metadata {
	return &Metadata{
		ID:          m.Id,
		Title:       m.Title,
		Description: m.Description,
		Author:      m.Author,
		ISBN:        m.Isbn,
	}
}
