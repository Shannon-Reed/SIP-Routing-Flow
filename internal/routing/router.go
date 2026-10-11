package routing

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CallRequest struct {
	CallID      string `json:"call_id"`
	Caller      string `json:"caller"`
	Destination string `json:"destination"`
	Region      string `json:"region"`
}

type RouteResponse struct {
	CallID string `json:"call_id"`
	Route  string `json:"route"`
	Status string `json:"status"`
}

type Router struct {
	DB *pgxpool.Pool
}

func NewRouter(db *pgxpool.Pool) *Router {
	return &Router{DB: db}
}

func (router *Router) FindRoute(
	ctx context.Context,
	call CallRequest,
) (RouteResponse, error) {

	if strings.TrimSpace(call.CallID) == "" {
		return RouteResponse{}, errors.New("call_id is required")
	}

	if strings.TrimSpace(call.Destination) == "" {
		return RouteResponse{}, errors.New("destination is required")
	}

	region := strings.ToUpper(strings.TrimSpace(call.Region))
	if region == "" {
		region = "DEFAULT"
	}

	const query = `
		SELECT sip_uri
		FROM sip_routes
		WHERE enabled = TRUE
		  AND region IN ($1, 'DEFAULT')
		ORDER BY CASE WHEN region = $1 THEN 0 ELSE 1 END
		LIMIT 1
	`

	var sipURI string

	err := router.DB.QueryRow(ctx, query, region).Scan(&sipURI)

	if errors.Is(err, pgx.ErrNoRows) {
		return RouteResponse{
			CallID: call.CallID,
			Status: "NO_ROUTE",
		}, nil
	}

	if err != nil {
		return RouteResponse{}, fmt.Errorf("route lookup failed: %w", err)
	}

	return RouteResponse{
		CallID: call.CallID,
		Route:  sipURI,
		Status: "ROUTE_FOUND",
	}, nil
}
