
import logging
import psycopg
from fastapi import FastAPI, HTTPException, Query
from .database import get_connection


logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

app = FastAPI(
    title="SIP Routing Flow Analytics API",
    description=(
        "Analytics service for SIP routing decisions "
        "stored in PostgreSQL."
    ),
    version="0.1.0",
)


@app.get("/health")
def health():
    return {
        "status": "healthy",
        "service": "sip-routing-analytics",
    }


@app.get("/analytics/summary")
def get_summary():
    query = """
        SELECT
            COUNT(*) AS total_attempts,
            COUNT(*) FILTER (
                WHERE routing_status = 'ROUTE_FOUND'
            ) AS routes_found,
            COUNT(*) FILTER (
                WHERE routing_status = 'NO_ROUTE'
            ) AS no_route,
            ROUND(
                100.0 * COUNT(*) FILTER (
                    WHERE routing_status = 'ROUTE_FOUND'
                ) / NULLIF(COUNT(*), 0),
                2
            ) AS route_found_percentage
        FROM call_records
    """

    try:
        with get_connection() as conn:
            with conn.cursor() as cursor:
                cursor.execute(query)
                result = cursor.fetchone()

        return {
            **result,
            "route_found_percentage": float(
                result["route_found_percentage"] or 0
            ),
        }

    except psycopg.Error:
        logger.exception("Failed to retrieve routing summary")
        raise HTTPException(
            status_code=503,
            detail="Analytics database unavailable",
        )


@app.get("/analytics/regions")
def get_regions():
    query = """
        SELECT
            COALESCE(NULLIF(UPPER(TRIM(region)), ''), 'DEFAULT')
                AS region,
            COUNT(*) AS total_attempts,
            COUNT(*) FILTER (
                WHERE routing_status = 'ROUTE_FOUND'
            ) AS routes_found,
            COUNT(*) FILTER (
                WHERE routing_status = 'NO_ROUTE'
            ) AS no_route
        FROM call_records
        GROUP BY 1
        ORDER BY total_attempts DESC, region
    """

    try:
        with get_connection() as conn:
            with conn.cursor() as cursor:
                cursor.execute(query)
                return cursor.fetchall()

    except psycopg.Error:
        logger.exception("Failed to retrieve regional analytics")
        raise HTTPException(
            status_code=503,
            detail="Analytics database unavailable",
        )


@app.get("/analytics/recent")
def get_recent(
    limit: int = Query(default=10, ge=1, le=100),
):
    query = """
        SELECT
            call_id,
            region,
            selected_route,
            routing_status,
            created_at
        FROM call_records
        ORDER BY created_at DESC, id DESC
        LIMIT %s
    """

    try:
        with get_connection() as conn:
            with conn.cursor() as cursor:
                cursor.execute(query, (limit,))
                records = cursor.fetchall()

        return records

    except psycopg.Error:
        logger.exception("Failed to retrieve recent routing records")
        raise HTTPException(
            status_code=503,
            detail="Analytics database unavailable",
        )
