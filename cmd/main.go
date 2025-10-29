package main

import (
	"context"
	"fmt"
	"guimox/internal/github"
	"guimox/internal/whatsapp"
	"log"
	"os"
	"regexp"
	"strconv"
)

func main() {
	fmt.Println("Starting WhatsApp Newsletter Counter and GitHub Updater...")

	// Get detailed newsletter subscriber data from WhatsApp
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

	// Load GitHub configuration and create updater
	fmt.Println("Loading GitHub configuration...")
	config, err := github.LoadConfig()
	if err != nil {
		log.Printf("Error loading GitHub config: %v", err)
		os.Exit(1)
	}

	updater := github.NewGitHubUpdater(config)

	// Update GitHub repository with subscriber data
	fmt.Println("Updating GitHub repository...")
	ctx := context.Background()

	if err := updater.UpdateRepoDescription(ctx, newsletterData.Total); err != nil {
		log.Printf("Error updating repository description: %v", err)
		os.Exit(1)
	}

	if err := updater.UpdateDetailedDAU(ctx, &github.NewsletterData{
		Total:       newsletterData.Total,
		Newsletters: convertNewsletterInfo(newsletterData.Newsletters),
	}); err != nil {
		log.Printf("Error updating README DAU: %v", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Successfully updated GitHub repository with %d subscribers across %d newsletters!\n",
		newsletterData.Total, len(newsletterData.Newsletters))
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

func extractSubscriberCount(info string) (int, error) {
	// Extract number from "Total subscribers: X" format
	re := regexp.MustCompile(`Total subscribers: (\d+)`)
	matches := re.FindStringSubmatch(info)

	if len(matches) < 2 {
		return 0, fmt.Errorf("could not extract subscriber count from: %s", info)
	}

	count, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, fmt.Errorf("failed to parse subscriber count: %v", err)
	}

	return count, nil
}
