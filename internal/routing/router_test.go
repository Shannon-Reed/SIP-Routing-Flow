package routing

import "testing"

func TestFindRoute(t *testing.T) {
	router := NewRouter()

	tests := []struct {
		name          string
		call          CallRequest
		expectedRoute string
		expectError   bool
	}{
		{
			name: "US call",
			call: CallRequest{
				CallID:      "CALL-1001",
				Destination: "+18005550100",
				Region:      "US",
			},
			expectedRoute: "sip:us-gateway@pbx.internal",
		},
		{
			name: "EU call",
			call: CallRequest{
				CallID:      "CALL-1002",
				Destination: "+442071234567",
				Region:      "EU",
			},
			expectedRoute: "sip:eu-gateway@pbx.internal",
		},
		{
			name: "Unknown region uses default",
			call: CallRequest{
				CallID:      "CALL-1003",
				Destination: "+18005550100",
				Region:      "APAC",
			},
			expectedRoute: "sip:default@pbx.internal",
		},
		{
			name: "Region is case insensitive",
			call: CallRequest{
				CallID:      "CALL-1004",
				Destination: "+18005550100",
				Region:      "us",
			},
			expectedRoute: "sip:us-gateway@pbx.internal",
		},
		{
			name: "Missing call ID",
			call: CallRequest{
				Destination: "+18005550100",
				Region:      "US",
			},
			expectError: true,
		},
		{
			name: "Missing destination",
			call: CallRequest{
				CallID: "CALL-1005",
				Region: "US",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := router.FindRoute(tt.call)

			if tt.expectError {
				if err == nil {
					t.Fatal("Expected an error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if response.Route != tt.expectedRoute {
				t.Errorf(
					"Expected route %s, got %s",
					tt.expectedRoute,
					response.Route,
				)
			}

			if response.Status != "ROUTE_FOUND" {
				t.Errorf("Unexpected status: %s", response.Status)
			}
		})
	}
}
