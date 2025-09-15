package prompt_manager

import (
	"context"
	"frank/app/dto"
	"frank/app/service/telegram_reply"
	"frank/pkg/config"
	"sync"

	"github.com/google/uuid"
	"github.com/samber/do"
)

type Service struct {
	appCtx       context.Context
	cfg          *config.Config
	replyService *telegram_reply.Service

	handleMap map[uuid.UUID]*promptHandle
	mu        sync.Mutex
}

func New(di *do.Injector) (*Service, error) {
	return &Service{
		appCtx:       do.MustInvoke[context.Context](di),
		cfg:          do.MustInvoke[*config.Config](di),
		replyService: do.MustInvoke[*telegram_reply.Service](di),
		handleMap:    make(map[uuid.UUID]*promptHandle),
	}, nil
}

func (s *Service) CreatePrompt(messageID int, text string) dto.Prompt {
	ctx, cancel := context.WithCancel(s.appCtx)

	prompt := dto.Prompt{
		ID:          uuid.New(),
		MessageID:   messageID,
		Text:        text,
		Depth:       0,
		TextHistory: nil,
		Attachments: nil,
		Ctx:         ctx,
		Cancel:      cancel,
	}

	s.handleMap[prompt.ID] = &promptHandle{
		counter:   0,
		messageID: messageID,
		cancel:    prompt.Cancel,
	}

	return prompt
}

func (s *Service) IncPromptCounter(id uuid.UUID) {
	s.mu.Lock()
	defer s.mu.Unlock()

	handle, ok := s.handleMap[id]
	if !ok {
		return
	}

	if handle.counter == 0 {
		s.replyService.SetReaction(s.appCtx, handle.messageID, "👀")
	}
	handle.counter++
}

func (s *Service) DecPromptCounter(id uuid.UUID) {
	s.mu.Lock()
	defer s.mu.Unlock()

	handle, ok := s.handleMap[id]
	if !ok {
		return
	}

	handle.counter--
	if handle.counter == 0 {
		handle.cancel()
		delete(s.handleMap, id)

		s.replyService.SetReaction(s.appCtx, handle.messageID, "👍")
	}
}

func (s *Service) CancelAll() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, handle := range s.handleMap {
		handle.cancel()
	}
}
