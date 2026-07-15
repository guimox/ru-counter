package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
		lambda.Start(handler)

	} else {
		if err := handler(); err != nil {
			log.Fatal(err)
		}
	}
}

func handler() error {
	log.Printf("Starting WhatsApp Newsletter Counter and GitHub Updater...")

	updateTime := time.Now()
	utcTime := updateTime.UTC()
	log.Printf("Update started at: %s\n", utcTime.Format("02/01/2006 15:04:05 MST"))

	log.Printf("Fetching WhatsApp newsletter data...")
	newsletterData, err := GetDetailedNewsletterData()
	if err != nil {
		return fmt.Errorf("Error getting newsletter data: %v", err)
	}

	log.Printf("✅ WhatsApp data retrieved successfully!")
	log.Printf("Total subscribers: %d\n", newsletterData.Total)
	for _, newsletter := range newsletterData.Newsletters {
		log.Printf("- %s: %d subscribers\n", newsletter.Name, newsletter.Subscribers)
	}

	log.Printf("Loading GitHub configuration...")
	config, err := LoadConfig()
	if err != nil {
		log.Printf("Error loading GitHub config: %v", err)
		os.Exit(1)
	}

	updater := NewGitHubUpdater(config)

	log.Printf("Updating GitHub repository...")
	ctx := context.Background()

	if err := updater.UpdateRepoDescription(ctx, newsletterData.Total); err != nil {
		return fmt.Errorf("Error updating repository description: %v", err)
	}

	if err := updater.UpdateDetailedDAU(ctx, &NewsletterData{
		Total:       newsletterData.Total,
		Newsletters: convertNewsletterInfo(newsletterData.Newsletters),
		UpdatedAt:   updateTime,
	}); err != nil {
		log.Printf("Error updating README DAU: %v", err)
		os.Exit(1)
	}

	log.Printf("Successfully updated GitHub repository with %d subscribers across %d newsletters!\n",
		newsletterData.Total, len(newsletterData.Newsletters))
	log.Printf("Last updated at %s\n", utcTime.Format("02/01/2006 15:04:05 MST"))

	return nil
}

func convertNewsletterInfo(whatsappNewsletters []NewsletterInfo) []NewsletterInfo {
	var githubNewsletters []NewsletterInfo
	for _, wn := range whatsappNewsletters {
		githubNewsletters = append(githubNewsletters, NewsletterInfo{
			Name:        wn.Name,
			JID:         wn.JID,
			Subscribers: wn.Subscribers,
		})
	}
	return githubNewsletters
}
