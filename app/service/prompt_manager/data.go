package prompt_manager

import "context"

type promptHandle struct {
	counter   int
	messageID int
	cancel    context.CancelFunc
}
