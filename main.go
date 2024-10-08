package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/template"

	"golang.org/x/sync/errgroup"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

var (
	username     = envdefault("GITHUB_USERNAME", "jokestax")
	rssfeed      = envdefault("RSS_FEED", "")
	templateFile = envdefault("TEMPLATE_FILE", "template.md.gotmpl")
	maxPRs       = envintdefault("MAX_PULL_REQUESTS", 10)
	maxStarred   = envintdefault("MAX_STARRED_REPOS", 10)
	maxArticles  = envintdefault("MAX_ARTICLES", 5)
)

func run() error {
	data, err := os.ReadFile(templateFile)
	if err != nil {
		return fmt.Errorf("failed to read template file %q: %w", templateFile, err)
	}

	tmpl, err := template.New("template").Funcs(fncs).Parse(string(data))
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	eg := errgroup.Group{}

	var prs []PullRequest
	var starredRepos []StarredRepo
	var articles []Article

	eg.Go(func() error {
		var err error
		prs, err = getPullRequests(username, maxPRs)
		if err != nil {
			return fmt.Errorf("failed to get pull requests: %w", err)
		}
		return nil
	})

	eg.Go(func() error {
		var err error
		starredRepos, err = getStarredRepos(username, maxStarred)
		if err != nil {
			return fmt.Errorf("failed to get starred repos: %w", err)
		}
		return nil
	})

	// eg.Go(func() error {
	// 	var err error
	// 	articles, err = getArticles(rssfeed, maxArticles)
	// 	if err != nil {
	// 		return fmt.Errorf("failed to read feed %q %w", rssfeed, err)
	// 	}
	// 	return nil
	// })

	if err := eg.Wait(); err != nil {
		return err
	}

	if err := tmpl.Execute(os.Stdout, struct {
		PullRequests []PullRequest
		StarredRepos []StarredRepo
		Articles     []Article
	}{
		PullRequests: prs,
		StarredRepos: starredRepos,
		Articles:     articles,
	}); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}

func envdefault(key, def string) string {
	if val := strings.TrimSpace(os.Getenv(key)); val != "" {
		return val
	}

	return def
}

func envintdefault(key string, defval int) int {
	if val := strings.TrimSpace(os.Getenv(key)); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}

	return defval
}
