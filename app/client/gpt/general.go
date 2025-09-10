package gpt

type Options struct {
	MaxTokens   int     `json:"maxTokens"`
	Temperature float64 `json:"temperature"`
}

type Prompt struct {
	SystemText  string  `json:"systemText"`
	Text        string  `json:"text"`
	Temperature float64 `json:"temperature"`
}
