package importer

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/maxghenis/openmessage/internal/db"
)

// gchatMessage matches the Google Chat Takeout JSON format.
type gchatMessage struct {
	Creator struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		UserType string `json:"user_type"`
	} `json:"creator"`
	CreatedDate string `json:"created_date"`
	Text        string `json:"text"`
	TopicID     string `json:"topic_id"`
	MessageID   string `json:"message_id"`
}

type gchatExport struct {
	Messages []gchatMessage `json:"messages"`
}

// gchatGroupInfo matches the group_info.json format.
type gchatGroupInfo struct {
	Name    string `json:"name"`
	Members []struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"members"`
}

// GChat imports Google Chat Takeout messages from a single messages.json file.
type GChat struct {
	// MyEmail identifies the current user for is_from_me detection.
	// If empty, no messages are marked as from me.
	MyEmail string
	// ConversationName overrides the conversation name (useful when group_info.json isn't available).
	ConversationName string
}

func (g *GChat) Import(store *db.Store, source io.Reader) (*ImportResult, error) {
	data, err := io.ReadAll(source)
	if err != nil {
		return nil, fmt.Errorf("read source: %w", err)
	}

	var export gchatExport
	if err := json.Unmarshal(data, &export); err != nil {
		return nil, fmt.Errorf("parse JSON: %w", err)
	}

	result := &ImportResult{}
	if len(export.Messages) == 0 {
		return result, nil
	}

	// Derive conversation ID from the first message's topic_id
	convID := "gchat:" + export.Messages[0].TopicID

	// Build participants list and detect group chat
	participants := g.buildParticipants(export.Messages)
	isGroup := len(participants) > 2

	convName := g.ConversationName
	if convName == "" {
		convName = g.deriveConversationName(participants)
	}

	// Create conversation
	participantsJSON, _ := json.Marshal(participants)
	var maxTS int64
	for _, m := range export.Messages {
		ts := parseGChatDate(m.CreatedDate)
		if ts > maxTS {
			maxTS = ts
		}
	}

	if err := store.UpsertConversation(&db.Conversation{
		ConversationID: convID,
		Name:           convName,
		IsGroup:        isGroup,
		Participants:   string(participantsJSON),
		LastMessageTS:  maxTS,
		SourcePlatform: "gchat",
	}); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("create conversation: %v", err))
		return result, nil
	}
	result.ConversationsCreated = 1

	// Import messages
	for _, m := range export.Messages {
		if m.Text == "" {
			continue
		}

		ts := parseGChatDate(m.CreatedDate)
		sourceID := m.MessageID
		msgID := "gchat:" + sourceID

		isFromMe := g.MyEmail != "" && strings.EqualFold(m.Creator.Email, g.MyEmail)

		msg := &db.Message{
			MessageID:      msgID,
			ConversationID: convID,
			SenderName:     m.Creator.Name,
			SenderNumber:   m.Creator.Email,
			Body:           m.Text,
			TimestampMS:    ts,
			Status:         "delivered",
			IsFromMe:       isFromMe,
			SourcePlatform: "gchat",
			SourceID:       sourceID,
		}

		if err := store.UpsertMessage(msg); err != nil {
			if strings.Contains(err.Error(), "UNIQUE constraint") {
				result.MessagesDuplicate++
			} else {
				result.Errors = append(result.Errors, fmt.Sprintf("message %s: %v", msgID, err))
			}
			continue
		}
		result.MessagesImported++
	}

	return result, nil
}

func (g *GChat) buildParticipants(msgs []gchatMessage) []map[string]string {
	seen := map[string]bool{}
	var participants []map[string]string
	for _, m := range msgs {
		key := m.Creator.Email
		if key == "" {
			key = m.Creator.Name
		}
		if seen[key] {
			continue
		}
		seen[key] = true
		participants = append(participants, map[string]string{
			"name":   m.Creator.Name,
			"number": m.Creator.Email,
		})
	}
	return participants
}

func (g *GChat) deriveConversationName(participants []map[string]string) string {
	var names []string
	for _, p := range participants {
		name := p["name"]
		if name == "" {
			name = p["number"]
		}
		if g.MyEmail != "" && strings.EqualFold(p["number"], g.MyEmail) {
			continue
		}
		names = append(names, name)
	}
	if len(names) == 0 {
		return "Google Chat"
	}
	return strings.Join(names, ", ")
}

// parseGChatDate handles the Google Chat Takeout date format:
// "Monday, February 18, 2025 at 3:45:22 PM UTC" (with possible \u202f narrow no-break spaces)
func parseGChatDate(s string) int64 {
	// Strip narrow no-break spaces
	s = strings.ReplaceAll(s, "\u202f", " ")
	// Remove day name prefix: "Monday, "
	if idx := strings.Index(s, ", "); idx >= 0 && idx < 15 {
		s = s[idx+2:]
	}
	// Remove " at "
	s = strings.Replace(s, " at ", " ", 1)
	// Try common formats
	for _, layout := range []string{
		"January 2, 2006 3:04:05 PM UTC",
		"January 2, 2006 3:04:05 PM",
		"January 2, 2006 3:04 PM UTC",
		"January 2, 2006 3:04 PM",
	} {
		if t, err := time.Parse(layout, s); err == nil {
			return t.UnixMilli()
		}
	}
	return 0
}

// ImportGChatDirectory imports all conversation directories from a Google Chat Takeout export.
// baseDir should be the "Google Chat/Groups/" directory containing subdirectories.
func ImportGChatDirectory(store *db.Store, baseDir, myEmail string) (*ImportResult, error) {
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return nil, fmt.Errorf("read directory: %w", err)
	}

	total := &ImportResult{}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		messagesPath := filepath.Join(baseDir, entry.Name(), "messages.json")
		if _, err := os.Stat(messagesPath); os.IsNotExist(err) {
			continue
		}

		f, err := os.Open(messagesPath)
		if err != nil {
			total.Errors = append(total.Errors, fmt.Sprintf("%s: %v", entry.Name(), err))
			continue
		}

		// Try to load group_info.json for conversation name
		convName := ""
		infoPath := filepath.Join(baseDir, entry.Name(), "group_info.json")
		if infoData, err := os.ReadFile(infoPath); err == nil {
			var info gchatGroupInfo
			if json.Unmarshal(infoData, &info) == nil && info.Name != "" {
				convName = info.Name
			}
		}

		importer := &GChat{MyEmail: myEmail, ConversationName: convName}
		result, err := importer.Import(store, f)
		f.Close()
		if err != nil {
			total.Errors = append(total.Errors, fmt.Sprintf("%s: %v", entry.Name(), err))
			continue
		}

		total.ConversationsCreated += result.ConversationsCreated
		total.MessagesImported += result.MessagesImported
		total.MessagesDuplicate += result.MessagesDuplicate
		total.Errors = append(total.Errors, result.Errors...)
	}

	return total, nil
}

// gchatHash creates a short deterministic ID from input.
func gchatHash(input string) string {
	h := sha256.Sum256([]byte(input))
	return fmt.Sprintf("%x", h[:8])
}
