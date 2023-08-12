package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
)

const (
	TelegramBotToken = "" // Ask BotFather
	ChatID           = "" // Invite this bot to get your group ChatID, @chat_id_echo_bot / https://web.telegram.org/a/#1513323938
	GitLabURL        = "YOUR_GITLAB_INSTANCE_URL"
)

type GitLabEvent struct {
	ObjectKind   string `json:"object_kind"`
	UserUsername string `json:"user_username"`
	Ref          string `json:"ref"`
	Project      struct {
		WebURL        string `json:"web_url"`
		Name          string `json:"name"`
		DefaultBranch string `json:"default_branch"`
	} `json:"project"`
	ObjectAttributes struct {
		URL   string `json:"url"`
		Title string `json:"title"`
	} `json:"object_attributes"`
	User struct {
		Username string `json:"username"`
	} `json:"user"`
}

func main() {
	http.HandleFunc("/", handleWebhook)
	log.Println("Listening at 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var event GitLabEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Error decoding JSON: %s", err)
		return
	}

	message := generateMessage(event)

	sendTelegramMessage(message)

	w.WriteHeader(http.StatusOK)
}

func generateMessage(event GitLabEvent) string {
	switch event.ObjectKind {
	case "push":
		branchName := strings.TrimPrefix(event.Ref, "refs/heads/")
		branchLink := fmt.Sprintf("%s/-/tree/%s", event.Project.WebURL, url.QueryEscape(branchName))

		return fmt.Sprintf("New push by %s to %s:\n\nRef: %s\nBranch Page: %s",
			event.UserUsername, event.Project.Name, event.Ref, branchLink)
	case "merge_request":
		return fmt.Sprintf("New merge request by %s:\n\nTitle: %s\nLink: %s",
			event.User.Username, event.ObjectAttributes.Title, event.ObjectAttributes.URL)
	case "note":
		return fmt.Sprintf("New comment by %s:\n\nLink: %s",
			event.User.Username, event.ObjectAttributes.URL)
	default:
		return "Unknown event type"
	}
}

func sendTelegramMessage(message string) {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", TelegramBotToken)
	data := url.Values{
		"chat_id": {ChatID},
		"text":    {message},
	}

	resp, err := http.PostForm(apiURL, data)
	if err != nil {
		log.Printf("Error sending message to Telegram: %s", err)
		return
	}
	defer resp.Body.Close()
}
