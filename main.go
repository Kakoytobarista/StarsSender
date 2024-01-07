package main

import (
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	baseURL     = "https://api.github.com"
	searchQuery = "stars:>100"
	message     = "Hello! I'm an experienced back-end developer with expertise in crafting robust and scalable " +
		"server-side applications. My skills include designing RESTful APIs, managing databases, " +
		"and implementing business logic. I have a proven track record in Automated Quality Assurance, " +
		"ensuring high standards of product quality. Let's discuss how I can contribute to your team."
)

type Repository struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	Description string `json:"description"`
	HTMLURL     string `json:"html_url"`
	Owner       struct {
		Login string `json:"login"`
	} `json:"owner"`
}

func main() {
	// Load environment variables from the .env file
	if err := godotenv.Load(); err != nil {
		fmt.Println("Error loading .env file:", err)
		return
	}

	// Get GitHub token from environment variable
	token := os.Getenv("GITHUB_TOKEN")

	// Check if the token is empty
	if token == "" {
		fmt.Println("GitHub token is missing. Please set GITHUB_TOKEN environment variable.")
		return
	}

	rand.Seed(time.Now().UnixNano())

	// Search for repositories
	repositories, err := searchRepositories(searchQuery, 100)
	if err != nil {
		fmt.Println("Error searching repositories:", err)
		return
	}

	for _, repo := range repositories {
		err := starRepository(token, repo.FullName)
		if err == nil {
			fmt.Printf("Starred repository: %s\n", repo.FullName)
		} else {
			fmt.Printf("Failed to star repository %s. Error: %v\n", repo.FullName, err)
		}
	}
}

func searchRepositories(query string, count int) ([]Repository, error) {
	languageFilter := "language:python"
	apiURL := fmt.Sprintf("%s/search/repositories?q=%s+%s&per_page=%d&sort=stars&order=asc", baseURL, query, languageFilter, count)

	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP request failed with status code: %d", resp.StatusCode)
	}

	var result struct {
		Items []Repository `json:"items"`
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	return result.Items, nil
}

func starRepository(token, fullName string) error {
	apiURL := fmt.Sprintf("%s/user/starred/%s", baseURL, fullName)

	client := &http.Client{}

	req, err := http.NewRequest("PUT", apiURL, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return nil
	} else if resp.StatusCode == http.StatusUnauthorized {
		remaining := resp.Header.Get("X-RateLimit-Remaining")
		reset := resp.Header.Get("X-RateLimit-Reset")

		fmt.Printf("Rate limit exceeded. Remaining requests: %s. Reset time: %s\n", remaining, reset)

		// Implement a simple backoff strategy
		resetTime, err := strconv.ParseInt(reset, 10, 64)
		if err != nil {
			return err
		}

		waitTime := time.Until(time.Unix(resetTime, 0))
		fmt.Printf("Waiting for %s before making the next request...\n", waitTime)
		time.Sleep(waitTime)

		// Retry the request
		return starRepository(token, fullName)
	} else {
		return fmt.Errorf("failed to star repository. Status Code: %d", resp.StatusCode)
	}
}
