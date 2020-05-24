package langserver

import (
	"encoding/json"
	"log"
)

// Settings contains settings for the language-server
type Settings struct {
	Yolol YololSettings
}

// YololSettings contains settings specific to yolol
type YololSettings struct {
	Formatting FormatSettings
}

// FormatSettings contains formatting settings
type FormatSettings struct {
	Mode string
}

func (s *Settings) Read(inp interface{}) error {
	by, err := json.Marshal(inp)
	if err != nil {
		return err
	}
	log.Println(string(by))
	return json.Unmarshal(by, s)
}

// DefaultSettings returns the default-settings for the server
func DefaultSettings() *Settings {
	return &Settings{
		Yolol: YololSettings{
			Formatting: FormatSettings{
				Mode: "Compact",
			},
		},
	}
}
