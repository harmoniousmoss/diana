package handlers

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"myfiberproject/database"
	"myfiberproject/libs"
	"myfiberproject/models"

	"github.com/gofiber/fiber/v2"
)

// StoreMediaMonitoring fetches and stores emails from a specific sender into MongoDB.
func StoreMediaMonitoring(c *fiber.Ctx) error {
	accessToken, err := libs.GetMicrosoftAccessToken()
	if err != nil {
		log.Println("Error retrieving Microsoft Graph access token:", err)
		return c.Status(fiber.StatusInternalServerError).
			JSON(fiber.Map{"error": "Failed to authenticate with Microsoft Graph API"})
	}
	log.Println("Successfully retrieved Microsoft Graph access token.")

	// Get the monitored sender email from environment variables.
	monitoredSender := "your_sender@example.com" // Replace with dynamic retrieval if needed
	userEmail := "your_user@example.com"         // Replace with dynamic retrieval if needed

	// Construct and encode the filter query for the specific sender.
	filterStr := "from/emailAddress/address eq '" + monitoredSender + "'"
	encodedFilter := url.QueryEscape(filterStr)

	// Initial Graph API request URL for fetching emails.
	nextLink := "https://graph.microsoft.com/v1.0/users/" + userEmail + "/messages?$filter=" + encodedFilter + "&$top=50"
	log.Println("Graph API request initiated.")

	client := &http.Client{Timeout: 10 * time.Second}
	var allEmails []libs.Email

	// Loop while there's a next page of emails to fetch.
	for nextLink != "" {
		req, err := http.NewRequest("GET", nextLink, nil)
		if err != nil {
			log.Println("Error creating request:", err)
			return c.Status(fiber.StatusInternalServerError).
				JSON(fiber.Map{"error": "Failed to create request"})
		}

		req.Header.Set("Authorization", "Bearer "+accessToken)
		req.Header.Set("Accept", "application/json")
		req.Header.Set("ConsistencyLevel", "eventual") // Required for $filter queries

		resp, err := client.Do(req)
		if err != nil {
			log.Println("Error sending request:", err)
			return c.Status(fiber.StatusInternalServerError).
				JSON(fiber.Map{"error": "Failed to fetch emails"})
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			log.Println("Error reading response body:", err)
			return c.Status(fiber.StatusInternalServerError).
				JSON(fiber.Map{"error": "Failed to read response body"})
		}

		log.Println("Graph API response received with status:", resp.Status)
		if resp.StatusCode != http.StatusOK {
			return c.Status(fiber.StatusInternalServerError).
				JSON(fiber.Map{"error": "Failed to retrieve emails", "details": string(bodyBytes)})
		}

		var emailResp libs.EmailResponse
		if err := json.Unmarshal(bodyBytes, &emailResp); err != nil {
			log.Println("Error decoding JSON response:", err)
			return c.Status(fiber.StatusInternalServerError).
				JSON(fiber.Map{"error": "Failed to decode email response"})
		}

		allEmails = append(allEmails, emailResp.Value...)
		log.Printf("Fetched %d emails from the current page", len(emailResp.Value))
		nextLink = emailResp.NextLink
	}

	log.Printf("Total emails retrieved: %d", len(allEmails))

	// Prepare documents for MongoDB insertion.
	var documents []interface{}
	for _, email := range allEmails {
		receivedTime, err := time.Parse(time.RFC3339, email.ReceivedDateTime)
		if err != nil {
			log.Printf("Skipping email with invalid ReceivedDateTime (%s): %v", email.ReceivedDateTime, err)
			continue
		}

		doc := models.GraphEmail{
			ID:               email.ID,
			ReceivedDateTime: receivedTime,
			Subject:          email.Subject,
			From: models.EmailSender{
				Name:    email.From.EmailAddress.Name,
				Address: email.From.EmailAddress.Address,
			},
			Body: models.EmailBody{
				ContentType: email.Body.ContentType,
				Content:     email.Body.Content,
			},
			FetchedAt: time.Now(),
		}
		documents = append(documents, doc)
	}

	if len(documents) == 0 {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "No new emails to insert", "count": 0})
	}

	// Insert documents into MongoDB.
	db := database.GetMongoClient().Database("your_database") // Replace with actual DB name
	coll := db.Collection("your_collection")                  // Replace with actual collection name

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	res, err := coll.InsertMany(ctx, documents)
	if err != nil {
		log.Println("Error inserting documents into MongoDB:", err)
		return c.Status(fiber.StatusInternalServerError).
			JSON(fiber.Map{"error": "Failed to insert emails into database"})
	}

	log.Printf("Successfully inserted %d documents into MongoDB.", len(res.InsertedIDs))
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Emails inserted successfully",
		"count":   len(res.InsertedIDs),
	})
}
