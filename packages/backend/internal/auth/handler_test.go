package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"hydroponic-backend/internal/platform/config"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type listUsersResponse struct {
	Code int `json:"code"`
	Data struct {
		Items []struct {
			ID        uint64   `json:"id"`
			Username  string   `json:"username"`
			Nickname  string   `json:"nickname"`
			Phone     string   `json:"phone"`
			Email     string   `json:"email"`
			Status    string   `json:"status"`
			Roles     []string `json:"roles"`
			CreatedAt string   `json:"created_at"`
		} `json:"items"`
	} `json:"data"`
}

func TestListUsers_IncludesProfileFields(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&User{}, &Role{}, &UserRole{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := ensureAuditLogTable(db); err != nil {
		t.Fatalf("create audit_logs table: %v", err)
	}

	role := Role{Name: RoleOperator, Description: "operator"}
	if err := db.Create(&role).Error; err != nil {
		t.Fatalf("create role: %v", err)
	}

	user := User{
		Username:     "operator1",
		PasswordHash: "hashed",
		Nickname:     "操作员甲",
		Phone:        "13800138000",
		Email:        "operator1@example.com",
		Status:       UserStatusEnabled,
		CreatedAt:    time.Date(2026, 5, 12, 8, 30, 0, 0, time.UTC),
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	if err := db.Model(&user).Association("Roles").Replace(&role); err != nil {
		t.Fatalf("bind roles: %v", err)
	}

	h := NewHandler(config.AuthConfig{}, db)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/users?page=1&page_size=20", nil)
	c.Set("request_id", "req-list-users")

	h.ListUsers(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body=%s", w.Code, w.Body.String())
	}

	var resp listUsersResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp.Data.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(resp.Data.Items))
	}

	item := resp.Data.Items[0]
	if item.Nickname != user.Nickname {
		t.Fatalf("expected nickname %q, got %q", user.Nickname, item.Nickname)
	}
	if item.Phone != user.Phone {
		t.Fatalf("expected phone %q, got %q", user.Phone, item.Phone)
	}
	if item.Email != user.Email {
		t.Fatalf("expected email %q, got %q", user.Email, item.Email)
	}
	if item.CreatedAt == "" {
		t.Fatalf("expected created_at to be present")
	}
	if len(item.Roles) != 1 || item.Roles[0] != role.Name {
		t.Fatalf("expected roles [%q], got %v", role.Name, item.Roles)
	}
}

func TestCreateUser_PersistsPhoneAndEmail(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&User{}, &Role{}, &UserRole{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := ensureAuditLogTable(db); err != nil {
		t.Fatalf("create audit_logs table: %v", err)
	}

	role := Role{Name: RoleViewer, Description: "viewer"}
	if err := db.Create(&role).Error; err != nil {
		t.Fatalf("create role: %v", err)
	}

	h := NewHandler(config.AuthConfig{}, db)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	reqBody := `{"username":"viewer1","password":"pass1234","nickname":"访客","phone":"13900139000","email":"viewer1@example.com","roles":["VIEWER"]}`
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	c.Set("user_id", uint64(1))
	c.Set("request_id", "req-create-user")

	h.CreateUser(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body=%s", w.Code, w.Body.String())
	}

	var user User
	if err := db.Where("username = ?", "viewer1").First(&user).Error; err != nil {
		t.Fatalf("load user: %v", err)
	}
	if user.Phone != "13900139000" {
		t.Fatalf("expected phone to be saved, got %q", user.Phone)
	}
	if user.Email != "viewer1@example.com" {
		t.Fatalf("expected email to be saved, got %q", user.Email)
	}
}

func ensureAuditLogTable(db *gorm.DB) error {
	return db.Exec(`
		CREATE TABLE IF NOT EXISTS audit_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			action TEXT NOT NULL,
			target_type TEXT NOT NULL,
			target_id INTEGER NULL,
			detail JSON NULL,
			request_id TEXT NULL,
			before_data JSON NULL,
			after_data JSON NULL,
			created_at DATETIME NULL
		)
	`).Error
}
