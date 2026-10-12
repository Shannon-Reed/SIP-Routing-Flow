package routing

import (
	"context"
	"testing"
)

func TestFindRouteValidation(t *testing.T) {
	// A nil database is safe here because invalid requests
	// are rejected before any SQL queries execute.
	router := NewRouter(nil)

	tests := []struct {
		name          string
		call          CallRequest
		expectedError string
	}{
		{
			name: "Missing Call ID",
			call: CallRequest{
				Caller:      "+14045550123",
				Destination: "+18005550100",
				Region:      "US",
			},
			expectedError: "call_id is required",
		},
		{
			name: "Missing Destination",
			call: CallRequest{
				CallID: "CALL-1001",
				Caller: "+14045550123",
				Region: "US",
			},
			expectedError: "destination is required",
		},
		{
			name: "Empty Call ID",
			call: CallRequest{
				CallID:      "",
				Destination: "+18005550100",
				Region:      "US",
			},
			expectedError: "call_id is required",
		},
		{
			name: "Whitespace Call ID",
			call: CallRequest{
				CallID:      "   ",
				Destination: "+18005550100",
				Region:      "US",
			},
			expectedError: "call_id is required",
		},
		{
			name: "Whitespace Destination",
			call: CallRequest{
				CallID:      "CALL-1002",
				Destination: "   ",
				Region:      "US",
			},
			expectedError: "destination is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := router.FindRoute(
				context.Background(),
				tt.call,
			)

			if err == nil {
				t.Fatalf(
					"Expected error %q, but received nil",
					tt.expectedError,
				)
			}

			if err.Error() != tt.expectedError {
				t.Errorf(
					"Expected error %q, got %q",
					tt.expectedError,
					err.Error(),
				)
			}

			if response.Status != "" {
				t.Errorf(
					"Expected empty response, got status %q",
					response.Status,
				)
			}
		})
	}
}
