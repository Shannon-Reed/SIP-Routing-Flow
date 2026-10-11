package routing

import (
	"errors"
	"strings"
)

// CallRequest contains information needed to route a call.
type CallRequest struct {
	CallID      string `json:"call_id"`
	Caller      string `json:"caller"`
	Destination string `json:"destination"`
	Region      string `json:"region"`
}

// RouteResponse contains the routing decision.
type RouteResponse struct {
	CallID string `json:"call_id"`
	Route  string `json:"route"`
	Status string `json:"status"`
}

// Router manages SIP routing decisions.
type Router struct {
	Routes map[string]string
}

// NewRouter creates a router with configured destinations.
func NewRouter() *Router {
	return &Router{
		Routes: map[string]string{
			"US": "sip:us-gateway@pbx.internal",
			"EU": "sip:eu-gateway@pbx.internal",
		},
	}
}

// FindRoute selects the SIP destination for a call.
func (router *Router) FindRoute(call CallRequest) (RouteResponse, error) {
	if strings.TrimSpace(call.CallID) == "" {
		return RouteResponse{}, errors.New("call_id is required")
	}

	if strings.TrimSpace(call.Destination) == "" {
		return RouteResponse{}, errors.New("destination is required")
	}

	region := strings.ToUpper(strings.TrimSpace(call.Region))

	route, exists := router.Routes[region]
	if !exists {
		route = "sip:default@pbx.internal"
	}

	return RouteResponse{
		CallID: call.CallID,
		Route:  route,
		Status: "ROUTE_FOUND",
	}, nil
}
