package importer

import (
	"bufio"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/maxghenis/openmessage/internal/db"
)

// WhatsApp line format: [MM/DD/YYYY, HH:MM:SS] Sender: Text
// Also handles: [DD/MM/YYYY, HH:MM:SS] and [M/D/YY, H:MM:SS AM/PM] variants.
var whatsappLineRe = regexp.MustCompile(`^\[(\d{1,2}/\d{1,2}/\d{2,4},\s*\d{1,2}:\d{2}(?::\d{2})?(?:\s*[AP]M)?)\]\s+(.+?):\s(.+)$`)

// WhatsApp imports messages from a WhatsApp text export file.
type WhatsApp struct {
	// MyName identifies the current user for is_from_me detection.
	MyName string
	// ConversationName overrides the auto-detected conversation name.
	ConversationName string
}

func (w *WhatsApp) Import(store *db.Store, source io.Reader) (*ImportResult, error) {
	scanner := bufio.NewScanner(source)
	// WhatsApp exports can have very long lines (media descriptions, etc.)
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024)

	result := &ImportResult{}

	var messages []waMessage
	var currentMsg *waMessage

	for scanner.Scan() {
		line := scanner.Text()
		matches := whatsappLineRe.FindStringSubmatch(line)
		if matches != nil {
			// Save previous multi-line message
			if currentMsg != nil {
				messages = append(messages, *currentMsg)
			}
			ts := parseWhatsAppDate(matches[1])
			currentMsg = &waMessage{
				timestamp: ts,
				sender:    matches[2],
				text:      matches[3],
			}
		} else if currentMsg != nil {
			// Continuation of previous message (multi-line)
			currentMsg.text += "\n" + line
		}
	}
	if currentMsg != nil {
		messages = append(messages, *currentMsg)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan: %w", err)
	}

	if len(messages) == 0 {
		return result, nil
	}

	// Build participants and conversation
	participants := w.buildParticipants(messages)
	isGroup := len(participants) > 2

	convName := w.ConversationName
	if convName == "" {
		convName = w.deriveConversationName(participants)
	}

	// Generate a stable conversation ID from participant names
	convID := "whatsapp:" + hashString(convName)

	participantsJSON, _ := json.Marshal(participants)
	var maxTS int64
	for _, m := range messages {
		ts := m.timestamp.UnixMilli()
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
		SourcePlatform: "whatsapp",
	}); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("create conversation: %v", err))
		return result, nil
	}
	result.ConversationsCreated = 1

	for _, m := range messages {
		// Skip system messages (media omitted, encryption notice, etc.)
		if strings.Contains(m.text, "<Media omitted>") {
			continue
		}

		ts := m.timestamp.UnixMilli()
		sourceID := hashString(fmt.Sprintf("%s:%d:%s", m.sender, ts, m.text))
		msgID := "whatsapp:" + sourceID

		isFromMe := w.MyName != "" && strings.EqualFold(m.sender, w.MyName)

		msg := &db.Message{
			MessageID:      msgID,
			ConversationID: convID,
			SenderName:     m.sender,
			Body:           m.text,
			TimestampMS:    ts,
			Status:         "delivered",
			IsFromMe:       isFromMe,
			SourcePlatform: "whatsapp",
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

func (w *WhatsApp) buildParticipants(messages []waMessage) []map[string]string {
	seen := map[string]bool{}
	var participants []map[string]string
	for _, m := range messages {
		if seen[m.sender] {
			continue
		}
		seen[m.sender] = true
		participants = append(participants, map[string]string{
			"name":   m.sender,
			"number": "",
		})
	}
	return participants
}

func (w *WhatsApp) deriveConversationName(participants []map[string]string) string {
	var names []string
	for _, p := range participants {
		name := p["name"]
		if w.MyName != "" && strings.EqualFold(name, w.MyName) {
			continue
		}
		names = append(names, name)
	}
	if len(names) == 0 {
		return "WhatsApp Chat"
	}
	return strings.Join(names, ", ")
}

// parseWhatsAppDate handles common WhatsApp date formats.
func parseWhatsAppDate(s string) time.Time {
	s = strings.TrimSpace(s)
	for _, layout := range []string{
		"1/2/06, 3:04:05 PM",
		"1/2/2006, 3:04:05 PM",
		"1/2/06, 3:04 PM",
		"1/2/2006, 3:04 PM",
		"01/02/2006, 15:04:05",
		"01/02/2006, 15:04",
		"1/2/06, 15:04:05",
		"1/2/06, 15:04",
		"02/01/2006, 15:04:05", // DD/MM/YYYY variant
	} {
		if t, err := time.Parse(layout, s); err == nil {
			return t
		}
	}
	return time.Time{}
}

func hashString(s string) string {
	h := sha256.Sum256([]byte(s))
	return fmt.Sprintf("%x", h[:8])
}

type waMessage struct {
	timestamp time.Time
	sender    string
	text      string
}
