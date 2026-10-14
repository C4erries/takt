"""E2E tests for Takt Inbox Tasks CRUD and Tenant Isolation (IDOR protection)."""

import uuid
import httpx
import pytest


def test_create_and_list_inbox_tasks(auth_client: httpx.Client, auth_user: dict) -> None:
    """
    Test creating a new task in Inbox and listing all tasks.
    Verifies 201 Created with correct position, fields, and 200 OK list.
    """
    create_payload = {
        "title": "Купить чай сенча",
        "descriptionMarkdown": "Заказать премиальный сорт из Киото",
        "priority": "high",
        "dueDate": "2026-10-25",
    }

    create_resp = auth_client.post("/api/v1/tasks/inbox", json=create_payload)
    assert create_resp.status_code == 201, f"Task creation failed: {create_resp.text}"

    task = create_resp.json()
    assert task["title"] == create_payload["title"]
    assert task["descriptionMarkdown"] == create_payload["descriptionMarkdown"]
    assert task["priority"] == "high"
    assert task["dueDate"] == "2026-10-25"
    assert task["isCompleted"] is False
    assert task["completedAt"] is None
    assert task["workspaceId"] == auth_user["workspace"]["id"]
    # 3 onboarding tasks are positions 0, 1, 2; next position must be 3
    assert task["position"] == 3

    # List inbox tasks
    list_resp = auth_client.get("/api/v1/tasks/inbox")
    assert list_resp.status_code == 200
    tasks = list_resp.json()
    assert len(tasks) == 4

    # Verify positions are monotonically ordered
    positions = [t["position"] for t in tasks]
    assert positions == sorted(positions)
    assert tasks[-1]["id"] == task["id"]


def test_toggle_task_completion_lifecycle(auth_client: httpx.Client) -> None:
    """
    Test toggling task completion state:
    1. Initially incomplete (isCompleted=False, completedAt=None)
    2. Toggle -> isCompleted=True, completedAt filled
    3. Toggle again -> isCompleted=False, completedAt=None
    """
    create_resp = auth_client.post(
        "/api/v1/tasks/inbox",
        json={"title": "Позвонить партнеру"},
    )
    assert create_resp.status_code == 201
    task_id = create_resp.json()["id"]

    # Toggle to completed
    toggle_1_resp = auth_client.patch(f"/api/v1/tasks/{task_id}/toggle")
    assert toggle_1_resp.status_code == 200
    toggled_1 = toggle_1_resp.json()
    assert toggled_1["isCompleted"] is True
    assert toggled_1["completedAt"] is not None

    # Toggle back to incomplete
    toggle_2_resp = auth_client.patch(f"/api/v1/tasks/{task_id}/toggle")
    assert toggle_2_resp.status_code == 200
    toggled_2 = toggle_2_resp.json()
    assert toggled_2["isCompleted"] is False
    assert toggled_2["completedAt"] is None


def test_delete_task(auth_client: httpx.Client) -> None:
    """
    Test task deletion:
    - DELETE /api/v1/tasks/{id} returns 204 No Content
    - Subsequent toggle or delete on the same ID returns 404 Not Found
    - Task disappears from GET /api/v1/tasks/inbox
    """
    create_resp = auth_client.post(
        "/api/v1/tasks/inbox",
        json={"title": "Временная задача для удаления"},
    )
    assert create_resp.status_code == 201
    task_id = create_resp.json()["id"]

    # Delete task
    del_resp = auth_client.delete(f"/api/v1/tasks/{task_id}")
    assert del_resp.status_code == 204

    # Repeat delete -> 404
    repeat_del = auth_client.delete(f"/api/v1/tasks/{task_id}")
    assert repeat_del.status_code == 404
    assert repeat_del.json().get("code") == "NOT_FOUND"

    # Toggle deleted task -> 404
    toggle_resp = auth_client.patch(f"/api/v1/tasks/{task_id}/toggle")
    assert toggle_resp.status_code == 404
    assert toggle_resp.json().get("code") == "NOT_FOUND"

    # Verify task is not in inbox
    inbox_resp = auth_client.get("/api/v1/tasks/inbox")
    assert inbox_resp.status_code == 200
    inbox_ids = [t["id"] for t in inbox_resp.json()]
    assert task_id not in inbox_ids


def test_tenant_isolation_idor_protection(
    auth_client: httpx.Client,
    other_client: httpx.Client,
    auth_user: dict,
    other_user: dict,
) -> None:
    """
    Critical Tenant Isolation Security Test:
    User A creates a task in Workspace A.
    User B (Workspace B) must NOT be able to:
    1. See User A's task in their inbox (data leakage)
    2. Toggle User A's task (IDOR modification -> 404)
    3. Delete User A's task (IDOR deletion -> 404)
    User A's task must remain untouched.
    """
    # 1. User A creates private task
    create_resp = auth_client.post(
        "/api/v1/tasks/inbox",
        json={"title": "Сверхсекретная стратегия User A"},
    )
    assert create_resp.status_code == 201
    user_a_task = create_resp.json()
    task_id = user_a_task["id"]

    # 2. User B fetches their inbox -> User A's task must NOT be visible
    user_b_inbox_resp = other_client.get("/api/v1/tasks/inbox")
    assert user_b_inbox_resp.status_code == 200
    user_b_task_ids = [t["id"] for t in user_b_inbox_resp.json()]
    assert task_id not in user_b_task_ids

    # 3. User B attempts to toggle User A's task -> 404 Not Found (IDOR protection)
    b_toggle_resp = other_client.patch(f"/api/v1/tasks/{task_id}/toggle")
    assert b_toggle_resp.status_code == 404, (
        f"IDOR vulnerability! Expected 404, got {b_toggle_resp.status_code}"
    )
    assert b_toggle_resp.json().get("code") == "NOT_FOUND"

    # 4. User B attempts to delete User A's task -> 404 Not Found (IDOR protection)
    b_del_resp = other_client.delete(f"/api/v1/tasks/{task_id}")
    assert b_del_resp.status_code == 404, (
        f"IDOR vulnerability! Expected 404, got {b_del_resp.status_code}"
    )
    assert b_del_resp.json().get("code") == "NOT_FOUND"

    # 5. User A verifies their task is untouched and still incomplete
    a_inbox_resp = auth_client.get("/api/v1/tasks/inbox")
    assert a_inbox_resp.status_code == 200
    matching = [t for t in a_inbox_resp.json() if t["id"] == task_id]
    assert len(matching) == 1
    assert matching[0]["isCompleted"] is False


def test_inbox_validation_and_unauthorized_access(
    api_client: httpx.Client,
    auth_client: httpx.Client,
) -> None:
    """
    Test validation, security, and error handling:
    - 401 Unauthorized for unauthenticated requests
    - 400 Bad Request on empty title or invalid dueDate
    - 400 Bad Request on invalid UUID format
    - 404 Not Found on nonexistent UUID
    """
    # 1. Unauthenticated requests
    assert api_client.get("/api/v1/tasks/inbox").status_code == 401
    assert api_client.post("/api/v1/tasks/inbox", json={"title": "Test"}).status_code == 401
    assert api_client.patch(f"/api/v1/tasks/{uuid.uuid4()}/toggle").status_code == 401
    assert api_client.delete(f"/api/v1/tasks/{uuid.uuid4()}").status_code == 401

    # 2. Validation: Empty title -> 400
    bad_title_resp = auth_client.post("/api/v1/tasks/inbox", json={"title": "   "})
    assert bad_title_resp.status_code == 400
    assert bad_title_resp.json().get("code") == "BAD_REQUEST"

    # 3. Validation: Malformed dueDate -> 400
    bad_date_resp = auth_client.post(
        "/api/v1/tasks/inbox",
        json={"title": "Valid title", "dueDate": "not-a-date"},
    )
    assert bad_date_resp.status_code == 400
    assert bad_date_resp.json().get("code") == "BAD_REQUEST"

    # 4. Invalid UUID format in path -> 400
    assert auth_client.patch("/api/v1/tasks/not-a-valid-uuid/toggle").status_code == 400
    assert auth_client.delete("/api/v1/tasks/not-a-valid-uuid").status_code == 400

    # 5. Nonexistent UUID -> 404
    random_uuid = str(uuid.uuid4())
    assert auth_client.patch(f"/api/v1/tasks/{random_uuid}/toggle").status_code == 404
    assert auth_client.delete(f"/api/v1/tasks/{random_uuid}").status_code == 404
