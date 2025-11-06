package main

import (
	"context"
	"fmt"
	"guimox/internal/github"
	"guimox/internal/whatsapp"
	"log"
	"os"
	"time"
)

func main() {
	fmt.Println("Starting WhatsApp Newsletter Counter and GitHub Updater...")

	updateTime := time.Now()
	utcTime := updateTime.UTC()
	fmt.Printf("Update started at: %s\n", utcTime.Format("02/01/2006 15:04:05 MST"))

	fmt.Println("Fetching WhatsApp newsletter data...")
	newsletterData, err := whatsapp.GetDetailedNewsletterData()
	if err != nil {
		log.Printf("Error getting newsletter data: %v", err)
		os.Exit(1)
	}

	fmt.Println("✅ WhatsApp data retrieved successfully!")
	fmt.Printf("Total subscribers: %d\n", newsletterData.Total)
	for _, newsletter := range newsletterData.Newsletters {
		fmt.Printf("- %s: %d subscribers\n", newsletter.Name, newsletter.Subscribers)
	}

	fmt.Println("Loading GitHub configuration...")
	config, err := github.LoadConfig()
	if err != nil {
		log.Printf("Error loading GitHub config: %v", err)
		os.Exit(1)
	}

	updater := github.NewGitHubUpdater(config)

	fmt.Println("Updating GitHub repository...")
	ctx := context.Background()

	if err := updater.UpdateRepoDescription(ctx, newsletterData.Total); err != nil {
		log.Printf("Error updating repository description: %v", err)
		os.Exit(1)
	}

	if err := updater.UpdateDetailedDAU(ctx, &github.NewsletterData{
		Total:       newsletterData.Total,
		Newsletters: convertNewsletterInfo(newsletterData.Newsletters),
		UpdatedAt:   updateTime,
	}); err != nil {
		log.Printf("Error updating README DAU: %v", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Successfully updated GitHub repository with %d subscribers across %d newsletters!\n",
		newsletterData.Total, len(newsletterData.Newsletters))
	fmt.Printf("Last updated at %s\n", utcTime.Format("02/01/2006 15:04:05 MST"))
}

func convertNewsletterInfo(whatsappNewsletters []whatsapp.NewsletterInfo) []github.NewsletterInfo {
	var githubNewsletters []github.NewsletterInfo
	for _, wn := range whatsappNewsletters {
		githubNewsletters = append(githubNewsletters, github.NewsletterInfo{
			Name:        wn.Name,
			JID:         wn.JID,
			Subscribers: wn.Subscribers,
		})
	}
	return githubNewsletters
}
