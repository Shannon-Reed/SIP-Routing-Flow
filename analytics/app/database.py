
import os
from pathlib import Path

import psycopg
from dotenv import load_dotenv
from psycopg.rows import dict_row


# Locate the repository root from analytics/app/database.py
PROJECT_ROOT = Path(__file__).resolve().parents[2]

# Reuse the existing root .env file used by the Go API
load_dotenv(PROJECT_ROOT / ".env")


def get_connection():
    database_url = os.getenv("DATABASE_URL")

    if not database_url:
        raise RuntimeError(
            "DATABASE_URL environment variable is missing"
        )

    return psycopg.connect(
        database_url,
        row_factory=dict_row,
        connect_timeout=5,
    )
