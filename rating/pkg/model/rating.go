package model

type RecordID string

type RecordType string

const (
	RecordTypeBook = RecordType("book")
)

type UserID string

type RatingValue int

// The raw rating definition.
type Rating struct {
	RecordID   RecordID    `json:"recordId"`
	RecordType RecordType  `json:"recordType"`
	UserID     UserID      `json:"userId"`
	Value      RatingValue `json:"value"`
}

type RatingEventType string

const (
	RatingEventTypePut    = RatingEventType("string")
	RatingEventTypeDelete = RatingEventType("delete")
)

// RatingEvent consists of rating, the provider
// it belong to and the event type.
type RatingEvent struct {
	Rating
	ProviderID string          `json:"providerId"`
	EventType  RatingEventType `json:"eventType"`
}
