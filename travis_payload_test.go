package hookworm_test

import (
	"testing"
)

func TestDeserializesTravisPayloads(t *testing.T) {
	payload := getTravisPayload("success")
	if payload.ID == 0 {
		t.Errorf("ID not deserialized: %+v\n", payload)
	}
	if payload.Repository == nil {
		t.Errorf("Repository not deserialized: %+v\n", payload)
	}
	if payload.Repository.ID == 0 {
		t.Errorf("Repository.ID not deserialized: %+v\n", payload)
	}
	if payload.Repository.Name == "" {
		t.Errorf("Repository.Name not deserialized: %+v\n", payload)
	}
	if payload.Repository.OwnerName == "" {
		t.Errorf("Repository.OwnerName not deserialized: %+v\n", payload)
	}
	if payload.Repository.URL == "" {
		t.Errorf("Repository.URL not deserialized: %+v\n", payload)
	}
	if payload.Number == "" {
		t.Errorf("Number not deserialized: %+v\n", payload)
	}
	if payload.Config == nil {
		t.Errorf("Config not deserialized: %+v\n", payload)
	}
	// Status and Result are a bit tricky, since the values in the "success"
	// payload are zero, but zero is also the default for uninitialized
	// integers.  Hm.
	if payload.Status != 0 {
		t.Errorf("Status not correctly deserialized: %+v\n", payload)
	}
	if payload.Result != 0 {
		t.Errorf("Result not correctly deserialized: %+v\n", payload)
	}
	if payload.StatusMessage == "" {
		t.Errorf("StatusMessage not deserialized: %+v\n", payload)
	}
	if payload.ResultMessage == "" {
		t.Errorf("ResultMessage not deserialized: %+v\n", payload)
	}
	if payload.StartedAt == nil {
		t.Errorf("StartedAt not deserialized: %+v\n", payload)
	}
	if payload.FinishedAt == nil {
		t.Errorf("FinishedAt not deserialized: %+v\n", payload)
	}
}
