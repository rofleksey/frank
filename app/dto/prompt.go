package dto

import "github.com/google/uuid"

type Attachment struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

type Prompt struct {
	ID          uuid.UUID    `json:"id"`
	Text        string       `json:"text"`
	Depth       int          `json:"depth"`
	Attachments []Attachment `json:"attachments"`
}

func (p *Prompt) BranchWithNewText(text string) Prompt {
	attachmentsCopy := make([]Attachment, len(p.Attachments))
	copy(attachmentsCopy, p.Attachments)

	return Prompt{
		ID:          p.ID,
		Text:        text,
		Depth:       p.Depth + 1,
		Attachments: attachmentsCopy,
	}
}

func (p *Prompt) BranchWithNewAttachment(newAttachment Attachment) Prompt {
	attachmentsCopy := make([]Attachment, len(p.Attachments))
	copy(attachmentsCopy, p.Attachments)

	attachmentsCopy = append(attachmentsCopy, newAttachment)

	return Prompt{
		ID:          p.ID,
		Text:        p.Text,
		Depth:       p.Depth + 1,
		Attachments: attachmentsCopy,
	}
}
