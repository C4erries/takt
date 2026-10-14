"""Pytest configuration and fixtures for Takt E2E testing suite."""

import os
import socket
import subprocess
import time
import uuid
from typing import Callable, Generator
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

    # Always rebuild server binary to ensure up-to-date code under test
    subprocess.run(
        ["go", "build", "-o", "bin/server", "./cmd/server"],
        cwd=backend_dir,
        check=True,
        capture_output=True,
    )

    env = os.environ.copy()
    env["PORT"] = str(port)
    env["ENV"] = "test"
    if "DATABASE_URL" not in env:
        env["DATABASE_URL"] = (
            "postgres://postgres:postgres@localhost:5432/takt_db?sslmode=disable"
        )

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


@pytest.fixture
def api_client(base_url: str) -> Generator[httpx.Client, None, None]:
    """
    Unauthenticated HTTP client configured with base URL and default timeout for E2E tests.
    Fresh instance per test function to prevent cookie leakage.
    """
    with httpx.Client(base_url=base_url, timeout=10.0) as client:
        yield client


@pytest.fixture
def user_factory(
    base_url: str,
) -> Callable[..., dict]:
    """
    Factory fixture to dynamically register new isolated test users.
    Returns user metadata, JWT token, and created personal workspace.
    """

    def _create_user(
        email: str | None = None,
        password: str = "secret123",
        display_name: str = "QA Tester",
    ) -> dict:
        if email is None:
            unique_suffix = uuid.uuid4().hex[:10]
            email = f"user_{unique_suffix}@example.com"

        with httpx.Client(base_url=base_url, timeout=10.0) as reg_client:
            resp = reg_client.post(
                "/api/v1/auth/register",
                json={
                    "email": email,
                    "password": password,
                    "displayName": display_name,
                },
            )
            assert resp.status_code == 201, (
                f"User registration failed ({resp.status_code}): {resp.text}"
            )
            data = resp.json()
            assert "token" in data
            assert "user" in data
            assert "workspace" in data

            return {
                "id": data["user"]["id"],
                "email": email,
                "password": password,
                "displayName": display_name,
                "token": data["token"],
                "user": data["user"],
                "workspace": data["workspace"],
            }
    return _create_user


@pytest.fixture
def auth_user(user_factory: Callable[..., dict]) -> dict:
    """Primary registered test user fixture."""
    return user_factory(display_name="Primary User")


@pytest.fixture
def auth_client(
    base_url: str, auth_user: dict
) -> Generator[httpx.Client, None, None]:
    """
    HTTP client authenticated as primary test user via Authorization: Bearer header.
    """
    headers = {"Authorization": f"Bearer {auth_user['token']}"}
    with httpx.Client(base_url=base_url, timeout=10.0, headers=headers) as client:
        yield client


@pytest.fixture
def cookie_client(
    base_url: str, auth_user: dict
) -> Generator[httpx.Client, None, None]:
    """
    HTTP client authenticated as primary test user via HTTP-only session cookie.
    """
    cookies = {"token": auth_user["token"]}
    with httpx.Client(base_url=base_url, timeout=10.0, cookies=cookies) as client:
        yield client


@pytest.fixture
def other_user(user_factory: Callable[..., dict]) -> dict:
    """Second registered test user for multi-tenancy & IDOR testing."""
    return user_factory(display_name="Other Tenant User")


@pytest.fixture
def other_client(
    base_url: str, other_user: dict
) -> Generator[httpx.Client, None, None]:
    """
    HTTP client authenticated as second test user for multi-tenant isolation tests.
    """
    headers = {"Authorization": f"Bearer {other_user['token']}"}
    with httpx.Client(base_url=base_url, timeout=10.0, headers=headers) as client:
        yield client
