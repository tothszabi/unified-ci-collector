package main

import (
	"flag"
	"fmt"
	"log"
	"os/exec"
	"sort"
	"strings"
)

var statePath = ""

func init() {
	var path string
	flag.StringVar(&path, "state", "", "Path to the TF state")
	flag.Parse()

	statePath = path
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	if statePath == "" {
		return fmt.Errorf("missing state path")
	}

	cmd := exec.Command("terraform", "state", "list", "-state", statePath)
	bytes, err := cmd.Output()

	if err != nil {
		return err
	}

	unifiedCISet := make(map[string]struct{})
	legacySet := make(map[string]struct{})

	lines := strings.Split(string(bytes), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		components := strings.Split(line, ".")
		if len(components) < 4 {
			return fmt.Errorf("incorrect format: %s has less than 4 components", line)
		}

		repoName := components[3]

		if strings.Contains(line, "legacy_webhook") {
			legacySet[repoName] = struct{}{}
			continue
		}

		if strings.Contains(line, "unified_ci_webhook") {
			unifiedCISet[repoName] = struct{}{}
		}
	}

	unifiedCIRepos := extractKeys(unifiedCISet)
	sort.Strings(unifiedCIRepos)

	legacyRepos := extractKeys(legacySet)
	sort.Strings(legacyRepos)

	fmt.Printf("Unified CI (%d):\n", len(unifiedCIRepos))
	fmt.Printf("%s\n\n", strings.Join(unifiedCIRepos, "\n"))

	fmt.Printf("Legacy CI (%d):\n", len(legacyRepos))
	fmt.Printf("%s\n", strings.Join(legacyRepos, "\n"))

	return nil
}

func extractKeys(set map[string]struct{}) []string {
	var keys []string
	for key := range set {
		keys = append(keys, key)
	}
	return keys
}
