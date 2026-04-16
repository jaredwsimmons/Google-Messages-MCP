package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/maxghenis/openmessage/internal/app"
	"github.com/maxghenis/openmessage/internal/db"
)

var (
	sendWhatsAppText = func(a *app.App, conversationID, body, replyToID string) (*db.Message, error) {
		return a.SendWhatsAppText(conversationID, body, replyToID)
	}
	sendSignalText = func(a *app.App, conversationID, body, replyToID string) (*db.Message, error) {
		return a.SendSignalText(conversationID, body, replyToID)
	}
)

func sendToConversationTool() mcp.Tool {
	return mcp.NewTool("send_to_conversation",
		mcp.WithDescription("Send a text message to an existing conversation by conversation ID across supported platforms"),
		mcp.WithString("conversation_id", mcp.Required(), mcp.Description("Existing conversation ID from list_conversations or get_conversation")),
		mcp.WithString("message", mcp.Required(), mcp.Description("Message text to send")),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithIdempotentHintAnnotation(false),
	)
}

func sendToConversationHandler(a *app.App) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		conversationID := strArg(args, "conversation_id")
		message := strArg(args, "message")

		if conversationID == "" {
			return errorResult("conversation_id is required"), nil
		}
		if message == "" {
			return errorResult("message is required"), nil
		}

		conv, err := a.Store.GetConversation(conversationID)
		if err != nil {
			return errorResult(fmt.Sprintf("failed to load conversation: %v", err)), nil
		}
		if conv == nil {
			return errorResult(fmt.Sprintf("conversation %s not found", conversationID)), nil
		}

		switch conv.SourcePlatform {
		case "whatsapp":
			msg, err := sendWhatsAppText(a, conversationID, message, "")
			if err != nil {
				return errorResult(fmt.Sprintf("failed to send: %v", err)), nil
			}
			if err := a.Store.RecordOutgoingMessage(msg, ""); err != nil {
				return errorResult(fmt.Sprintf("failed to persist sent message: %v", err)), nil
			}
			return textResult(fmt.Sprintf("Message sent to %s (%s): %s", conv.Name, conversationID, message)), nil
		case "signal":
			msg, err := sendSignalText(a, conversationID, message, "")
			if err != nil {
				return errorResult(fmt.Sprintf("failed to send: %v", err)), nil
			}
			if err := a.Store.RecordOutgoingMessage(msg, ""); err != nil {
				return errorResult(fmt.Sprintf("failed to persist sent message: %v", err)), nil
			}
			return textResult(fmt.Sprintf("Message sent to %s (%s): %s", conv.Name, conversationID, message)), nil
		case "", "sms":
			// Fall through to the Google Messages client below.
		default:
			return errorResult(fmt.Sprintf("sending is not supported for platform %s via OpenMessage MCP yet", conv.SourcePlatform)), nil
		}

		cli := a.GetClient()
		if cli == nil {
			return errorResult(app.ErrNotConnected), nil
		}

		payload := app.BuildSendPayload(conversationID, message, "", "", nil)
		_, err = cli.GM.SendMessage(payload)
		if err != nil {
			return errorResult(fmt.Sprintf("failed to send: %v", err)), nil
		}

		return textResult(fmt.Sprintf("Message sent to %s (%s): %s", conv.Name, conversationID, message)), nil
	}
}
