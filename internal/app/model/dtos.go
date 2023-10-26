package model

type (
	ShortenRequestDto struct {
		URL string `json:"url"`
	}
	ShortenResponseDto struct {
		Result string `json:"result"`
	}
)
