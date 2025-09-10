package gpt

type YandexBody struct {
	ModelUri          string                  `json:"modelUri"`
	CompletionOptions YandexCompletionOptions `json:"completionOptions"`
	Messages          []YandexMessage         `json:"messages"`
}

type YandexCompletionOptions struct {
	MaxTokens   int     `json:"maxTokens"`
	Temperature float64 `json:"temperature"`
}

type YandexMessage struct {
	Role string `json:"role"`
	Text string `json:"text"`
}

type YandexResponse struct {
	Result YandexResult `json:"result"`
}

type YandexResult struct {
	Alternatives []YandexAlternative `json:"alternatives"`
}

type YandexAlternative struct {
	Message YandexMessage `json:"message"`
	Status  string        `json:"status"`
}
