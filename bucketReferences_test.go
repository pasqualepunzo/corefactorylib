package lib

import (
	"testing"
)

func TestParseBucketReferences_LegacyHash(t *testing.T) {
	hash := "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4"
	refs, err := ParseBucketReferences(hash)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if refs.Oggetto != hash {
		t.Errorf("Oggetto = %q, want %q", refs.Oggetto, hash)
	}
	if refs.MicroserviceJsonObject != "" {
		t.Errorf("MicroserviceJsonObject should be empty, got %q", refs.MicroserviceJsonObject)
	}
	if refs.MicroserviceMetadataObject != "" {
		t.Errorf("MicroserviceMetadataObject should be empty, got %q", refs.MicroserviceMetadataObject)
	}
}

func TestParseBucketReferences_JSON(t *testing.T) {
	input := `{"oggetto":"hash06","MicroserviceJsonObject":"hash07","MicroserviceMetadataObject":"hash08"}`
	refs, err := ParseBucketReferences(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if refs.Oggetto != "hash06" {
		t.Errorf("Oggetto = %q, want %q", refs.Oggetto, "hash06")
	}
	if refs.MicroserviceJsonObject != "hash07" {
		t.Errorf("MicroserviceJsonObject = %q, want %q", refs.MicroserviceJsonObject, "hash07")
	}
	if refs.MicroserviceMetadataObject != "hash08" {
		t.Errorf("MicroserviceMetadataObject = %q, want %q", refs.MicroserviceMetadataObject, "hash08")
	}
}

func TestParseBucketReferences_Empty(t *testing.T) {
	refs, err := ParseBucketReferences("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if refs.Oggetto != "" {
		t.Errorf("Oggetto should be empty, got %q", refs.Oggetto)
	}
}

func TestParseBucketReferences_JSONPartial(t *testing.T) {
	input := `{"oggetto":"hash06"}`
	refs, err := ParseBucketReferences(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if refs.Oggetto != "hash06" {
		t.Errorf("Oggetto = %q, want %q", refs.Oggetto, "hash06")
	}
	if refs.MicroserviceJsonObject != "" {
		t.Errorf("MicroserviceJsonObject should be empty")
	}
}

func TestToBucketRefJSON(t *testing.T) {
	refs := BucketReferences{
		Oggetto:                    "hash06",
		MicroserviceJsonObject:     "hash07",
		MicroserviceMetadataObject: "hash08",
	}
	s, err := ToBucketRefJSON(refs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Round-trip
	parsed, err := ParseBucketReferences(s)
	if err != nil {
		t.Fatalf("round-trip parse error: %v", err)
	}
	if parsed != refs {
		t.Errorf("round-trip mismatch: got %+v, want %+v", parsed, refs)
	}
}

func TestToBucketRefJSON_OmitsEmpty(t *testing.T) {
	refs := BucketReferences{Oggetto: "hash06"}
	s, err := ToBucketRefJSON(refs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := `{"oggetto":"hash06"}`
	if s != expected {
		t.Errorf("got %q, want %q", s, expected)
	}
}
