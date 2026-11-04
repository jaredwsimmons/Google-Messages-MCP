// cmd/genviz/main.go — One-shot CLI to generate relationship vizzes.
// Usage: go run ./cmd/genviz -name "Mary MacLeod" -output /tmp/mary.html -person2 "Mary" -primary "#c4717a" -secondary "#d4a574"
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/maxghenis/openmessage/internal/db"
	"github.com/maxghenis/openmessage/internal/story"
	"github.com/maxghenis/openmessage/internal/viz"
)

func main() {
	name := flag.String("name", "", "Person name to search (required)")
	output := flag.String("output", "", "Output HTML path (required)")
	person1 := flag.String("person1", "Max", "First person display name")
	person2 := flag.String("person2", "", "Second person display name (default: matched name)")
	timezone := flag.String("tz", "America/New_York", "Timezone")
	password := flag.String("password", "", "Password hash for gate (empty = no gate)")
	primary := flag.String("primary", "", "Primary CSS color")
	secondary := flag.String("secondary", "", "Secondary CSS color")
	accent := flag.String("accent", "", "Accent CSS color")
	bg := flag.String("bg", "", "Background CSS color")
	style := flag.String("style", "intimate", "Story style")
	apiKey := flag.String("api-key", "", "Anthropic API key for narrative")
	photosDir := flag.String("photos", "", "Directory of photos to include in gallery")
	maxPhotos := flag.Int("max-photos", 30, "Maximum number of photos to include")
	flag.Parse()

	if *name == "" || *output == "" {
		flag.Usage()
		os.Exit(1)
	}

	// Open the database
	dataDir := os.Getenv("OPENMESSAGES_DATA_DIR")
	if dataDir == "" {
		home, _ := os.UserHomeDir()
		dataDir = home + "/.local/share/openmessage"
	}
	dbPath := dataDir + "/messages.db"

	store, err := db.New(dbPath)
	if err != nil {
		log.Fatalf("open db %s: %v", dbPath, err)
	}

	// Find matching conversations
	convs, err := store.ListConversations(1000)
	if err != nil {
		log.Fatalf("list conversations: %v", err)
	}

	nameLower := strings.ToLower(*name)
	var matchIDs []string
	var convNames []string
	for _, c := range convs {
		if c.IsGroup {
			continue
		}
		if strings.Contains(strings.ToLower(c.Name), nameLower) ||
			strings.Contains(strings.ToLower(c.Participants), nameLower) {
			matchIDs = append(matchIDs, c.ConversationID)
			platform := c.SourcePlatform
			if platform == "" {
				platform = "sms"
			}
			convNames = append(convNames, fmt.Sprintf("%s [%s]", c.Name, platform))
		}
	}

	if len(matchIDs) == 0 {
		log.Fatalf("no conversations found matching %q", *name)
	}
	fmt.Printf("Found %d conversations: %s\n", len(matchIDs), strings.Join(convNames, ", "))

	// Load messages
	msgs, err := store.GetMessagesByConversations(matchIDs, 500000)
	if err != nil {
		log.Fatalf("get messages: %v", err)
	}
	fmt.Printf("Loaded %d messages\n", len(msgs))

	// Deduplicate
	msgs = dedup(msgs)
	fmt.Printf("After dedup: %d messages\n", len(msgs))

	// Parse timezone
	tz, err := time.LoadLocation(*timezone)
	if err != nil {
		log.Fatalf("invalid timezone %q: %v", *timezone, err)
	}

	// Compute stats
	stats := story.ComputeStats(msgs, tz)
	fmt.Printf("Date range: %s to %s\n", stats.DateRange.Start, stats.DateRange.End)

	// Generate story
	st, err := story.Generate(msgs, story.GenerateConfig{
		Style:             *style,
		APIKey:            *apiKey,
		MaxSampleMessages: 200,
	})
	if err != nil {
		log.Fatalf("generate story: %v", err)
	}
	st.Stats = stats
	fmt.Printf("Generated %d chapters\n", len(st.Chapters))

	// Determine person2 name
	p2 := *person2
	if p2 == "" {
		for sender := range stats.SenderSplit {
			if sender != "me" {
				p2 = sender
				break
			}
		}
		if p2 == "" {
			p2 = *name
		}
	}

	// Encode photos if directory provided
	var photos []viz.Photo
	if *photosDir != "" {
		var err error
		photos, err = viz.EncodePhotosFromDir(*photosDir, *maxPhotos)
		if err != nil {
			log.Fatalf("encode photos: %v", err)
		}
		fmt.Printf("Encoded %d photos\n", len(photos))
	}

	// Build config
	config := viz.VizConfig{
		Person1:         *person1,
		Person2:         p2,
		PrimaryColor:    *primary,
		SecondaryColor:  *secondary,
		AccentColor:     *accent,
		BackgroundColor: *bg,
		Timezone:        *timezone,
		PasswordHash:    *password,
		Photos:          photos,
	}

	// Render
	html, err := viz.RenderHTML(stats, st, config)
	if err != nil {
		log.Fatalf("render: %v", err)
	}

	if err := os.WriteFile(*output, html, 0o644); err != nil {
		log.Fatalf("write: %v", err)
	}

	info, _ := os.Stat(*output)
	fmt.Printf("Written %s (%d KB)\n", *output, info.Size()/1024)
}

func dedup(msgs []*db.Message) []*db.Message {
	if len(msgs) <= 1 {
		return msgs
	}
	sort.Slice(msgs, func(i, j int) bool {
		return msgs[i].TimestampMS < msgs[j].TimestampMS
	})
	var result []*db.Message
	for _, m := range msgs {
		isDup := false
		start := len(result) - 20
		if start < 0 {
			start = 0
		}
		for i := len(result) - 1; i >= start; i-- {
			prev := result[i]
			tsDiff := m.TimestampMS - prev.TimestampMS
			if tsDiff > 2000 {
				break
			}
			if tsDiff >= 0 && tsDiff <= 2000 && m.Body == prev.Body && m.IsFromMe == prev.IsFromMe {
				isDup = true
				break
			}
		}
		if !isDup {
			result = append(result, m)
		}
	}
	return result
}
