package langserver

import (
	"encoding/json"
	"strings"

	"github.com/dbaumgarten/yodk/pkg/validators"
)

const (
	FormatModeSpaceless     = "Spaceless"
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
	ChipType       string              `json:"chipType"`
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
	err = json.Unmarshal(by, s)
	if err != nil {
		return err
	}

	s.Yolol.ChipType = strings.ToLower(s.Yolol.ChipType)

	return nil
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
			ChipType: validators.ChipTypeAuto,
		},
	}
}
