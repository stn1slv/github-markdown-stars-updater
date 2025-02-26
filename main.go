package main

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/google/go-github/v68/github"
	"golang.org/x/oauth2"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: markdown-github-stars-updater <path_to_markdown_file>")
		return
	}
	// Read markdown content
	filePath := os.Args[1]

	markdownContent, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println("Error reading the markdown file:", err)
		return
	}

	// Parse markdown content
	updatedMarkdown, err := updateStarCounts(string(markdownContent))
	if err != nil {
		fmt.Println("Error updating star counts:", err)
		return
	}

	// Write the updated content back to the file
	err = os.WriteFile(filePath, []byte(updatedMarkdown), 0644)
	if err != nil {
		fmt.Println("Error writing updated markdown to file:", err)
		return
	}

	fmt.Println("Markdown file updated successfully.")
}

/*
updateStarCounts finds GitHub repository links in the given markdownContent, fetches the current star counts,
updates the star count information in markdownContent, and returns the updated content.
*/
func updateStarCounts(markdownContent string) (string, error) {
	// Regular expression to find GitHub repository links
	re := regexp.MustCompile(`\[([^\]]+)\]\((https:\/\/github\.com\/[^\/)]+\/[^\/)]+)\)`)
	matches := re.FindAllStringSubmatch(markdownContent, -1)

	for _, match := range matches {
		itemName := match[1]
		repoURL := match[2]
		starCount, err := getStarsCount(repoURL)
		if err != nil {
			return "", err
		}

		// Update star count in the link title
		updatedLink := fmt.Sprintf("[%s (⭐%s)](%s)", removeStarsInfo(itemName), formatStarCount(starCount), repoURL)
		markdownContent = strings.Replace(markdownContent, match[0], updatedLink, 1)
	}

	return markdownContent, nil
}

// getStarsCount takes a GitHub repository URL and returns the current number of stars
func getStarsCount(repoURL string) (int, error) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: getAccessToken()},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	owner, repo := parseRepoName(repoURL[len("https://github.com/"):])
	repository, _, err := client.Repositories.Get(ctx, owner, repo)
	if err != nil {
		return 0, err
	}

	return repository.GetStargazersCount(), nil
}

// parseRepoName takes a GitHub repository name (formatted as "owner/repo") and returns the owner and repo parts.
func parseRepoName(repoName string) (string, string) {
	parts := strings.Split(repoName, "/")
	if len(parts) != 2 {
		panic("Invalid repository name format")
	}
	return parts[0], parts[1]
}

// removeStarsInfo removes the existing star count information from the input string.
func removeStarsInfo(input string) string {
	// Create a regular expression to find the "(⭐...)" pattern
	re := regexp.MustCompile(`\(⭐.*\)`)
	// Replace the matched substrings with an empty string
	result := re.ReplaceAllString(input, "")
	return strings.TrimSpace(result)
}

// formatStarCount formats the given star count for display in the markdown content.
func formatStarCount(stars int) string {
	if stars < 1000 {
		return fmt.Sprintf("%d", stars)
	} else if stars < 10000 {
		wholePart := stars / 1000
		decimalPart := (stars % 1000) / 100
		if decimalPart == 0 {
			return fmt.Sprintf("%dk", wholePart)
		}
		return fmt.Sprintf("%d.%dk", wholePart, decimalPart)
	} else {
		return fmt.Sprintf("%dk", stars/1000)
	}
}

// getAccessToken retrieves the GitHub access token from the environment variable
func getAccessToken() string {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		fmt.Println("Error: GITHUB_TOKEN environment variable not set")
		os.Exit(1)
	}
	return token
}
