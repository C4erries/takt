"""E2E tests for Unified Workspaces, Owner/Member roles, and Invite links (Sprint 0.4.1)."""

import uuid
import httpx
import pytest


def test_workspace_creation_and_switching(
    auth_client: httpx.Client, auth_user: dict, base_url: str
) -> None:
    """Verify creating a second workspace, automations seeding, and context switching."""
    default_ws = auth_user["workspace"]
    assert default_ws["name"] == "Личное"
    assert default_ws["role"] == "owner"

    # 1. Create second workspace
    create_resp = auth_client.post(
        "/api/v1/workspaces",
        json={"name": "Работа"},
    )
    assert create_resp.status_code == 201, f"Failed to create workspace: {create_resp.text}"
    work_ws = create_resp.json()
    assert work_ws["name"] == "Работа"
    assert work_ws["role"] == "owner"
    work_ws_id = work_ws["id"]

    # 2. Check list of workspaces
    list_resp = auth_client.get("/api/v1/workspaces")
    assert list_resp.status_code == 200
    workspaces = list_resp.json()
    assert len(workspaces) >= 2
    ws_map = {w["id"]: w for w in workspaces}
    assert default_ws["id"] in ws_map and ws_map[default_ws["id"]]["role"] == "owner"
    assert work_ws_id in ws_map and ws_map[work_ws_id]["role"] == "owner"

    # 3. Switch active workspace to second workspace
    switch_resp = auth_client.post(f"/api/v1/workspaces/{work_ws_id}/switch")
    assert switch_resp.status_code == 200
    switch_data = switch_resp.json()
    assert "token" in switch_data and len(switch_data["token"]) > 20
    assert switch_data["workspace"]["id"] == work_ws_id

    # 4. Verify automations seeded in the new workspace using switched client
    switched_headers = {"Authorization": f"Bearer {switch_data['token']}"}
    with httpx.Client(base_url=base_url, timeout=10.0, headers=switched_headers) as sw_client:
        auto_resp = sw_client.get(f"/api/v1/workspaces/{work_ws_id}/automations")
        assert auto_resp.status_code == 200
        automations = auto_resp.json()
        assert len(automations) == 3
        rule_names = {a["name"] for a in automations}
        assert {"Done & Archive", "Auto-Focus", "Inbox Zero"} <= rule_names

        # 5. Verify operations under switched token
        task_resp = sw_client.post("/api/v1/tasks/inbox", json={"title": "Задача в воркспейсе Работа"})
        assert task_resp.status_code == 201
        assert task_resp.json()["workspaceId"] == work_ws_id


def test_workspace_rename(auth_client: httpx.Client, auth_user: dict) -> None:
    """Verify owner can successfully rename a workspace via PATCH."""
    ws_id = auth_user["workspace"]["id"]

    patch_resp = auth_client.patch(
        f"/api/v1/workspaces/{ws_id}",
        json={"name": "Основное пространство"},
    )
    assert patch_resp.status_code == 200
    assert patch_resp.json()["name"] == "Основное пространство"

    # Verify persistence
    list_resp = auth_client.get("/api/v1/workspaces")
    assert list_resp.status_code == 200
    ws = next(w for w in list_resp.json() if w["id"] == ws_id)
    assert ws["name"] == "Основное пространство"


def test_invite_lifecycle_and_acceptance(
    auth_client: httpx.Client,
    auth_user: dict,
    user_factory,
    api_client: httpx.Client,
    base_url: str,
) -> None:
    """Verify invite link generation by owner, public preview, acceptance, and member listing."""
    ws_id = auth_user["workspace"]["id"]

    # 1. Owner generates invite
    inv_resp = auth_client.post(f"/api/v1/workspaces/{ws_id}/invites")
    assert inv_resp.status_code == 201, f"Failed to create invite: {inv_resp.text}"
    invite = inv_resp.json()
    assert "token" in invite and len(invite["token"]) > 10
    token = invite["token"]

    # 2. Public preview without auth
    preview_resp = api_client.get(f"/api/v1/invites/{token}")
    assert preview_resp.status_code == 200
    preview = preview_resp.json()
    assert preview["token"] == token
    assert preview["workspace_name"] == auth_user["workspace"]["name"]
    assert preview["invited_by_name"] == auth_user["displayName"]

    # 3. User B registers and accepts invite
    user_b = user_factory(display_name="Коллега Алексей")
    b_headers = {"Authorization": f"Bearer {user_b['token']}"}
    with httpx.Client(base_url=base_url, timeout=10.0, headers=b_headers) as b_client:
        accept_resp = b_client.post(f"/api/v1/invites/{token}/accept")
        assert accept_resp.status_code == 200, f"Failed to accept invite: {accept_resp.text}"
        accept_data = accept_resp.json()
        assert accept_data["workspace"]["id"] == ws_id
        assert accept_data["workspace"]["role"] == "member"

        # User B now sees this workspace in their list with role 'member'
        b_workspaces = b_client.get("/api/v1/workspaces").json()
        joined_ws = next(w for w in b_workspaces if w["id"] == ws_id)
        assert joined_ws["role"] == "member"

    # 4. Check members list
    members_resp = auth_client.get(f"/api/v1/workspaces/{ws_id}/members")
    assert members_resp.status_code == 200
    members = members_resp.json()
    assert len(members) == 2
    roles = {m["user_id"]: m["role"] for m in members}
    assert roles[auth_user["id"]] == "owner"
    assert roles[user_b["id"]] == "member"


def test_workspace_rbac_enforcement(
    auth_client: httpx.Client,
    auth_user: dict,
    user_factory,
    base_url: str,
) -> None:
    """Verify RBAC permissions for Owner vs Member."""
    ws_id = auth_user["workspace"]["id"]

    # Invite user B
    inv = auth_client.post(f"/api/v1/workspaces/{ws_id}/invites").json()
    user_b = user_factory(display_name="Участник Боб")
    b_headers = {"Authorization": f"Bearer {user_b['token']}"}

    with httpx.Client(base_url=base_url, timeout=10.0, headers=b_headers) as b_client:
        acc_resp = b_client.post(f"/api/v1/invites/{inv['token']}/accept")
        assert acc_resp.status_code == 200
        switched_token = acc_resp.json()["token"]

    # Use client with switched active workspace token for User B
    b_active_headers = {"Authorization": f"Bearer {switched_token}"}
    with httpx.Client(base_url=base_url, timeout=10.0, headers=b_active_headers) as b_active:
        # 1. Member tries to rename workspace -> 403 Forbidden
        rename_resp = b_active.patch(
            f"/api/v1/workspaces/{ws_id}",
            json={"name": "Хакерский воркспейс"},
        )
        assert rename_resp.status_code == 403

        # 2. Member tries to generate invite -> 403 Forbidden
        gen_inv_resp = b_active.post(f"/api/v1/workspaces/{ws_id}/invites")
        assert gen_inv_resp.status_code == 403

        # 3. Member tries to remove owner -> 400 or 403
        del_owner_resp = b_active.delete(f"/api/v1/workspaces/{ws_id}/members/{auth_user['id']}")
        assert del_owner_resp.status_code in (400, 403)

        # 4. Member successfully creates a task in this workspace
        create_task_resp = b_active.post(
            "/api/v1/tasks/inbox",
            json={"title": "Командная задача от Боба"},
        )
        assert create_task_resp.status_code == 201
        assert create_task_resp.json()["workspaceId"] == ws_id

        # 5. Member voluntarily leaves workspace
        leave_resp = b_active.delete(f"/api/v1/workspaces/{ws_id}/members/{user_b['id']}")
        assert leave_resp.status_code == 204

    # 6. Owner tries to remove self -> 400 Bad Request
    del_self_resp = auth_client.delete(f"/api/v1/workspaces/{ws_id}/members/{auth_user['id']}")
    assert del_self_resp.status_code == 400

    # Verify User B is no longer in members list
    members = auth_client.get(f"/api/v1/workspaces/{ws_id}/members").json()
    assert len(members) == 1
    assert members[0]["user_id"] == auth_user["id"]


def test_invite_already_member_conflict(
    auth_client: httpx.Client,
    auth_user: dict,
    user_factory,
    base_url: str,
) -> None:
    """Verify 400/409 when user already belongs to workspace and 404 for invalid invite."""
    ws_id = auth_user["workspace"]["id"]
    token = auth_client.post(f"/api/v1/workspaces/{ws_id}/invites").json()["token"]

    user_c = user_factory(display_name="Участник Чарли")
    c_headers = {"Authorization": f"Bearer {user_c['token']}"}
    with httpx.Client(base_url=base_url, timeout=10.0, headers=c_headers) as c_client:
        # First accept -> 200 OK
        resp1 = c_client.post(f"/api/v1/invites/{token}/accept")
        assert resp1.status_code == 200

        # Second accept -> 400 / 409
        resp2 = c_client.post(f"/api/v1/invites/{token}/accept")
        assert resp2.status_code in (400, 409)
        assert resp2.json().get("code") in ("ALREADY_MEMBER", "BAD_REQUEST", "CONFLICT")

        # Non-existent token -> 404
        fake_token = uuid.uuid4().hex
        resp_nf = c_client.post(f"/api/v1/invites/{fake_token}/accept")
        assert resp_nf.status_code == 404


def test_workspaces_security_and_unauthorized(
    api_client: httpx.Client,
    auth_client: httpx.Client,
    auth_user: dict,
    other_client: httpx.Client,
) -> None:
    """Verify 401 for unauthenticated requests and 403/404 when switching to another's workspace."""
    auth_ws = auth_user["workspace"]["id"]

    # 1. 401 Unauthorized without auth
    assert api_client.get("/api/v1/workspaces").status_code == 401
    assert api_client.post("/api/v1/workspaces", json={"name": "Тест"}).status_code == 401
    assert api_client.post(f"/api/v1/workspaces/{auth_ws}/switch").status_code == 401
    assert api_client.get(f"/api/v1/workspaces/{auth_ws}/members").status_code == 401

    # 2. 403 Forbidden: other tenant cannot switch to a workspace they are not member of
    unauthorized_switch = other_client.post(f"/api/v1/workspaces/{auth_ws}/switch")
    assert unauthorized_switch.status_code in (403, 404)
