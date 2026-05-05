package model

type RecordID string

type RecordType string

const (
	RecordTypeBook = RecordType("book")
)

type UserID string

type RatingValue int

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

type RatingEvent struct {
	Rating
	ProviderID string          `json:"providerId"`
	EventType  RatingEventType `json:"eventType"`
}
