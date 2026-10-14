# tests/e2e/test_projects_and_bento.py
from datetime import date, timedelta
import uuid
import httpx
import pytest


def test_project_and_columns_lifecycle(auth_client: httpx.Client, auth_user: dict) -> None:
    workspace_id = auth_user["workspace"]["id"]

    # 1. Create project
    proj_payload = {
        "name": "Проект Киото",
        "description": "Запуск персонального GTD трекера",
        "color": "matcha",
    }
    create_resp = auth_client.post(f"/api/v1/workspaces/{workspace_id}/projects", json=proj_payload)
    assert create_resp.status_code == 201
    project = create_resp.json()
    project_id = project["id"]
    assert project["name"] == proj_payload["name"]
    assert project["color"] == "matcha"

    # 2. Verify auto-generated default board and 3 GTD columns
    boards_resp = auth_client.get(f"/api/v1/workspaces/{workspace_id}/projects/{project_id}/boards")
    assert boards_resp.status_code == 200
    boards = boards_resp.json()
    assert len(boards) >= 1
    board_id = boards[0]["id"]

    columns_resp = auth_client.get(f"/api/v1/workspaces/{workspace_id}/boards/{board_id}/columns")
    assert columns_resp.status_code == 200
    columns = columns_resp.json()
    assert len(columns) == 3
    categories = [c["statusCategory"] for c in columns]
    assert "todo" in categories
    assert "in_progress" in categories
    assert "done" in categories

    todo_col = next(c for c in columns if c["statusCategory"] == "todo")
    in_prog_col = next(c for c in columns if c["statusCategory"] == "in_progress")

    # 3. Create a task and move it to project column (DND move)
    task_resp = auth_client.post("/api/v1/tasks/inbox", json={"title": "Карточка Канбана"})
    assert task_resp.status_code == 201
    task_id = task_resp.json()["id"]

    move_resp = auth_client.post(
        f"/api/v1/workspaces/{workspace_id}/tasks/{task_id}/move",
        json={"projectId": project_id, "columnId": todo_col["id"], "position": 1500.0},
    )
    assert move_resp.status_code == 200
    moved_task = move_resp.json()
    assert moved_task["projectId"] == project_id
    assert moved_task["columnId"] == todo_col["id"]
    assert moved_task["position"] == 1500.0

    # 4. Safety Guard 1: Deleting column with task fails with 409 Conflict
    del_col_resp = auth_client.delete(f"/api/v1/workspaces/{workspace_id}/boards/{board_id}/columns/{todo_col['id']}")
    assert del_col_resp.status_code == 409

    # Move task to another column
    auth_client.post(
        f"/api/v1/workspaces/{workspace_id}/tasks/{task_id}/move",
        json={"projectId": project_id, "columnId": in_prog_col["id"], "position": 2500.0},
    )

    # Now empty todo column can be deleted
    del_col_ok = auth_client.delete(f"/api/v1/workspaces/{workspace_id}/boards/{board_id}/columns/{todo_col['id']}")
    assert del_col_ok.status_code == 204

    # 5. Safety Guard 2: Deleting project unlinks task safely back to Inbox
    del_proj_resp = auth_client.delete(f"/api/v1/workspaces/{workspace_id}/projects/{project_id}")
    assert del_proj_resp.status_code == 204

    # Verify task is back in Inbox (projectId=None, columnId=None)
    inbox_resp = auth_client.get("/api/v1/tasks/inbox")
    assert inbox_resp.status_code == 200
    inbox_tasks = inbox_resp.json()
    unlinked_task = next(t for t in inbox_tasks if t["id"] == task_id)
    assert unlinked_task.get("projectId") is None
    assert unlinked_task.get("columnId") is None


def test_tags_and_inspector_updates(auth_client: httpx.Client, auth_user: dict) -> None:
    workspace_id = auth_user["workspace"]["id"]

    # 1. Create tags
    tag_resp = auth_client.post(
        f"/api/v1/workspaces/{workspace_id}/tags",
        json={"name": "@работа", "colorHex": "matcha"},
    )
    assert tag_resp.status_code == 201
    tag = tag_resp.json()
    tag_id = tag["id"]
    assert tag["name"] == "@работа"

    # List tags
    tags_list_resp = auth_client.get(f"/api/v1/workspaces/{workspace_id}/tags")
    assert tags_list_resp.status_code == 200
    assert any(t["id"] == tag_id for t in tags_list_resp.json())

    # 2. Create task and update fields via inspector (PATCH)
    task_resp = auth_client.post("/api/v1/tasks/inbox", json={"title": "Подготовить отчет"})
    assert task_resp.status_code == 201
    task_id = task_resp.json()["id"]

    tomorrow = (date.today() + timedelta(days=1)).isoformat()
    patch_payload = {
        "title": "Подготовить отчет за квартал (ред.)",
        "descriptionMarkdown": "### Заметки\n- Пункт 1\n- Пункт 2",
        "dueDate": tomorrow,
        "priority": "urgent",
    }
    patch_resp = auth_client.patch(
        f"/api/v1/workspaces/{workspace_id}/tasks/{task_id}",
        json=patch_payload,
    )
    assert patch_resp.status_code == 200
    updated_task = patch_resp.json()
    assert updated_task["title"] == patch_payload["title"]
    assert updated_task["descriptionMarkdown"] == patch_payload["descriptionMarkdown"]
    assert updated_task["dueDate"] == tomorrow
    assert updated_task["priority"] == "urgent"

    # 3. Add tag to task
    add_tag_resp = auth_client.post(f"/api/v1/workspaces/{workspace_id}/tasks/{task_id}/tags/{tag_id}")
    assert add_tag_resp.status_code in (200, 204)

    # 4. Remove tag from task
    rem_tag_resp = auth_client.delete(f"/api/v1/workspaces/{workspace_id}/tasks/{task_id}/tags/{tag_id}")
    assert rem_tag_resp.status_code == 204

    # 5. Delete tag
    del_tag_resp = auth_client.delete(f"/api/v1/workspaces/{workspace_id}/tags/{tag_id}")
    assert del_tag_resp.status_code == 204


def test_today_focus_hybrid_selection(auth_client: httpx.Client, auth_user: dict) -> None:
    workspace_id = auth_user["workspace"]["id"]
    today_str = date.today().isoformat()
    yesterday_str = (date.today() - timedelta(days=1)).isoformat()
    future_str = (date.today() + timedelta(days=7)).isoformat()

    # Task 1: Due today -> should appear in Today
    t1_resp = auth_client.post(
        "/api/v1/tasks/inbox",
        json={"title": "Задача на сегодня", "dueDate": today_str},
    )
    assert t1_resp.status_code == 201
    t1_id = t1_resp.json()["id"]

    # Task 2: Overdue (yesterday) -> should appear in Today
    t2_resp = auth_client.post(
        "/api/v1/tasks/inbox",
        json={"title": "Просроченная задача", "dueDate": yesterday_str},
    )
    assert t2_resp.status_code == 201
    t2_id = t2_resp.json()["id"]

    # Task 3: Future date -> NOT in today initially
    t3_resp = auth_client.post(
        "/api/v1/tasks/inbox",
        json={"title": "Будущая задача", "dueDate": future_str},
    )
    assert t3_resp.status_code == 201
    t3_id = t3_resp.json()["id"]

    # Check today list
    today_resp = auth_client.get(f"/api/v1/workspaces/{workspace_id}/tasks/today")
    assert today_resp.status_code == 200
    today_ids = [t["id"] for t in today_resp.json()]
    assert t1_id in today_ids
    assert t2_id in today_ids
    assert t3_id not in today_ids

    # Add future task explicitly to focus
    focus_resp = auth_client.post(
        f"/api/v1/workspaces/{workspace_id}/tasks/{t3_id}/focus",
        json={"focusOrder": 1},
    )
    assert focus_resp.status_code == 200

    # Now future task should be in Today due to focusOrder
    today_resp_2 = auth_client.get(f"/api/v1/workspaces/{workspace_id}/tasks/today")
    assert today_resp_2.status_code == 200
    today_ids_2 = [t["id"] for t in today_resp_2.json()]
    assert t3_id in today_ids_2
