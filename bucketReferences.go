package lib

import (
	"encoding/json"
	"strings"
)

// BucketReferences represents the unified JSON structure for XMICROSERVSRC06.
// It consolidates the previously separate SRC06, SRC07, and SRC08 hash references.
type BucketReferences struct {
	Oggetto                    string `json:"oggetto"`
	MicroserviceJsonObject     string `json:"MicroserviceJsonObject,omitempty"`
	MicroserviceMetadataObject string `json:"MicroserviceMetadataObject,omitempty"`
}

// ParseBucketReferences parses a XMICROSERVSRC06 value.
// If the value starts with '{' it is treated as JSON and unmarshalled.
// Otherwise it is treated as a legacy plain MD5 hash and placed in Oggetto.
func ParseBucketReferences(value string) (BucketReferences, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return BucketReferences{}, nil
	}

	if strings.HasPrefix(value, "{") {
		var refs BucketReferences
		if err := json.Unmarshal([]byte(value), &refs); err != nil {
			return BucketReferences{}, err
		}
		return refs, nil
	}

	// Legacy: plain hash → treat as Oggetto
	return BucketReferences{Oggetto: value}, nil
}

// ToBucketRefJSON serializes BucketReferences to a JSON string.
func ToBucketRefJSON(refs BucketReferences) (string, error) {
	b, err := json.Marshal(refs)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
