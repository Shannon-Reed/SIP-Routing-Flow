
FROM python:3.12-slim

WORKDIR /app

# Prevent Python from creating bytecode files
ENV PYTHONDONTWRITEBYTECODE=1

# Ensure Python logs appear immediately
ENV PYTHONUNBUFFERED=1

# Copy and install dependencies
COPY requirements.txt ./
RUN pip install --no-cache-dir -r requirements.txt

# Copy the FastAPI application
COPY app/ ./app/

# Run without root privileges
USER 10001:10001

EXPOSE 8000

CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "8000"]
