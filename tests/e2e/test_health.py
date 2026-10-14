"""E2E tests for Takt health check endpoints (/health and /api/v1/health)."""

import httpx


def test_root_health_endpoint(api_client: httpx.Client) -> None:
    """
    Test GET /health root healthcheck endpoint.
    Verifies HTTP 200 OK, status 'ok' or 'degraded', version, uptime_seconds, and database.status.
    """
    response = api_client.get("/health")

    assert response.status_code == 200, f"Expected 200 OK, got {response.status_code}: {response.text}"
    assert "application/json" in response.headers.get("content-type", "")

    payload = response.json()
    assert isinstance(payload, dict), f"Payload must be a JSON object, got {type(payload)}"

    # Required fields
    assert "status" in payload, "Missing 'status' in health response"
    assert payload["status"] in ["ok", "degraded"], f"Unexpected status: {payload['status']}"

    assert "version" in payload, "Missing 'version' in health response"
    assert isinstance(payload["version"], str) and len(payload["version"]) > 0

    assert "uptime_seconds" in payload, "Missing 'uptime_seconds' in health response"
    assert isinstance(payload["uptime_seconds"], int) and payload["uptime_seconds"] >= 0

    assert "database" in payload, "Missing 'database' object in health response"
    assert isinstance(payload["database"], dict), "Database field must be a JSON object"
    assert "status" in payload["database"], "Missing 'status' inside database object"
    assert payload["database"]["status"] in ["connected", "disconnected"]


def test_api_v1_health_endpoint(api_client: httpx.Client) -> None:
    """
    Test GET /api/v1/health endpoint per OpenAPI specification.
    Verifies response conforms to OpenAPI schema: status, version, uptime_seconds, and database.
    """
    response = api_client.get("/api/v1/health")

    assert response.status_code == 200, f"Expected 200 OK, got {response.status_code}: {response.text}"
    assert "application/json" in response.headers.get("content-type", "")

    payload = response.json()
    assert isinstance(payload, dict), f"Payload must be a JSON object, got {type(payload)}"

    # Schema conformance checks
    assert payload["status"] in ["ok", "degraded", "error"], f"Invalid status: {payload['status']}"
    assert isinstance(payload["version"], str) and len(payload["version"]) > 0
    assert isinstance(payload["uptime_seconds"], int) and payload["uptime_seconds"] >= 0

    assert "database" in payload, "Field 'database' is required by OpenAPI schema"
    assert isinstance(payload["database"], dict)
    assert payload["database"]["status"] in ["connected", "disconnected"]


def test_health_endpoint(api_client: httpx.Client) -> None:
    """
    Verify DoD criteria: endpoint /api/v1/health responds with 200 OK.
    """
    response = api_client.get("/api/v1/health")
    assert response.status_code == 200
    assert response.json()["status"] in ["ok", "degraded"]
