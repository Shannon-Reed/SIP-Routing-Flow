
from datetime import datetime, timezone
from decimal import Decimal
from unittest.mock import MagicMock, patch

import psycopg
import pytest
from fastapi.testclient import TestClient

from app.main import app


client = TestClient(app)


@pytest.fixture
def mock_database():
    """
    Replace the PostgreSQL connection with a mock.

    No real database connection is created.
    """
    with patch("app.main.get_connection") as mock_connect:
        connection = MagicMock()
        cursor = MagicMock()

        # Support:
        # with get_connection() as conn:
        #     with conn.cursor() as cursor:
        mock_connect.return_value.__enter__.return_value = connection
        connection.cursor.return_value.__enter__.return_value = cursor

        yield cursor


def test_health():
    response = client.get("/health")

    assert response.status_code == 200
    assert response.json() == {
        "status": "healthy",
        "service": "sip-routing-analytics",
    }


def test_summary_success(mock_database):
    mock_database.fetchone.return_value = {
        "total_attempts": 10,
        "routes_found": 8,
        "no_route": 2,
        "route_found_percentage": Decimal("80.00"),
    }

    response = client.get("/analytics/summary")

    assert response.status_code == 200

    data = response.json()

    assert data["total_attempts"] == 10
    assert data["routes_found"] == 8
    assert data["no_route"] == 2
    assert data["route_found_percentage"] == 80.0

    mock_database.execute.assert_called_once()
    mock_database.fetchone.assert_called_once()


def test_summary_empty_database(mock_database):
    mock_database.fetchone.return_value = {
        "total_attempts": 0,
        "routes_found": 0,
        "no_route": 0,
        "route_found_percentage": None,
    }

    response = client.get("/analytics/summary")

    assert response.status_code == 200

    data = response.json()

    assert data["total_attempts"] == 0
    assert data["routes_found"] == 0
    assert data["no_route"] == 0
    assert data["route_found_percentage"] == 0.0


def test_regions_success(mock_database):
    mock_database.fetchall.return_value = [
        {
            "region": "US",
            "total_attempts": 6,
            "routes_found": 5,
            "no_route": 1,
        },
        {
            "region": "EU",
            "total_attempts": 4,
            "routes_found": 3,
            "no_route": 1,
        },
    ]

    response = client.get("/analytics/regions")

    assert response.status_code == 200

    data = response.json()

    assert len(data) == 2
    assert data[0]["region"] == "US"
    assert data[0]["total_attempts"] == 6
    assert data[1]["region"] == "EU"
    assert data[1]["routes_found"] == 3

    mock_database.execute.assert_called_once()
    mock_database.fetchall.assert_called_once()


def test_regions_empty_database(mock_database):
    mock_database.fetchall.return_value = []

    response = client.get("/analytics/regions")

    assert response.status_code == 200
    assert response.json() == []


def test_recent_success(mock_database):
    mock_database.fetchall.return_value = [
        {
            "call_id": "CALL-3001",
            "region": "US",
            "selected_route": "sip:us-backup@pbx.internal",
            "routing_status": "ROUTE_FOUND",
            "created_at": datetime(
                2026, 10, 12, 7, 36, 9,
                tzinfo=timezone.utc,
            ),
        }
    ]

    response = client.get("/analytics/recent?limit=5")

    assert response.status_code == 200

    data = response.json()

    assert len(data) == 1
    assert data[0]["call_id"] == "CALL-3001"
    assert data[0]["routing_status"] == "ROUTE_FOUND"
    assert data[0]["selected_route"] == "sip:us-backup@pbx.internal"
    assert data[0]["created_at"] == "2026-10-12T07:36:09+00:00"

    # Verify that the limit is passed as a SQL parameter,
    # rather than interpolated into the SQL string.
    args, kwargs = mock_database.execute.call_args

    assert args[1] == (5,)


@pytest.mark.parametrize("limit", [0, -1, 101])
def test_recent_invalid_limit(limit):
    response = client.get(f"/analytics/recent?limit={limit}")

    assert response.status_code == 422


def test_recent_non_numeric_limit():
    response = client.get("/analytics/recent?limit=abc")

    assert response.status_code == 422


def test_recent_default_limit(mock_database):
    mock_database.fetchall.return_value = []

    response = client.get("/analytics/recent")

    assert response.status_code == 200

    args, kwargs = mock_database.execute.call_args

    assert args[1] == (10,)


def test_summary_database_failure():
    """
    Simulate PostgreSQL being unavailable.
    """
    with patch(
        "app.main.get_connection",
        side_effect=psycopg.OperationalError(
            "Simulated database connection failure"
        ),
    ):
        response = client.get("/analytics/summary")

    assert response.status_code == 503
    assert response.json() == {
        "detail": "Analytics database unavailable"
    }


def test_regions_database_failure():
    with patch(
        "app.main.get_connection",
        side_effect=psycopg.OperationalError(
            "Simulated database connection failure"
        ),
    ):
        response = client.get("/analytics/regions")

    assert response.status_code == 503


def test_recent_database_failure():
    with patch(
        "app.main.get_connection",
        side_effect=psycopg.OperationalError(
            "Simulated database connection failure"
        ),
    ):
        response = client.get("/analytics/recent")

    assert response.status_code == 503
