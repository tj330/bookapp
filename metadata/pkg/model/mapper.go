package model

import "github.com/tj330/bookapp/gen"

// MetadataToProto is the mapper used to convert the
// metadadata to equivalent protobuf definition.
func MetadataToProto(m *Metadata) *gen.Metadata {
	return &gen.Metadata{
		Id:          m.ID,
		Title:       m.Title,
		Description: m.Description,
		Author:      m.Author,
		Isbn:        m.ISBN,
	}
}

// MetadataFromProto is the mapper used to get the
// corresponding metadadata from a protobuf definition.
func MetadataFromProto(m *gen.Metadata) *Metadata {
	return &Metadata{
		ID:          m.Id,
		Title:       m.Title,
		Description: m.Description,
		Author:      m.Author,
		ISBN:        m.Isbn,
	}
}
