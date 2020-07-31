package langserver

import (
	"encoding/json"
)

const (
	FormatModeCompact       = "Compact"
	FormatModeReadale       = "Readable"
	LengthCheckModeStrict   = "Strict"
	LengthCheckModeOptimize = "Optimize"
	LengthCheckModeOff      = "Off"
)

// Settings contains settings for the language-server
type Settings struct {
	Yolol YololSettings `json:"yolol"`
}

// YololSettings contains settings specific to yolol
type YololSettings struct {
	Formatting     FormatSettings      `json:"formatting"`
	LengthChecking LengthCheckSettings `json:"lengthChecking"`
}

// FormatSettings contains formatting settings
type FormatSettings struct {
	Mode string `json:"mode"`
}

// LengthCheckSettings contains settings for the lenght-validation
type LengthCheckSettings struct {
	Mode string `json:"mode"`
}

func (s *Settings) Read(inp interface{}) error {
	by, err := json.Marshal(inp)
	if err != nil {
		return err
	}
	return json.Unmarshal(by, s)
}

// DefaultSettings returns the default-settings for the server
func DefaultSettings() *Settings {
	return &Settings{
		Yolol: YololSettings{
			Formatting: FormatSettings{
				Mode: FormatModeCompact,
			},
			LengthChecking: LengthCheckSettings{
				Mode: LengthCheckModeStrict,
			},
		},
	}
}
