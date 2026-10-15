"""E2E tests for Takt Authentication, Onboarding, and Session endpoints."""

import uuid
import httpx
import pytest

ONBOARDING_TASK_TITLES = [
    "Нажми J и K для клавиатурного перемещения по строкам",
    "Нажми Пробел или кликни круглый чекбокс, чтобы завершить задачу",
    "Нажми C в любом месте экрана для мгновенного сброса мысли (Quick Capture)",
]


def test_registration_success(api_client: httpx.Client) -> None:
    """
    Test user registration flow.
    Verifies 201 Created, Set-Cookie header with HttpOnly token,
    and automatic creation of 'Личное' personal workspace with owner role.
    """
    unique_suffix = uuid.uuid4().hex[:8]
    email = f"test_reg_{unique_suffix}@example.com"
    password = "secret_password_123"
    display_name = "Такт Пользователь"

    resp = api_client.post(
        "/api/v1/auth/register",
        json={
            "email": email,
            "password": password,
            "displayName": display_name,
        },
    )

    assert resp.status_code == 201, f"Expected 201 Created, got {resp.status_code}: {resp.text}"

    # Verify session cookie was set with HttpOnly and Lax
    set_cookie = resp.headers.get("set-cookie", "")
    assert "token=" in set_cookie
    assert "httponly" in set_cookie.lower()
    assert "samesite=lax" in set_cookie.lower()

    payload = resp.json()
    assert "token" in payload and len(payload["token"]) > 20
    assert "user" in payload
    assert "workspace" in payload

    user = payload["user"]
    assert user["email"] == email
    assert user["displayName"] == display_name
    assert "id" in user
    assert "createdAt" in user

    workspace = payload["workspace"]
    assert workspace["name"] == "Личное"
    assert workspace["role"] == "owner"
    assert "id" in workspace


def test_onboarding_starter_tasks_in_inbox(
    api_client: httpx.Client,
    base_url: str,
) -> None:
    """
    Verify onboarding experience: immediately after registration,
    the user's Inbox contains exactly 3 starter tutorial tasks in proper order.
    """
    unique_suffix = uuid.uuid4().hex[:8]
    email = f"onboard_{unique_suffix}@example.com"

    reg_resp = api_client.post(
        "/api/v1/auth/register",
        json={
            "email": email,
            "password": "password123",
            "displayName": "Onboarding Tester",
        },
    )
    assert reg_resp.status_code == 201
    data = reg_resp.json()
    token = data["token"]
    workspace_id = data["workspace"]["id"]

    # Fetch inbox tasks using Bearer token
    headers = {"Authorization": f"Bearer {token}"}
    with httpx.Client(base_url=base_url, timeout=10.0, headers=headers) as client:
        inbox_resp = client.get("/api/v1/tasks/inbox")

    assert inbox_resp.status_code == 200, f"Failed to get inbox: {inbox_resp.text}"
    tasks = inbox_resp.json()
    assert isinstance(tasks, list)
    assert len(tasks) == 3, f"Expected exactly 3 starter onboarding tasks, got {len(tasks)}"

    for i, expected_title in enumerate(ONBOARDING_TASK_TITLES):
        task = tasks[i]
        assert task["title"] == expected_title
        assert task["position"] == i
        assert task["isCompleted"] is False
        assert task["completedAt"] is None
        assert task["workspaceId"] == workspace_id


def test_login_flow(api_client: httpx.Client) -> None:
    """
    Test user login with valid credentials, invalid password, and nonexistent email.
    Verifies 200 OK + cookie vs 401 Unauthorized.
    """
    unique_suffix = uuid.uuid4().hex[:8]
    email = f"login_{unique_suffix}@example.com"
    password = "correct_password_123"

    # Register first
    reg_resp = api_client.post(
        "/api/v1/auth/register",
        json={"email": email, "password": password, "displayName": "Login Test"},
    )
    assert reg_resp.status_code == 201

    # 1. Valid login
    login_resp = api_client.post(
        "/api/v1/auth/login",
        json={"email": email, "password": password},
    )
    assert login_resp.status_code == 200, f"Login failed: {login_resp.text}"
    payload = login_resp.json()
    assert "token" in payload
    assert payload["user"]["email"] == email
    assert payload["workspace"]["name"] == "Личное"
    assert "token=" in login_resp.headers.get("set-cookie", "")

    # 2. Invalid password
    bad_pass_resp = api_client.post(
        "/api/v1/auth/login",
        json={"email": email, "password": "wrong_password_999"},
    )
    assert bad_pass_resp.status_code == 401
    assert bad_pass_resp.json().get("code") == "UNAUTHORIZED"

    # 3. Nonexistent user
    no_user_resp = api_client.post(
        "/api/v1/auth/login",
        json={"email": f"ghost_{unique_suffix}@example.com", "password": password},
    )
    assert no_user_resp.status_code == 401
    assert no_user_resp.json().get("code") == "UNAUTHORIZED"


def test_me_endpoint_authorization(
    api_client: httpx.Client,
    auth_client: httpx.Client,
    cookie_client: httpx.Client,
    auth_user: dict,
) -> None:
    """
    Test /api/v1/auth/me endpoint security:
    - 401 Unauthorized without credentials
    - 401 Unauthorized with invalid token
    - 200 OK with Bearer header
    - 200 OK with session Cookie
    """
    # 1. Unauthenticated
    unauth_resp = api_client.get("/api/v1/auth/me")
    assert unauth_resp.status_code == 401
    assert unauth_resp.json().get("code") == "UNAUTHORIZED"

    # 2. Invalid Bearer token
    bad_token_resp = api_client.get(
        "/api/v1/auth/me",
        headers={"Authorization": "Bearer invalid.jwt.signature"},
    )
    assert bad_token_resp.status_code == 401
    assert bad_token_resp.json().get("code") == "UNAUTHORIZED"

    # 3. Authorized via Bearer header
    bearer_resp = auth_client.get("/api/v1/auth/me")
    assert bearer_resp.status_code == 200
    me_data = bearer_resp.json()
    assert me_data["user"]["id"] == auth_user["id"]
    assert me_data["user"]["email"] == auth_user["email"]
    assert me_data["workspace"]["name"] == "Личное"
    assert me_data["workspace"]["role"] == "owner"

    # 4. Authorized via HttpOnly Cookie
    cookie_resp = cookie_client.get("/api/v1/auth/me")
    assert cookie_resp.status_code == 200
    cookie_data = cookie_resp.json()
    assert cookie_data["user"]["id"] == auth_user["id"]
    assert cookie_data["workspace"]["id"] == auth_user["workspace"]["id"]


def test_logout_clears_cookie(api_client: httpx.Client) -> None:
    """
    Test POST /api/v1/auth/logout clears the session cookie.
    """
    resp = api_client.post("/api/v1/auth/logout")
    assert resp.status_code == 200
    assert resp.json().get("message") == "Logged out successfully"

    # Cookie clearing check
    set_cookie = resp.headers.get("set-cookie", "")
    assert "token=" in set_cookie
    assert "max-age=0" in set_cookie.lower() or "expires=" in set_cookie.lower()


def test_registration_validation_and_conflicts(api_client: httpx.Client) -> None:
    """
    Test edge cases and validation for registration:
    - 409 Conflict when user email already registered
    - 400 Bad Request when password is under 6 characters
    - 400 Bad Request on empty or missing fields
    """
    unique_suffix = uuid.uuid4().hex[:8]
    email = f"conflict_{unique_suffix}@example.com"

    # Initial registration
    first_resp = api_client.post(
        "/api/v1/auth/register",
        json={"email": email, "password": "valid_password", "displayName": "Original"},
    )
    assert first_resp.status_code == 201

    # Duplicate registration -> 409
    dup_resp = api_client.post(
        "/api/v1/auth/register",
        json={"email": email, "password": "valid_password", "displayName": "Duplicate"},
    )
    assert dup_resp.status_code == 409
    assert dup_resp.json().get("code") == "CONFLICT"

    # Password too short (< 6 chars) -> 400
    short_pwd_resp = api_client.post(
        "/api/v1/auth/register",
        json={"email": f"short_{unique_suffix}@example.com", "password": "123", "displayName": "Short"},
    )
    assert short_pwd_resp.status_code == 400
    assert short_pwd_resp.json().get("code") == "BAD_REQUEST"

    # Empty email -> 400
    empty_email_resp = api_client.post(
        "/api/v1/auth/register",
        json={"email": "   ", "password": "valid_password", "displayName": "User"},
    )
    assert empty_email_resp.status_code == 400
    assert empty_email_resp.json().get("code") == "BAD_REQUEST"
