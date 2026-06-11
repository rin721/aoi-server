package model

type ConfigSnapshot struct {
	Sections []ConfigSection `json:"sections"`
}

type ConfigSection struct {
	Code        string       `json:"code"`
	Description string       `json:"description"`
	Icon        string       `json:"icon"`
	Items       []ConfigItem `json:"items"`
	Label       string       `json:"label"`
	Order       int          `json:"order"`
}

type ConfigItem struct {
	Description string `json:"description"`
	Key         string `json:"key"`
	Label       string `json:"label"`
	Secret      bool   `json:"secret"`
	Source      string `json:"source"`
	Value       any    `json:"value"`
}
