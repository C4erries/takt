"""E2E tests for Event Engine and 3 Preset Automations v1."""

from datetime import date, timedelta
import time
import uuid
import httpx
import pytest


def wait_for_condition(predicate, timeout: float = 3.0, interval: float = 0.1):
    """Poll predicate until truthy value or timeout expires."""
    deadline = time.time() + timeout
    while time.time() < deadline:
        res = predicate()
        if res:
            return res
        time.sleep(interval)
    res = predicate()
    if res:
        return res
    raise TimeoutError("Condition was not met within timeout")


def test_workspace_seeded_with_three_automations(auth_client: httpx.Client, auth_user: dict) -> None:
    """Verify that newly created workspace is seeded with 3 default automation rules."""
    workspace_id = auth_user["workspace"]["id"]
    resp = auth_client.get(f"/api/v1/workspaces/{workspace_id}/automations")
    assert resp.status_code == 200, f"Failed to list automations: {resp.text}"

    automations = resp.json()
    assert len(automations) == 3, f"Expected 3 automations, got {len(automations)}"
    rules_by_name = {a["name"]: a for a in automations}

    # 1. Done & Archive preset (task_completed, is_enabled: True)
    done_rule = rules_by_name["Done & Archive"]
    assert done_rule["isEnabled"] is True
    assert done_rule["triggerType"] == "task_completed"
    assert done_rule["workspaceId"] == workspace_id
    assert done_rule["actionsJson"] == {"action": "move_to_done"}

    # 2. Auto-Focus preset (task_updated, is_enabled: True)
    focus_rule = rules_by_name["Auto-Focus"]
    assert focus_rule["isEnabled"] is True
    assert focus_rule["triggerType"] == "task_updated"
    assert focus_rule["workspaceId"] == workspace_id
    assert focus_rule["actionsJson"] == {"action": "set_focus"}

    # 3. Inbox Zero preset (task_updated, is_enabled: False)
    inbox_rule = rules_by_name["Inbox Zero"]
    assert inbox_rule["isEnabled"] is False
    assert inbox_rule["triggerType"] == "task_updated"
    assert inbox_rule["workspaceId"] == workspace_id
    assert inbox_rule["actionsJson"] == {"action": "move_to_first_column"}


def test_toggle_automation_state(auth_client: httpx.Client, auth_user: dict) -> None:
    """Verify toggling automation state: False -> True -> False via PATCH /toggle."""
    workspace_id = auth_user["workspace"]["id"]
    list_resp = auth_client.get(f"/api/v1/workspaces/{workspace_id}/automations")
    inbox_rule = next(a for a in list_resp.json() if a["name"] == "Inbox Zero")
    rule_id = inbox_rule["id"]

    # Toggle to True
    toggle_1 = auth_client.patch(
        f"/api/v1/workspaces/{workspace_id}/automations/{rule_id}/toggle",
        json={"is_enabled": True},
    )
    assert toggle_1.status_code == 200 and toggle_1.json()["isEnabled"] is True

    # Check persistence via GET
    list_1 = auth_client.get(f"/api/v1/workspaces/{workspace_id}/automations")
    assert {a["id"]: a for a in list_1.json()}[rule_id]["isEnabled"] is True

    # Toggle back to False
    toggle_2 = auth_client.patch(
        f"/api/v1/workspaces/{workspace_id}/automations/{rule_id}/toggle",
        json={"is_enabled": False},
    )
    assert toggle_2.status_code == 200 and toggle_2.json()["isEnabled"] is False

    # Check persistence via GET
    list_2 = auth_client.get(f"/api/v1/workspaces/{workspace_id}/automations")
    assert {a["id"]: a for a in list_2.json()}[rule_id]["isEnabled"] is False


def test_preset_done_and_archive(auth_client: httpx.Client, auth_user: dict) -> None:
    """Verify 'Done & Archive': completing a task moves it to project's done column."""
    workspace_id = auth_user["workspace"]["id"]

    # 1. Create project with boards and columns
    proj_resp = auth_client.post(
        f"/api/v1/workspaces/{workspace_id}/projects",
        json={"name": "Проект Done & Archive", "color": "matcha"},
    )
    project_id = proj_resp.json()["id"]
    boards = auth_client.get(f"/api/v1/workspaces/{workspace_id}/projects/{project_id}/boards").json()
    cols = auth_client.get(f"/api/v1/workspaces/{workspace_id}/boards/{boards[0]['id']}/columns").json()
    todo_col = next(c for c in cols if c["statusCategory"] == "todo")
    done_col = next(c for c in cols if c["statusCategory"] == "done")

    # 2. Create task and place in 'todo' column
    task_resp = auth_client.post("/api/v1/tasks/inbox", json={"title": "Задача на завершение"})
    task_id = task_resp.json()["id"]
    auth_client.post(
        f"/api/v1/workspaces/{workspace_id}/tasks/{task_id}/move",
        json={"projectId": project_id, "columnId": todo_col["id"], "position": 1000.0},
    )

    # 3. Toggle task to completed
    auth_client.patch(f"/api/v1/tasks/{task_id}/toggle")

    # 4. Verify async EventBus automation moves task to 'done' column
    def check_task_in_done():
        res = auth_client.get(f"/api/v1/workspaces/{workspace_id}/projects/{project_id}/tasks")
        if res.status_code != 200:
            return None
        matches = [t for t in res.json() if t["id"] == task_id]
        if matches and matches[0].get("columnId") == done_col["id"] and matches[0].get("isCompleted"):
            return matches[0]
        return None

    done_task = wait_for_condition(check_task_in_done, timeout=3.0)
    assert done_task["columnId"] == done_col["id"]
    assert done_task["isCompleted"] is True
    assert done_task["completedAt"] is not None


def test_preset_auto_focus(auth_client: httpx.Client, auth_user: dict) -> None:
    """Verify 'Auto-Focus': task with due_date <= today automatically gets focus_order = 1."""
    workspace_id = auth_user["workspace"]["id"]
    today_str = date.today().isoformat()
    future_str = (date.today() + timedelta(days=5)).isoformat()

    # 1. Create task with due date today -> Auto-Focus triggers
    task_resp = auth_client.post(
        "/api/v1/tasks/inbox",
        json={"title": "Срочная задача на сегодня", "dueDate": today_str},
    )
    task_id = task_resp.json()["id"]

    def check_task_focused():
        res = auth_client.get(f"/api/v1/workspaces/{workspace_id}/tasks/today")
        if res.status_code != 200:
            return None
        matches = [t for t in res.json() if t["id"] == task_id]
        return matches[0] if matches and matches[0].get("focusOrder") == 1 else None

    focused_task = wait_for_condition(check_task_focused, timeout=3.0)
    assert focused_task["id"] == task_id and focused_task["focusOrder"] == 1

    # 2. Create task with future due date -> Auto-Focus does NOT trigger
    fut_resp = auth_client.post(
        "/api/v1/tasks/inbox",
        json={"title": "Будущая задача", "dueDate": future_str},
    )
    time.sleep(0.3)
    today_list = auth_client.get(f"/api/v1/workspaces/{workspace_id}/tasks/today").json()
    assert fut_resp.json()["id"] not in [t["id"] for t in today_list]


def test_preset_inbox_zero(auth_client: httpx.Client, auth_user: dict) -> None:
    """Verify 'Inbox Zero': assigning project_id moves task from inbox into project's todo column."""
    workspace_id = auth_user["workspace"]["id"]

    # 1. Enable 'Inbox Zero'
    list_resp = auth_client.get(f"/api/v1/workspaces/{workspace_id}/automations")
    inbox_rule = next(a for a in list_resp.json() if a["name"] == "Inbox Zero")
    auth_client.patch(
        f"/api/v1/workspaces/{workspace_id}/automations/{inbox_rule['id']}/toggle",
        json={"is_enabled": True},
    )

    # 2. Create project
    proj_resp = auth_client.post(
        f"/api/v1/workspaces/{workspace_id}/projects",
        json={"name": "Проект Inbox Zero", "color": "zen"},
    )
    project_id = proj_resp.json()["id"]
    boards = auth_client.get(f"/api/v1/workspaces/{workspace_id}/projects/{project_id}/boards").json()
    cols = auth_client.get(f"/api/v1/workspaces/{workspace_id}/boards/{boards[0]['id']}/columns").json()
    todo_col = next(c for c in cols if c["statusCategory"] == "todo")

    # 3. Create task in Inbox and assign project_id via PATCH
    task_resp = auth_client.post("/api/v1/tasks/inbox", json={"title": "Входящая задача"})
    task_id = task_resp.json()["id"]
    auth_client.patch(f"/api/v1/workspaces/{workspace_id}/tasks/{task_id}", json={"projectId": project_id})

    # 4. Verify async EventBus moves task to 'todo' column
    def check_task_in_todo():
        res = auth_client.get(f"/api/v1/workspaces/{workspace_id}/projects/{project_id}/tasks")
        if res.status_code != 200:
            return None
        matches = [t for t in res.json() if t["id"] == task_id]
        return matches[0] if matches and matches[0].get("columnId") == todo_col["id"] else None

    moved_task = wait_for_condition(check_task_in_todo, timeout=3.0)
    assert moved_task["columnId"] == todo_col["id"]
    assert moved_task["projectId"] == project_id

    # 5. Verify task left Inbox
    inbox_tasks = auth_client.get("/api/v1/tasks/inbox").json()
    assert task_id not in [t["id"] for t in inbox_tasks]


def test_loop_guard_protection(auth_client: httpx.Client, auth_user: dict) -> None:
    """Verify loop guard: cascade updates do not block or hang EventBus / HTTP server."""
    workspace_id = auth_user["workspace"]["id"]
    list_resp = auth_client.get(f"/api/v1/workspaces/{workspace_id}/automations")
    inbox_rule = next(a for a in list_resp.json() if a["name"] == "Inbox Zero")
    auth_client.patch(
        f"/api/v1/workspaces/{workspace_id}/automations/{inbox_rule['id']}/toggle",
        json={"is_enabled": True},
    )

    proj_resp = auth_client.post(
        f"/api/v1/workspaces/{workspace_id}/projects",
        json={"name": "Loop Guard Project", "color": "sumi"},
    )
    project_id = proj_resp.json()["id"]

    task_resp = auth_client.post(
        "/api/v1/tasks/inbox",
        json={"title": "Loop candidate", "dueDate": date.today().isoformat()},
    )
    task_id = task_resp.json()["id"]

    auth_client.patch(f"/api/v1/workspaces/{workspace_id}/tasks/{task_id}", json={"projectId": project_id})
    auth_client.patch(f"/api/v1/tasks/{task_id}/toggle")

    # Confirm immediate HTTP responsiveness
    t_start = time.perf_counter()
    health_resp = auth_client.get("/health")
    latency = time.perf_counter() - t_start
    assert health_resp.status_code == 200 and latency < 0.5


def test_automation_security_idor_and_validation(
    auth_client: httpx.Client, auth_user: dict, other_client: httpx.Client, other_user: dict
) -> None:
    """Negative tests: IDOR tenant protection (403/404) and validation errors (400/404)."""
    auth_ws = auth_user["workspace"]["id"]
    other_ws = other_user["workspace"]["id"]
    auth_rule_id = auth_client.get(f"/api/v1/workspaces/{auth_ws}/automations").json()[0]["id"]

    # 1. IDOR: other_user cannot toggle auth_user's automation in auth_user's workspace -> 403
    idor_resp = other_client.patch(
        f"/api/v1/workspaces/{auth_ws}/automations/{auth_rule_id}/toggle",
        json={"is_enabled": False},
    )
    assert idor_resp.status_code in (403, 404)

    # 2. IDOR: other_user cannot toggle auth_user's rule using their own workspace_id -> 404
    idor_cross = other_client.patch(
        f"/api/v1/workspaces/{other_ws}/automations/{auth_rule_id}/toggle",
        json={"is_enabled": False},
    )
    assert idor_cross.status_code == 404

    # 3. Validation: 400 Bad Request on malformed JSON payload
    bad_resp = auth_client.patch(
        f"/api/v1/workspaces/{auth_ws}/automations/{auth_rule_id}/toggle",
        content=b"not-json",
        headers={"Content-Type": "application/json"},
    )
    assert bad_resp.status_code == 400

    # 4. 404 Not Found on nonexistent automation ID
    random_id = str(uuid.uuid4())
    nf_resp = auth_client.patch(
        f"/api/v1/workspaces/{auth_ws}/automations/{random_id}/toggle",
        json={"is_enabled": True},
    )
    assert nf_resp.status_code == 404


def test_automations_unauthorized_access(api_client: httpx.Client, auth_user: dict) -> None:
    """Verify 401 Unauthorized when requesting automations without authentication."""
    workspace_id = auth_user["workspace"]["id"]
    random_id = str(uuid.uuid4())

    get_resp = api_client.get(f"/api/v1/workspaces/{workspace_id}/automations")
    assert get_resp.status_code == 401

    patch_resp = api_client.patch(
        f"/api/v1/workspaces/{workspace_id}/automations/{random_id}/toggle",
        json={"is_enabled": True},
    )
    assert patch_resp.status_code == 401
