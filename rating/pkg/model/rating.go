package model

type RecordID string
type RecordType string

const (
	RecordTypeMovie       = RecordType("movie")
	RatingEventTypePut    = "put"
	RatingEventTypeDelete = "delete"
)

type UserID string
type RatingValue int
type RatingEventType string

type Rating struct {
	RecordID   RecordID    `json:"recordId"`
	RecordType RecordType  `json:"recordType"`
	UserID     UserID      `json:"userId"`
	Value      RatingValue `json:"value"`
}

type RatingEvent struct {
	UserID     UserID          `json:"userId"`
	RecordID   RecordID        `json:"recordId"`
	RecordType RecordType      `json:"recordType"`
	Value      RatingValue     `json:"value"`
	EventType  RatingEventType `json:"eventType"`
}
