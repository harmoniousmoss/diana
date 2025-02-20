package models

import "time"

// GraphEmail represents an email fetched from Microsoft Graph that you plan to store.
type GraphEmail struct {
	ID               string      `bson:"_id"` // Using the Graph email ID as the MongoDB _id.
	ReceivedDateTime time.Time   `bson:"receivedDateTime"`
	Subject          string      `bson:"subject"`
	From             EmailSender `bson:"from"`
	Body             EmailBody   `bson:"body"`
	FetchedAt        time.Time   `bson:"fetchedAt"`
}

// EmailSender holds sender information.
type EmailSender struct {
	Name    string `bson:"name"`
	Address string `bson:"address"`
}

// EmailBody holds the email content details.
type EmailBody struct {
	ContentType string `bson:"contentType"`
	Content     string `bson:"content"`
}
