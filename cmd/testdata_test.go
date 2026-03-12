package cmd

// Canned JSON responses for tests, based on real Fizzy API shapes.

const boardListJSON = `[
  {
    "id": "board-1",
    "name": "Test Board",
    "all_access": true,
    "created_at": "2025-01-15T10:00:00Z",
    "url": "https://app.fizzy.do/test-account/boards/board-1",
    "creator": {"id": "user-1", "name": "Test User", "role": "admin", "active": true, "email_address": "test@example.com", "created_at": "2025-01-01T00:00:00Z", "url": ""}
  },
  {
    "id": "board-2",
    "name": "Another Board",
    "all_access": false,
    "created_at": "2025-02-20T14:30:00Z",
    "url": "https://app.fizzy.do/test-account/boards/board-2",
    "creator": {"id": "user-1", "name": "Test User", "role": "admin", "active": true, "email_address": "test@example.com", "created_at": "2025-01-01T00:00:00Z", "url": ""}
  }
]`

const boardListEmptyJSON = `[]`

const boardViewJSON = `{
  "id": "board-1",
  "name": "Test Board",
  "all_access": true,
  "created_at": "2025-01-15T10:00:00Z",
  "url": "https://app.fizzy.do/test-account/boards/board-1",
  "creator": {"id": "user-1", "name": "Test User", "role": "admin", "active": true, "email_address": "test@example.com", "created_at": "2025-01-01T00:00:00Z", "url": ""}
}`

const columnListJSON = `[
  {
    "id": "col-1",
    "name": "To Do",
    "color": {"name": "Blue", "value": "var(--color-card-default)"},
    "created_at": "2025-01-15T10:00:00Z"
  },
  {
    "id": "col-2",
    "name": "In Progress",
    "color": {"name": "Yellow", "value": "var(--color-card-3)"},
    "created_at": "2025-01-15T10:00:00Z"
  }
]`

const columnListEmptyJSON = `[]`

const cardListJSON = `[
  {
    "id": "card-aaa",
    "number": 1,
    "title": "First Card",
    "status": "triaged",
    "description": "",
    "tags": ["bug"],
    "closed": false,
    "golden": false,
    "last_active_at": "2025-03-01T12:00:00Z",
    "created_at": "2025-02-01T09:00:00Z",
    "url": "https://app.fizzy.do/test-account/cards/1",
    "board": {"id": "board-1", "name": "Test Board", "all_access": true, "created_at": "2025-01-15T10:00:00Z", "url": "", "creator": {"id": "user-1", "name": "Test User", "role": "admin", "active": true, "email_address": "", "created_at": "2025-01-01T00:00:00Z", "url": ""}},
    "column": {"id": "col-1", "name": "To Do", "color": {"name": "Blue", "value": ""}, "created_at": "2025-01-15T10:00:00Z"},
    "creator": {"id": "user-1", "name": "Test User", "role": "admin", "active": true, "email_address": "", "created_at": "2025-01-01T00:00:00Z", "url": ""},
    "comments_url": ""
  },
  {
    "id": "card-bbb",
    "number": 2,
    "title": "Second Card",
    "status": "new",
    "description": "A description",
    "tags": [],
    "closed": false,
    "golden": false,
    "last_active_at": "2025-03-02T12:00:00Z",
    "created_at": "2025-02-02T09:00:00Z",
    "url": "https://app.fizzy.do/test-account/cards/2",
    "board": {"id": "board-1", "name": "Test Board", "all_access": true, "created_at": "2025-01-15T10:00:00Z", "url": "", "creator": {"id": "user-1", "name": "Test User", "role": "admin", "active": true, "email_address": "", "created_at": "2025-01-01T00:00:00Z", "url": ""}},
    "creator": {"id": "user-1", "name": "Test User", "role": "admin", "active": true, "email_address": "", "created_at": "2025-01-01T00:00:00Z", "url": ""},
    "comments_url": ""
  }
]`

const cardListEmptyJSON = `[]`

const cardViewJSON = `{
  "id": "card-aaa",
  "number": 1,
  "title": "First Card",
  "status": "triaged",
  "description": "Card description here",
  "tags": ["bug", "urgent"],
  "closed": false,
  "golden": true,
  "last_active_at": "2025-03-01T12:00:00Z",
  "created_at": "2025-02-01T09:00:00Z",
  "url": "https://app.fizzy.do/test-account/cards/1",
  "board": {"id": "board-1", "name": "Test Board", "all_access": true, "created_at": "2025-01-15T10:00:00Z", "url": "", "creator": {"id": "user-1", "name": "Test User", "role": "admin", "active": true, "email_address": "", "created_at": "2025-01-01T00:00:00Z", "url": ""}},
  "column": {"id": "col-1", "name": "To Do", "color": {"name": "Blue", "value": ""}, "created_at": "2025-01-15T10:00:00Z"},
  "creator": {"id": "user-1", "name": "Test User", "role": "admin", "active": true, "email_address": "", "created_at": "2025-01-01T00:00:00Z", "url": ""},
  "comments_url": "",
  "steps": [
    {"id": "step-1", "content": "Step one", "completed": true},
    {"id": "step-2", "content": "Step two", "completed": false}
  ]
}`

const commentListJSON = `[
  {
    "id": "comment-1",
    "created_at": "2025-02-10T09:00:00Z",
    "updated_at": "2025-02-10T09:00:00Z",
    "body": {"plain_text": "This is a test comment", "html": "<p>This is a test comment</p>"},
    "creator": {"id": "user-1", "name": "Test User", "role": "admin", "active": true, "email_address": "", "created_at": "2025-01-01T00:00:00Z", "url": ""},
    "url": ""
  }
]`

const commentListEmptyJSON = `[]`

const apiGetJSON = `{"id":"board-1","name":"Test Board"}`

const pinListJSON = `[
  {
    "id": "card-aaa",
    "number": 1,
    "title": "Pinned Card",
    "status": "triaged",
    "description": "",
    "tags": [],
    "closed": false,
    "golden": false,
    "last_active_at": "2025-03-01T12:00:00Z",
    "created_at": "2025-02-01T09:00:00Z",
    "url": "https://app.fizzy.do/test-account/cards/1",
    "board": {"id": "board-1", "name": "Test Board", "all_access": true, "created_at": "2025-01-15T10:00:00Z", "url": "", "creator": {"id": "user-1", "name": "Test User", "role": "admin", "active": true, "email_address": "", "created_at": "2025-01-01T00:00:00Z", "url": ""}},
    "creator": {"id": "user-1", "name": "Test User", "role": "admin", "active": true, "email_address": "", "created_at": "2025-01-01T00:00:00Z", "url": ""},
    "comments_url": ""
  }
]`

const identityJSON = `{
  "accounts": [
    {
      "id": "acct-1",
      "name": "Test Org",
      "slug": "/test-account",
      "created_at": "2025-01-01T00:00:00Z",
      "user": {"id": "user-1", "name": "Test User", "role": "admin", "active": true, "email_address": "test@example.com", "created_at": "2025-01-01T00:00:00Z", "url": ""}
    }
  ]
}`

const notificationsEmptyJSON = `[]`

const notificationsReadJSON = `[
  {
    "id": "notif-1",
    "title": "Card updated",
    "body": "Already handled",
    "url": "https://app.fizzy.do/test-account/notifications/notif-1",
    "read": true,
    "read_at": "2025-03-02T10:00:00Z",
    "created_at": "2025-03-02T09:00:00Z",
    "card": {
      "id": "card-aaa",
      "title": "First Card",
      "status": "triaged",
      "url": "https://app.fizzy.do/test-account/cards/1"
    },
    "creator": {
      "id": "user-1",
      "name": "Test User",
      "role": "admin",
      "active": true,
      "email_address": "test@example.com",
      "created_at": "2025-01-01T00:00:00Z",
      "url": ""
    }
  }
]`
