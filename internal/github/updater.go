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

func (g *GitHubUpdater) UpdateRepoDescription(ctx context.Context, newDAU int) error {
	description := fmt.Sprintf("With DAU of %d users, this repo stores 2 microservices created for a solution for university students (UFPR) to receive the daily college restaurant menu on multiple WhatsApp Groups. Using AWS, the project includes a scraper for menu data extraction and a WhatsApp sender for distribution", newDAU)

	repo := &github.Repository{
		Description: github.String(description),
	}

	_, _, err := g.client.Repositories.Edit(ctx, g.config.Owner, g.config.Repo, repo)
	if err != nil {
		return fmt.Errorf("failed to update repository description: %w", err)
	}

	log.Printf("Successfully updated repository description with DAU: %d", newDAU)
	return nil
}

func (g *GitHubUpdater) UpdateReadmeDAU(ctx context.Context, newDAU int) error {
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

	formattedDAU := formatNumber(newDAU)
	updatedContent := re.ReplaceAllString(string(content), fmt.Sprintf("%s daily active users", formattedDAU))

	opts := &github.RepositoryContentFileOptions{
		Message: github.String(fmt.Sprintf("Update DAU to %s users", formattedDAU)),
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

	log.Printf("Successfully updated README with DAU: %s", formattedDAU)
	return nil
}

func formatNumber(n int) string {
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

func loadConfig() (Config, error) {
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
