package model

import (
	"encoding/json"
	"time"
)

// CustomTime adalah custom type untuk parsing tanggal dengan format fleksibel
type CustomTime struct {
	time.Time
}

// UnmarshalJSON mengimplementasikan custom JSON unmarshaling untuk CustomTime
// Mendukung format:
// - "2024-10-05" (ISO date: YYYY-MM-DD)
// - "2024-10-05T15:04:05Z07:00" (RFC3339 datetime)
// - "05-10-2024" (DD-MM-YYYY)
// - "05/10/2024" (DD/MM/YYYY)
// - {"time.Time": "2024-10-05"} (nested object)
// - null (untuk field optional)
func (ct *CustomTime) UnmarshalJSON(b []byte) error {
	s := string(b)
	
	// Handle null value
	if s == "null" {
		ct.Time = time.Time{}
		return nil
	}
	
	var objMap map[string]interface{}
	if err := json.Unmarshal(b, &objMap); err == nil && len(objMap) > 0 {
		// If it's an object, extract the value
		for _, v := range objMap {
			if strVal, ok := v.(string); ok {
				s = strVal
				break
			}
		}
	}
	
	// Hapus quotes dari JSON string
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		s = s[1 : len(s)-1]
	}

	// Handle empty string
	if s == "" {
		ct.Time = time.Time{}
		return nil
	}

	// Format 1: YYYY-MM-DD (ISO date)
	t, err := time.Parse("2006-01-02", s)
	if err == nil {
		ct.Time = t
		return nil
	}

	// Format 2: RFC3339 (datetime dengan timezone)
	t, err = time.Parse(time.RFC3339, s)
	if err == nil {
		ct.Time = t
		return nil
	}

	// Format 3: DD-MM-YYYY (format lokal)
	t, err = time.Parse("02-01-2006", s)
	if err == nil {
		ct.Time = t
		return nil
	}

	// Format 4: DD/MM/YYYY
	t, err = time.Parse("02/01/2006", s)
	if err == nil {
		ct.Time = t
		return nil
	}

	// Jika semua format gagal, return error terakhir
	return err
}
