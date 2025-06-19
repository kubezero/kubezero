package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"gopkg.in/yaml.v3"
)

type Stack struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Components  []string `yaml:"components"`
}

type StackCatalog struct {
	Stacks []Stack `yaml:"stacks"`
}

type CreatedIssue struct {
	Title  string
	URL    string
	Number int
}

func main() {
	data, err := ioutil.ReadFile("stacks/stack-catalog.yaml")
	if err != nil {
		fmt.Println("❌ Failed to read stacks/stack-catalog.yaml:", err)
		return
	}

	var catalog StackCatalog
	err = yaml.Unmarshal(data, &catalog)
	if err != nil {
		fmt.Println("❌ YAML unmarshal error:", err)
		return
	}

	repo := os.Getenv("REPO")
	token := os.Getenv("GITHUB_TOKEN")

	existing := getExistingIssueTitles(token, repo)
	var createdIssues []CreatedIssue

	for _, stack := range catalog.Stacks {
		title := fmt.Sprintf("stack(%s): %s", stack.Name, stack.Description)
		if issueNum, exists := existing[title]; exists {
			// Fetch the existing issue details
			issue := getIssue(token, repo, issueNum)
			if issue == nil {
				fmt.Println("⚠️ Could not fetch existing issue:", title)
				continue
			}
			// Generate the desired body and labels
			newBody := generateIssueBody(stack)
			newLabels := generateIssueLabels(stack)
			// Compare and update if needed
			if issue.Body != newBody || !compareStringSlices(issue.Labels, newLabels) {
				updateIssue(token, repo, issueNum, newBody, newLabels)
				fmt.Println("✏️ Updated issue:", title)
			} else {
				fmt.Println("🔁 Issue up-to-date:", title)
			}
			continue
		}
		created := createIssue(token, repo, stack)
		if created != nil {
			createdIssues = append(createdIssues, *created)
		}
	}

	if len(createdIssues) > 0 {
		updateEpicIssue(token, repo, createdIssues)
	}
}

// Helper to fetch a single issue by number
func getIssue(token, repo string, number int) *struct {
	Body   string
	Labels []string
} {
	url := fmt.Sprintf("https://api.github.com/repos/%s/issues/%d", repo, number)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("❌ Error creating request for issue #%d: %v\n", number, err)
		return nil
	}
	req.Header.Add("Authorization", "token "+token)
	req.Header.Add("Accept", "application/vnd.github+json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("❌ Error fetching issue #%d: %v\n", number, err)
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		fmt.Printf("❌ GitHub API returned status %d for issue #%d\n", resp.StatusCode, number)
		return nil
	}
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("❌ Error decoding issue #%d response: %v\n", number, err)
		return nil
	}
	body, ok := result["body"].(string)
	if !ok {
		fmt.Printf("❌ Issue #%d missing body\n", number)
		return nil
	}
	labels := []string{}
	if arr, ok := result["labels"].([]interface{}); ok {
		for _, l := range arr {
			if labelObj, ok := l.(map[string]interface{}); ok {
				if name, ok := labelObj["name"].(string); ok {
					labels = append(labels, name)
				}
			}
		}
	}
	return &struct {
		Body   string
		Labels []string
	}{Body: body, Labels: labels}
}

// Helper to update an issue's body and labels
func updateIssue(token, repo string, number int, body string, labels []string) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/issues/%d", repo, number)
	payload := map[string]interface{}{
		"body":   body,
		"labels": labels,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("❌ Error marshaling update payload for issue #%d: %v\n", number, err)
		return
	}
	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(data))
	if err != nil {
		fmt.Printf("❌ Error creating PATCH request for issue #%d: %v\n", number, err)
		return
	}
	req.Header.Add("Authorization", "token "+token)
	req.Header.Add("Accept", "application/vnd.github+json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("❌ Error updating issue #%d: %v\n", number, err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		fmt.Printf("❌ GitHub API returned status %d when updating issue #%d\n", resp.StatusCode, number)
	}
}

// Helper to generate issue body from stack
func generateIssueBody(stack Stack) string {
	body := fmt.Sprintf("### Stack: %s\n\n**Description:** %s\n\n**Components:**\n", stack.Name, stack.Description)
	for _, c := range stack.Components {
		body += fmt.Sprintf("- [ ] %s\n", c)
	}
	body += "\n---\nGenerated from `stacks/stack-catalog.yaml`"
	return body
}

// Helper to generate labels from stack
func generateIssueLabels(stack Stack) []string {
	labels := []string{"kind/stack", "status/planned"}
	for _, c := range stack.Components {
		labels = append(labels, "area/"+c)
	}
	return labels
}

// Helper to compare two string slices (order-insensitive)
func compareStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	m := make(map[string]int)
	for _, s := range a {
		m[s]++
	}
	for _, s := range b {
		if m[s] == 0 {
			return false
		}
		m[s]--
	}
	return true
}

func getExistingIssueTitles(token, repo string) map[string]int {
	url := fmt.Sprintf("https://api.github.com/repos/%s/issues?state=open&labels=kind/stack", repo)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", "token "+token)
	req.Header.Add("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error fetching issues:", err)
		return nil
	}
	defer resp.Body.Close()

	var issues []map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&issues)

	existing := make(map[string]int)
	for _, issue := range issues {
		title := issue["title"].(string)
		number := int(issue["number"].(float64))
		existing[title] = number
	}
	return existing
}

func createIssue(token, repo string, stack Stack) *CreatedIssue {
	title := fmt.Sprintf("stack(%s): %s", stack.Name, stack.Description)
	body := fmt.Sprintf("### Stack: %s\n\n**Description:** %s\n\n**Components:**\n", stack.Name, stack.Description)
	for _, c := range stack.Components {
		body += fmt.Sprintf("- [ ] %s\n", c)
	}
	body += "\n---\nGenerated from `stacks/stack-catalog.yaml`"

	labels := []string{"kind/stack", "status/planned"}
	for _, c := range stack.Components {
		labels = append(labels, "area/"+c)
	}

	payload := map[string]interface{}{
		"title":  title,
		"body":   body,
		"labels": labels,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("❌ Error marshaling payload for new issue: %v\n", err)
		return nil
	}
	url := fmt.Sprintf("https://api.github.com/repos/%s/issues", repo)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		fmt.Printf("❌ Error creating POST request for new issue: %v\n", err)
		return nil
	}
	req.Header.Add("Authorization", "token "+token)
	req.Header.Add("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("❌ Error creating issue: %s — %v\n", title, err)
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		fmt.Printf("❌ GitHub API returned status %d when creating issue: %s\n", resp.StatusCode, title)
		return nil
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("❌ Error decoding create issue response: %v\n", err)
		return nil
	}

	fmt.Printf("✔ Created issue: %s\n", title)

	return &CreatedIssue{
		Title:  title,
		URL:    result["html_url"].(string),
		Number: int(result["number"].(float64)),
	}
}

func updateEpicIssue(token, repo string, issues []CreatedIssue) {
	comment := "### 🔗 Synced Stack Sub-Tasks\n<!-- stack-sync -->\n"
	for _, i := range issues {
		comment += fmt.Sprintf("- [ ] [%s](%s)\n", i.Title, i.URL)
	}

	commentsURL := fmt.Sprintf("https://api.github.com/repos/%s/issues/30/comments", repo)
	req, err := http.NewRequest("GET", commentsURL, nil)
	if err != nil {
		fmt.Println("❌ Error creating GET request for epic comments:", err)
		return
	}
	req.Header.Add("Authorization", "token "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error checking existing comments:", err)
		return
	}
	defer resp.Body.Close()

	var comments []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&comments); err != nil {
		fmt.Println("❌ Error decoding comments response:", err)
		return
	}

	for _, c := range comments {
		body := c["body"].(string)
		if body != "" && containsStackSyncMarker(body) {
			commentID := int(c["id"].(float64))
			updateURL := fmt.Sprintf("https://api.github.com/repos/%s/issues/comments/%d", repo, commentID)
			postComment(updateURL, comment, token)
			fmt.Println("🔁 Updated stack-sync comment on issue #30")
			return
		}
	}

	postComment(commentsURL, comment, token)
	fmt.Println("➕ Posted new stack-sync comment on issue #30")
}

func containsStackSyncMarker(body string) bool {
	return bytes.Contains([]byte(body), []byte("<!-- stack-sync -->"))
}

func postComment(url, comment, token string) {
	data, err := json.Marshal(map[string]string{"body": comment})
	if err != nil {
		fmt.Println("❌ Error marshaling comment payload:", err)
		return
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		fmt.Println("❌ Error creating POST request for comment:", err)
		return
	}
	req.Header.Add("Authorization", "token "+token)
	req.Header.Add("Accept", "application/vnd.github+json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("❌ Error posting comment:", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		fmt.Printf("❌ GitHub API returned status %d when posting comment\n", resp.StatusCode)
	}
}
