package github

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"

	"github.com/google/go-github/v57/github"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
)

type Config struct {
	GithubToken string
	Owner       string
	Repo        string
}

type GitHubUpdater struct {
	client *github.Client
	config Config
}

func NewGitHubUpdater(config Config) *GitHubUpdater {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: config.GithubToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	return &GitHubUpdater{
		client: client,
		config: config,
	}
}

func (g *GitHubUpdater) UpdateRepoDescription(ctx context.Context, subscriberCount int) error {
	description := fmt.Sprintf("With %d DAU, this repo stores 2 microservices created for a solution for university students (UFPR) to receive the daily college restaurant menu on multiple WhatsApp Groups. Using AWS, the project includes a scraper for menu data extraction and a WhatsApp sender for distribution", subscriberCount)

	repo := &github.Repository{
		Description: github.String(description),
	}

	_, _, err := g.client.Repositories.Edit(ctx, g.config.Owner, g.config.Repo, repo)
	if err != nil {
		return fmt.Errorf("failed to update repository description: %w", err)
	}

	log.Printf("Successfully updated repository description with %d subscribers", subscriberCount)
	return nil
}

func (g *GitHubUpdater) UpdateReadmeDAU(ctx context.Context, subscriberCount int) error {
	fileContent, _, _, err := g.client.Repositories.GetContents(
		ctx,
		g.config.Owner,
		g.config.Repo,
		"README.md",
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to get README: %w", err)
	}

	content, err := base64.StdEncoding.DecodeString(*fileContent.Content)
	if err != nil {
		return fmt.Errorf("failed to decode README content: %w", err)
	}

	re := regexp.MustCompile(`(\d{1,3}(?:,\d{3})*|\d+)\s+daily active users`)

	formattedCount := formatNumber(subscriberCount)
	updatedContent := re.ReplaceAllString(string(content), fmt.Sprintf("%s daily active users", formattedCount))

	opts := &github.RepositoryContentFileOptions{
		Message: github.String(fmt.Sprintf("ru-counter: Update DAU to %s users (from WhatsApp newsletter subscribers)", formattedCount)),
		Content: []byte(updatedContent),
		SHA:     fileContent.SHA,
		Branch:  github.String("main"),
	}

	_, _, err = g.client.Repositories.UpdateFile(
		ctx,
		g.config.Owner,
		g.config.Repo,
		"README.md",
		opts,
	)
	if err != nil {
		return fmt.Errorf("failed to update README: %w", err)
	}

	log.Printf("Successfully updated README with %d subscribers as DAU", subscriberCount)
	return nil
}

// NewsletterInfo represents a WhatsApp newsletter with subscriber count
type NewsletterInfo struct {
	Name        string
	JID         string
	Subscribers int
}

// NewsletterData represents the complete newsletter data
type NewsletterData struct {
	Total       int
	Newsletters []NewsletterInfo
}

func (g *GitHubUpdater) UpdateDetailedDAU(ctx context.Context, data *NewsletterData) error {
	fileContent, _, _, err := g.client.Repositories.GetContents(
		ctx,
		g.config.Owner,
		g.config.Repo,
		"README.md",
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to get README: %w", err)
	}

	content, err := base64.StdEncoding.DecodeString(*fileContent.Content)
	if err != nil {
		return fmt.Errorf("failed to decode README content: %w", err)
	}

	// Update the total DAU count
	totalRe := regexp.MustCompile(`\*\*(\d{1,3}(?:,\d{3})*|\d+)\s+daily active users\*\*`)
	formattedTotal := formatNumber(data.Total)
	updatedContent := totalRe.ReplaceAllString(string(content), fmt.Sprintf("**%s daily active users**", formattedTotal))

	// Update individual newsletter subscriber counts
	for _, newsletter := range data.Newsletters {
		// Match pattern like "- RU [Name](link) = X users"
		pattern := fmt.Sprintf(`(- RU \[%s\]\([^)]+\)) = \d+ users`, regexp.QuoteMeta(newsletter.Name))
		re := regexp.MustCompile(pattern)
		replacement := fmt.Sprintf("${1} = %d users", newsletter.Subscribers)
		updatedContent = re.ReplaceAllString(updatedContent, replacement)
	}

	opts := &github.RepositoryContentFileOptions{
		Message: github.String(fmt.Sprintf("ru-counter: Update DAU to %s users with detailed breakdown (from WhatsApp newsletter subscribers)", formattedTotal)),
		Content: []byte(updatedContent),
		SHA:     fileContent.SHA,
		Branch:  github.String("main"),
	}

	_, _, err = g.client.Repositories.UpdateFile(
		ctx,
		g.config.Owner,
		g.config.Repo,
		"README.md",
		opts,
	)
	if err != nil {
		return fmt.Errorf("failed to update README: %w", err)
	}

	log.Printf("Successfully updated README with detailed DAU: %d total users across %d newsletters", data.Total, len(data.Newsletters))
	return nil
}

func formatNumber(n int) string {
	// Return plain numbers without any formatting
	return strconv.Itoa(n)
}

func (g *GitHubUpdater) GetCurrentDAU(ctx context.Context) (int, error) {
	fileContent, _, _, err := g.client.Repositories.GetContents(
		ctx,
		g.config.Owner,
		g.config.Repo,
		"README.md",
		nil,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to get README: %w", err)
	}

	content, err := base64.StdEncoding.DecodeString(*fileContent.Content)
	if err != nil {
		return 0, fmt.Errorf("failed to decode README content: %w", err)
	}

	re := regexp.MustCompile(`(\d{1,3}(?:,\d{3})*|\d+)\s+daily active users`)
	matches := re.FindStringSubmatch(string(content))

	if len(matches) < 2 {
		return 0, fmt.Errorf("could not find DAU in README")
	}

	dauStr := regexp.MustCompile(`,`).ReplaceAllString(matches[1], "")
	dau, err := strconv.Atoi(dauStr)
	if err != nil {
		return 0, fmt.Errorf("failed to parse DAU: %w", err)
	}

	return dau, nil
}

func (g *GitHubUpdater) PerformWeeklyUpdate() {
	ctx := context.Background()

	log.Println("Starting weekly update...")

	currentDAU, err := g.GetCurrentDAU(ctx)
	if err != nil {
		log.Printf("Error getting current DAU: %v", err)
		return
	}

	log.Printf("Current DAU: %d", currentDAU)

	newDAU := currentDAU + 100

	if err := g.UpdateRepoDescription(ctx, newDAU); err != nil {
		log.Printf("Error updating description: %v", err)
		return
	}

	if err := g.UpdateReadmeDAU(ctx, newDAU); err != nil {
		log.Printf("Error updating README: %v", err)
		return
	}

	log.Printf("Weekly update completed successfully! Updated DAU from %d to %d", currentDAU, newDAU)
}

func LoadConfig() (Config, error) {
	_ = godotenv.Load()

	githubPatToken := os.Getenv("GITHUB_PAT_TOKEN")
	if githubPatToken == "" {
		return Config{}, fmt.Errorf("GITHUB_PAT_TOKEN environment variable not set")
	}

	githubRepoOwner := os.Getenv("GITHUB_OWNER")
	if githubRepoOwner == "" {
		return Config{}, fmt.Errorf("GITHUB_OWNER environment variable not set")
	}

	githubRepoName := os.Getenv("GITHUB_REPO")
	if githubRepoName == "" {
		return Config{}, fmt.Errorf("GITHUB_REPO environment variable not set")
	}

	return Config{
		GithubToken: githubPatToken,
		Owner:       githubRepoOwner,
		Repo:        githubRepoName,
	}, nil
}
