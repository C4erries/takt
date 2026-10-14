"""Pytest configuration and fixtures for Takt E2E testing suite."""

import os
import socket
import subprocess
import time
from typing import Generator
import httpx
import pytest

DEFAULT_BASE_URL = "http://localhost:8080"


def find_free_port() -> int:
    """Find and return an available TCP port on localhost."""
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
        s.bind(("", 0))
        return s.getsockname()[1]


def is_service_healthy(url: str) -> bool:
    """Check if the Takt service is reachable and responding to health check."""
    try:
        with httpx.Client(timeout=1.0) as client:
            resp = client.get(f"{url}/health")
            if resp.status_code == 200:
                data = resp.json()
                return isinstance(data, dict) and "database" in data
    except Exception:
        return False
    return False


@pytest.fixture(scope="session")
def base_url() -> Generator[str, None, None]:
    """
    Resolve base URL for the API under test.
    Checks APP_URL and API_BASE_URL environment variables.
    If default http://localhost:8080 is not a running Takt service,
    automatically starts a local backend instance for the test session.
    """
    env_url = os.getenv("APP_URL") or os.getenv("API_BASE_URL")
    if env_url:
        yield env_url.rstrip("/")
        return

    # Check if default URL is already running Takt
    if is_service_healthy(DEFAULT_BASE_URL):
        yield DEFAULT_BASE_URL
        return

    # Fallback for local development when localhost:8080 is occupied or not running
    port = find_free_port()
    local_url = f"http://localhost:{port}"

    project_root = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
    backend_dir = os.path.join(project_root, "backend")
    binary_path = os.path.join(backend_dir, "bin", "server")

    # Build binary if not already present
    if not os.path.isfile(binary_path):
        subprocess.run(
            ["go", "build", "-o", "bin/server", "./cmd/server"],
            cwd=backend_dir,
            check=True,
            capture_output=True,
        )

    env = os.environ.copy()
    env["PORT"] = str(port)
    env["ENV"] = "test"

    proc = subprocess.Popen(
        [binary_path],
        cwd=backend_dir,
        env=env,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
    )

    ready = False
    deadline = time.time() + 5.0
    while time.time() < deadline:
        if is_service_healthy(local_url):
            ready = True
            break
        if proc.poll() is not None:
            break
        time.sleep(0.1)

    if not ready:
        proc.terminate()
        try:
            proc.wait(timeout=2)
        except subprocess.TimeoutExpired:
            proc.kill()
        out, err = proc.communicate()
        raise RuntimeError(
            f"Failed to start local Takt server: {err.decode('utf-8', errors='ignore')}"
        )

    try:
        yield local_url
    finally:
        proc.terminate()
        try:
            proc.wait(timeout=3)
        except subprocess.TimeoutExpired:
            proc.kill()


@pytest.fixture(scope="session")
def api_client(base_url: str) -> Generator[httpx.Client, None, None]:
    """
    HTTP client configured with base URL and default timeout for E2E tests.
    """
    with httpx.Client(base_url=base_url, timeout=10.0) as client:
        yield client
