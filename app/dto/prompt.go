package dto

import (
	"context"

	"github.com/google/uuid"
)

type Attachment struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

type Prompt struct {
	ID          uuid.UUID    `json:"id"`
	MessageID   int          `json:"message_id"`
	Text        string       `json:"text"`
	Depth       int          `json:"depth"`
	TextHistory []string     `json:"text_history"`
	Attachments []Attachment `json:"attachments"`

	Ctx    context.Context    `json:"-"`
	Cancel context.CancelFunc `json:"-"`
}

func (p *Prompt) BranchWithNewText(text string) Prompt {
	attachmentsCopy := make([]Attachment, len(p.Attachments))
	copy(attachmentsCopy, p.Attachments)

	textHistoryCopy := make([]string, len(p.TextHistory)+1)
	copy(textHistoryCopy[1:], p.TextHistory)
	textHistoryCopy[0] = p.Text

	return Prompt{
		ID:          p.ID,
		MessageID:   p.MessageID,
		Text:        text,
		Depth:       p.Depth + 1,
		TextHistory: textHistoryCopy,
		Attachments: attachmentsCopy,
		Ctx:         p.Ctx,
		Cancel:      p.Cancel,
	}
}

func (p *Prompt) BranchWithNewAttachment(newAttachment Attachment) Prompt {
	attachmentsCopy := make([]Attachment, len(p.Attachments)+1)
	copy(attachmentsCopy[1:], p.Attachments)
	attachmentsCopy[0] = newAttachment

	textHistoryCopy := make([]string, len(p.TextHistory))
	copy(textHistoryCopy, p.TextHistory)

	return Prompt{
		ID:          p.ID,
		MessageID:   p.MessageID,
		Text:        p.Text,
		Depth:       p.Depth + 1,
		TextHistory: textHistoryCopy,
		Attachments: attachmentsCopy,
		Ctx:         p.Ctx,
		Cancel:      p.Cancel,
	}
}

func (p *Prompt) CancelAllBranches() {
	p.Cancel()
}
