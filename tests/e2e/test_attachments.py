"""E2E tests for task attachments lifecycle, S3 storage, and multi-tenant isolation."""

import uuid
import httpx
import pytest

# Minimal 1x1 valid PNG image bytes for binary upload testing
TINY_PNG_BYTES = (
    b"\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR\x00\x00\x00\x01\x00\x00\x00\x01"
    b"\x08\x06\x00\x00\x00\x1f\x15c4\x00\x00\x00\nIDATx\x9cc\x00\x01"
    b"\x00\x00\x05\x00\x01\r\n-\xb4\x00\x00\x00\x00IEND\xaeB`\x82"
)


def _create_test_task(client: httpx.Client, title: str = "Task for attachments") -> dict:
    """Helper to create a task in the inbox and return its JSON payload."""
    resp = client.post("/api/v1/tasks/inbox", json={"title": title})
    assert resp.status_code == 201, f"Failed to create task: {resp.text}"
    return resp.json()


def test_attachment_upload_and_download(auth_client: httpx.Client, auth_user: dict) -> None:
    """
    Test uploading an attachment to a task via multipart/form-data,
    listing task attachments, and retrieving binary content inline and as attachment.
    """
    workspace_id = auth_user["workspace"]["id"]
    task = _create_test_task(auth_client, "Task with screenshot")
    task_id = task["id"]

    # 1. Upload file via multipart/form-data
    files = {"file": ("test_image.png", TINY_PNG_BYTES, "image/png")}
    upload_url = f"/api/v1/workspaces/{workspace_id}/tasks/{task_id}/attachments"
    upload_resp = auth_client.post(upload_url, files=files)

    assert upload_resp.status_code == 201, f"Upload failed ({upload_resp.status_code}): {upload_resp.text}"
    attachment = upload_resp.json()

    # Verify response schema
    assert "id" in attachment
    assert attachment["workspaceId"] == workspace_id
    assert attachment["taskId"] == task_id
    assert attachment["fileName"] == "test_image.png"
    assert attachment["fileSize"] == len(TINY_PNG_BYTES)
    assert attachment["contentType"] == "image/png"
    expected_content_url = f"/api/v1/workspaces/{workspace_id}/attachments/{attachment['id']}/content"
    assert attachment["contentUrl"] == expected_content_url
    assert "createdAt" in attachment

    attachment_id = attachment["id"]

    # 2. Verify attachment appears in task attachments list
    list_url = f"/api/v1/workspaces/{workspace_id}/tasks/{task_id}/attachments"
    list_resp = auth_client.get(list_url)
    assert list_resp.status_code == 200, f"List attachments failed: {list_resp.text}"
    items = list_resp.json()
    assert isinstance(items, list)
    assert len(items) == 1
    assert items[0]["id"] == attachment_id
    assert items[0]["fileName"] == "test_image.png"

    # 3. Retrieve binary content (default inline disposition)
    content_url = f"/api/v1/workspaces/{workspace_id}/attachments/{attachment_id}/content"
    content_resp = auth_client.get(content_url)
    assert content_resp.status_code == 200, f"Get content failed: {content_resp.text}"
    assert content_resp.content == TINY_PNG_BYTES
    assert "image/png" in content_resp.headers.get("content-type", "")
    assert "inline" in content_resp.headers.get("content-disposition", "")
    assert "test_image.png" in content_resp.headers.get("content-disposition", "")
    assert "max-age" in content_resp.headers.get("cache-control", "")

    # 4. Retrieve binary content with download=true disposition
    download_resp = auth_client.get(f"{content_url}?download=true")
    assert download_resp.status_code == 200
    assert download_resp.content == TINY_PNG_BYTES
    assert "attachment" in download_resp.headers.get("content-disposition", "")
    assert "test_image.png" in download_resp.headers.get("content-disposition", "")


def test_attachment_deletion(auth_client: httpx.Client, auth_user: dict) -> None:
    """
    Test single attachment deletion:
    - Upload an attachment
    - Delete via DELETE /workspaces/{wid}/attachments/{aid} -> 204
    - Verify content returns 404 Not Found
    - Verify task attachments list is empty
    """
    workspace_id = auth_user["workspace"]["id"]
    task = _create_test_task(auth_client, "Task for deletion test")
    task_id = task["id"]

    # Upload attachment
    files = {"file": ("notes.txt", b"Hello Attachment World!", "text/plain")}
    upload_resp = auth_client.post(
        f"/api/v1/workspaces/{workspace_id}/tasks/{task_id}/attachments",
        files=files,
    )
    assert upload_resp.status_code == 201
    attachment_id = upload_resp.json()["id"]

    # Delete attachment
    delete_url = f"/api/v1/workspaces/{workspace_id}/attachments/{attachment_id}"
    del_resp = auth_client.delete(delete_url)
    assert del_resp.status_code == 204, f"Delete failed: {del_resp.text}"

    # Verify content retrieval returns 404
    content_url = f"/api/v1/workspaces/{workspace_id}/attachments/{attachment_id}/content"
    content_resp = auth_client.get(content_url)
    assert content_resp.status_code == 404, f"Expected 404, got {content_resp.status_code}"

    # Verify list of attachments is now empty
    list_resp = auth_client.get(f"/api/v1/workspaces/{workspace_id}/tasks/{task_id}/attachments")
    assert list_resp.status_code == 200
    assert list_resp.json() == []


def test_attachment_cascade_on_task_delete(auth_client: httpx.Client, auth_user: dict) -> None:
    """
    Test cascade deletion:
    - Create task and upload an attachment
    - Delete task via DELETE /tasks/{task_id}
    - Verify attachment content endpoint returns 404 Not Found
    """
    workspace_id = auth_user["workspace"]["id"]
    task = _create_test_task(auth_client, "Task to be deleted with attachments")
    task_id = task["id"]

    # Upload attachment
    files = {"file": ("data.csv", b"id,val\n1,test\n2,pass", "text/csv")}
    upload_resp = auth_client.post(
        f"/api/v1/workspaces/{workspace_id}/tasks/{task_id}/attachments",
        files=files,
    )
    assert upload_resp.status_code == 201
    attachment_id = upload_resp.json()["id"]

    # Delete the parent task
    del_task_resp = auth_client.delete(f"/api/v1/tasks/{task_id}")
    assert del_task_resp.status_code == 204

    # Verify attachment content is no longer accessible (404)
    content_url = f"/api/v1/workspaces/{workspace_id}/attachments/{attachment_id}/content"
    content_resp = auth_client.get(content_url)
    assert content_resp.status_code == 404

    # Verify list attachments for deleted task returns 404
    list_resp = auth_client.get(f"/api/v1/workspaces/{workspace_id}/tasks/{task_id}/attachments")
    assert list_resp.status_code == 404


def test_attachment_tenant_isolation_security(
    auth_client: httpx.Client,
    auth_user: dict,
    other_client: httpx.Client,
    other_user: dict,
    api_client: httpx.Client,
) -> None:
    """
    Test multi-tenant isolation and authentication security:
    - User 1 uploads an attachment to their workspace
    - User 2 (different workspace) must get 403 or 404 when attempting to:
      * Get attachment content
      * Delete attachment
      * Upload an attachment to User 1's task
      * List User 1's task attachments
    - Unauthenticated client must get 401 Unauthorized on all attachment endpoints
    """
    user1_ws = auth_user["workspace"]["id"]
    user2_ws = other_user["workspace"]["id"]

    # User 1 creates task and uploads file
    task = _create_test_task(auth_client, "User 1 Private Task")
    task_id = task["id"]
    files = {"file": ("secret.pdf", b"%PDF-1.4 mock content", "application/pdf")}
    upload_resp = auth_client.post(
        f"/api/v1/workspaces/{user1_ws}/tasks/{task_id}/attachments",
        files=files,
    )
    assert upload_resp.status_code == 201
    att_id = upload_resp.json()["id"]

    # 1. User 2 attempts to get User 1's attachment content (both with user1_ws and user2_ws in path)
    resp = other_client.get(f"/api/v1/workspaces/{user1_ws}/attachments/{att_id}/content")
    assert resp.status_code in (403, 404), f"Expected 403/404, got {resp.status_code}"

    resp = other_client.get(f"/api/v1/workspaces/{user2_ws}/attachments/{att_id}/content")
    assert resp.status_code in (403, 404), f"Expected 403/404, got {resp.status_code}"

    # 2. User 2 attempts to delete User 1's attachment
    resp = other_client.delete(f"/api/v1/workspaces/{user1_ws}/attachments/{att_id}")
    assert resp.status_code in (403, 404), f"Expected 403/404, got {resp.status_code}"

    resp = other_client.delete(f"/api/v1/workspaces/{user2_ws}/attachments/{att_id}")
    assert resp.status_code in (403, 404), f"Expected 403/404, got {resp.status_code}"

    # 3. User 2 attempts to upload to User 1's task
    attacker_files = {"file": ("hack.exe", b"malicious", "application/octet-stream")}
    resp = other_client.post(
        f"/api/v1/workspaces/{user1_ws}/tasks/{task_id}/attachments",
        files=attacker_files,
    )
    assert resp.status_code in (403, 404), f"Expected 403/404, got {resp.status_code}"

    # 4. User 2 attempts to list User 1's task attachments
    resp = other_client.get(f"/api/v1/workspaces/{user1_ws}/tasks/{task_id}/attachments")
    assert resp.status_code in (403, 404), f"Expected 403/404, got {resp.status_code}"

    # 5. Unauthenticated client (no Authorization token)
    content_url = f"/api/v1/workspaces/{user1_ws}/attachments/{att_id}/content"
    list_url = f"/api/v1/workspaces/{user1_ws}/tasks/{task_id}/attachments"
    del_url = f"/api/v1/workspaces/{user1_ws}/attachments/{att_id}"
    up_url = f"/api/v1/workspaces/{user1_ws}/tasks/{task_id}/attachments"

    assert api_client.get(content_url).status_code == 401
    assert api_client.get(list_url).status_code == 401
    assert api_client.delete(del_url).status_code == 401
    assert api_client.post(up_url, files={"file": ("unauth.txt", b"x", "text/plain")}).status_code == 401


def test_attachment_validation_errors(auth_client: httpx.Client, auth_user: dict) -> None:
    """
    Test edge cases and validation error responses (400 Bad Request / 404 Not Found):
    - Upload request without 'file' multipart field -> 400 Bad Request
    - Upload empty file (0 bytes) -> 400 Bad Request
    - Upload to non-existent task UUID -> 404 Not Found
    - Retrieve non-existent attachment UUID -> 404 Not Found
    - Delete non-existent attachment UUID -> 404 Not Found
    """
    workspace_id = auth_user["workspace"]["id"]
    task = _create_test_task(auth_client, "Task for validation edge cases")
    task_id = task["id"]

    # 1. Missing 'file' multipart field (sending form data with wrong field name)
    resp = auth_client.post(
        f"/api/v1/workspaces/{workspace_id}/tasks/{task_id}/attachments",
        files={"wrong_field": ("test.png", TINY_PNG_BYTES, "image/png")},
    )
    assert resp.status_code == 400, f"Expected 400 for missing 'file' field, got {resp.status_code}"
    err = resp.json()
    assert "code" in err or "message" in err

    # 2. Empty file (0 bytes content)
    resp = auth_client.post(
        f"/api/v1/workspaces/{workspace_id}/tasks/{task_id}/attachments",
        files={"file": ("empty.txt", b"", "text/plain")},
    )
    assert resp.status_code == 400, f"Expected 400 for 0-byte file, got {resp.status_code}"

    # 3. Upload to non-existent task UUID
    random_task_id = str(uuid.uuid4())
    resp = auth_client.post(
        f"/api/v1/workspaces/{workspace_id}/tasks/{random_task_id}/attachments",
        files={"file": ("test.txt", b"content", "text/plain")},
    )
    assert resp.status_code == 404, f"Expected 404 for random task, got {resp.status_code}"

    # 4. Get content of non-existent attachment UUID
    random_att_id = str(uuid.uuid4())
    resp = auth_client.get(
        f"/api/v1/workspaces/{workspace_id}/attachments/{random_att_id}/content"
    )
    assert resp.status_code == 404

    # 5. Delete non-existent attachment UUID
    resp = auth_client.delete(
        f"/api/v1/workspaces/{workspace_id}/attachments/{random_att_id}"
    )
    assert resp.status_code == 404
