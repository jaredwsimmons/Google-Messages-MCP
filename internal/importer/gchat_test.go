package importer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/maxghenis/openmessage/internal/db"
)

func testStore(t *testing.T) *db.Store {
	t.Helper()
	dir := t.TempDir()
	store, err := db.New(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("create store: %v", err)
	}
	t.Cleanup(func() { store.Close() })
	return store
}

func TestGChatImport(t *testing.T) {
	store := testStore(t)

	jsonData := `{
		"messages": [
			{
				"creator": {"name": "Alice", "email": "alice@example.com", "user_type": "Human"},
				"created_date": "Monday, February 10, 2025 at 3:45:22\u202fPM UTC",
				"text": "Hey, how are you?",
				"topic_id": "abc123",
				"message_id": "abc123/msg1"
			},
			{
				"creator": {"name": "Bob", "email": "bob@example.com", "user_type": "Human"},
				"created_date": "Monday, February 10, 2025 at 3:46:00\u202fPM UTC",
				"text": "I'm good, thanks!",
				"topic_id": "abc123",
				"message_id": "abc123/msg2"
			}
		]
	}`

	importer := &GChat{MyEmail: "bob@example.com"}
	result, err := importer.Import(store, strings.NewReader(jsonData))
	if err != nil {
		t.Fatalf("import: %v", err)
	}

	if result.ConversationsCreated != 1 {
		t.Errorf("conversations created = %d, want 1", result.ConversationsCreated)
	}
	if result.MessagesImported != 2 {
		t.Errorf("messages imported = %d, want 2", result.MessagesImported)
	}

	// Verify conversation
	conv, err := store.GetConversation("gchat:abc123")
	if err != nil {
		t.Fatalf("get conversation: %v", err)
	}
	if conv.SourcePlatform != "gchat" {
		t.Errorf("source_platform = %q, want gchat", conv.SourcePlatform)
	}
	if conv.Name != "Alice" {
		t.Errorf("name = %q, want Alice", conv.Name)
	}

	// Verify messages
	msgs, err := store.GetMessagesByConversation("gchat:abc123", 10)
	if err != nil {
		t.Fatalf("get messages: %v", err)
	}
	if len(msgs) != 2 {
		t.Fatalf("got %d messages, want 2", len(msgs))
	}
	// Messages ordered DESC by timestamp
	if msgs[0].SenderName != "Bob" {
		t.Errorf("latest sender = %q, want Bob", msgs[0].SenderName)
	}
	if !msgs[0].IsFromMe {
		t.Error("Bob's message should be is_from_me=true")
	}
	if msgs[1].IsFromMe {
		t.Error("Alice's message should be is_from_me=false")
	}
	if msgs[0].SourcePlatform != "gchat" {
		t.Errorf("source_platform = %q, want gchat", msgs[0].SourcePlatform)
	}
}

func TestGChatImportDedup(t *testing.T) {
	store := testStore(t)

	jsonData := `{
		"messages": [
			{
				"creator": {"name": "Alice", "email": "alice@example.com"},
				"created_date": "Monday, February 10, 2025 at 3:45:22 PM UTC",
				"text": "Hello",
				"topic_id": "abc123",
				"message_id": "abc123/msg1"
			}
		]
	}`

	importer := &GChat{}

	// First import
	result1, _ := importer.Import(store, strings.NewReader(jsonData))
	if result1.MessagesImported != 1 {
		t.Errorf("first import: messages = %d, want 1", result1.MessagesImported)
	}

	// Second import (same data)
	result2, _ := importer.Import(store, strings.NewReader(jsonData))
	// UpsertMessage uses ON CONFLICT DO UPDATE, so it succeeds but effectively dedupes
	if result2.MessagesImported != 1 {
		t.Errorf("second import: messages = %d, want 1 (upsert)", result2.MessagesImported)
	}

	// Total messages in DB should be 1
	count, _ := store.MessageCount("gchat")
	if count != 1 {
		t.Errorf("total messages = %d, want 1", count)
	}
}

func TestGChatImportDirectory(t *testing.T) {
	store := testStore(t)

	// Create temp directory structure simulating Takeout
	dir := t.TempDir()
	groupDir := filepath.Join(dir, "DM abc123")
	os.MkdirAll(groupDir, 0755)
	os.WriteFile(filepath.Join(groupDir, "messages.json"), []byte(`{
		"messages": [
			{
				"creator": {"name": "Alice", "email": "alice@example.com"},
				"created_date": "Monday, February 10, 2025 at 3:45:22 PM UTC",
				"text": "Hi",
				"topic_id": "abc123",
				"message_id": "abc123/msg1"
			}
		]
	}`), 0644)
	os.WriteFile(filepath.Join(groupDir, "group_info.json"), []byte(`{
		"name": "Test Chat",
		"members": [{"name": "Alice", "email": "alice@example.com"}]
	}`), 0644)

	result, err := ImportGChatDirectory(store, dir, "bob@example.com")
	if err != nil {
		t.Fatalf("import dir: %v", err)
	}
	if result.ConversationsCreated != 1 {
		t.Errorf("conversations = %d, want 1", result.ConversationsCreated)
	}
	if result.MessagesImported != 1 {
		t.Errorf("messages = %d, want 1", result.MessagesImported)
	}

	// Verify conversation name came from group_info.json
	conv, _ := store.GetConversation("gchat:abc123")
	if conv.Name != "Test Chat" {
		t.Errorf("name = %q, want 'Test Chat'", conv.Name)
	}
}

func TestParseGChatDate(t *testing.T) {
	tests := []struct {
		input string
		want  int64
	}{
		{
			"Monday, February 10, 2025 at 3:45:22\u202fPM UTC",
			1739202322000,
		},
		{
			"Tuesday, January 1, 2013 at 12:00:00 AM UTC",
			1356998400000,
		},
	}

	for _, tt := range tests {
		got := parseGChatDate(tt.input)
		if got != tt.want {
			t.Errorf("parseGChatDate(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}
