"""E2E tests for Workspace Search API (tasks and projects fast search) and Tenant Isolation."""

import httpx
import pytest


def test_search_tasks_and_projects(auth_client: httpx.Client, auth_user: dict) -> None:
    """
    Test searching tasks and projects within a workspace:
    - Matches task by title and description
    - Matches project by name and description
    - Case-insensitive search (ILIKE)
    """
    workspace_id = auth_user["workspace"]["id"]

    # 1. Create a project with unique name and description
    proj_payload = {
        "name": "Проект Архитектура",
        "description": "Разработка микросервисов и структуры пакетов",
        "color": "matcha",
    }
    proj_resp = auth_client.post(
        f"/api/v1/workspaces/{workspace_id}/projects",
        json=proj_payload,
    )
    assert proj_resp.status_code == 201, f"Failed to create project: {proj_resp.text}"
    project = proj_resp.json()
    project_id = project["id"]

    # 2. Create a task with unique title and description
    task_payload = {
        "title": "Настроить поиск",
        "descriptionMarkdown": "Детальный поиск по подстроке и ключевым словам",
    }
    task_resp = auth_client.post("/api/v1/tasks/inbox", json=task_payload)
    assert task_resp.status_code == 201, f"Failed to create task: {task_resp.text}"
    task = task_resp.json()
    task_id = task["id"]

    # 3. Search for task by title substring ("поиск")
    search_task_resp = auth_client.get(
        f"/api/v1/workspaces/{workspace_id}/search",
        params={"q": "поиск"},
    )
    assert search_task_resp.status_code == 200, f"Search failed: {search_task_resp.text}"
    data = search_task_resp.json()
    assert "tasks" in data
    assert "projects" in data

    found_task_ids = [t["id"] for t in data["tasks"]]
    assert task_id in found_task_ids

    # 4. Search for project by name ("Архитектура")
    search_proj_resp = auth_client.get(
        f"/api/v1/workspaces/{workspace_id}/search",
        params={"q": "Архитектура"},
    )
    assert search_proj_resp.status_code == 200
    data_proj = search_proj_resp.json()
    found_proj_ids = [p["id"] for p in data_proj["projects"]]
    assert project_id in found_proj_ids

    # 5. Verify case-insensitivity: uppercase query ("ПОИСК") finds task
    upper_resp = auth_client.get(
        f"/api/v1/workspaces/{workspace_id}/search",
        params={"q": "ПОИСК"},
    )
    assert upper_resp.status_code == 200
    assert task_id in [t["id"] for t in upper_resp.json()["tasks"]]

    # 6. Verify case-insensitivity: lowercase query ("архитектура") finds project
    lower_resp = auth_client.get(
        f"/api/v1/workspaces/{workspace_id}/search",
        params={"q": "архитектура"},
    )
    assert lower_resp.status_code == 200
    assert project_id in [p["id"] for p in lower_resp.json()["projects"]]

    # 7. Search by task description substring ("подстроке")
    desc_resp = auth_client.get(
        f"/api/v1/workspaces/{workspace_id}/search",
        params={"q": "подстроке"},
    )
    assert desc_resp.status_code == 200
    assert task_id in [t["id"] for t in desc_resp.json()["tasks"]]


def test_search_tenant_isolation(
    auth_client: httpx.Client,
    auth_user: dict,
    other_client: httpx.Client,
    other_user: dict,
) -> None:
    """
    Test Tenant Isolation for workspace search:
    - User 1 creates task with secret word in workspace 1
    - User 2 searches in workspace 2 -> task must not be visible
    - User 2 attempts to search in workspace 1 -> 403 Forbidden
    """
    user1_ws = auth_user["workspace"]["id"]
    user2_ws = other_user["workspace"]["id"]

    # User 1 creates a task with secret title
    secret_title = "Конфиденциально Секрет123"
    task_resp = auth_client.post("/api/v1/tasks/inbox", json={"title": secret_title})
    assert task_resp.status_code == 201
    task_id = task_resp.json()["id"]

    # User 2 searches in their own workspace (workspace 2)
    user2_search_resp = other_client.get(
        f"/api/v1/workspaces/{user2_ws}/search",
        params={"q": "Секрет123"},
    )
    assert user2_search_resp.status_code == 200
    data = user2_search_resp.json()
    assert task_id not in [t["id"] for t in data["tasks"]]
    assert len(data["tasks"]) == 0

    # User 2 attempts to search in User 1's workspace (workspace 1) -> 403 Forbidden
    forbidden_resp = other_client.get(
        f"/api/v1/workspaces/{user1_ws}/search",
        params={"q": "Секрет123"},
    )
    assert forbidden_resp.status_code == 403, (
        f"Expected 403 Forbidden for cross-workspace search, got {forbidden_resp.status_code}"
    )


def test_search_validation_and_unauthorized(
    auth_client: httpx.Client,
    auth_user: dict,
    api_client: httpx.Client,
) -> None:
    """
    Test search validation errors and unauthorized access:
    - Missing 'q' query parameter -> 400 Bad Request
    - Whitespace-only 'q' query parameter -> 200 OK with empty tasks and projects
    - Unauthenticated request (no JWT token) -> 401 Unauthorized
    """
    workspace_id = auth_user["workspace"]["id"]

    # 1. Missing 'q' query parameter -> 400 Bad Request
    missing_q_resp = auth_client.get(f"/api/v1/workspaces/{workspace_id}/search")
    assert missing_q_resp.status_code == 400, (
        f"Expected 400 for missing 'q', got {missing_q_resp.status_code}"
    )

    # 2. Whitespace-only 'q' parameter -> 200 OK with empty arrays
    blank_q_resp = auth_client.get(
        f"/api/v1/workspaces/{workspace_id}/search",
        params={"q": "   "},
    )
    assert blank_q_resp.status_code == 200, (
        f"Expected 200 for whitespace 'q', got {blank_q_resp.status_code}"
    )
    data = blank_q_resp.json()
    assert data["tasks"] == []
    assert data["projects"] == []

    # 3. Unauthenticated request without Authorization header -> 401 Unauthorized
    unauth_resp = api_client.get(
        f"/api/v1/workspaces/{workspace_id}/search",
        params={"q": "test"},
    )
    assert unauth_resp.status_code == 401, (
        f"Expected 401 for unauthenticated search, got {unauth_resp.status_code}"
    )
