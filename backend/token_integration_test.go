package main_test
package main_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"task-manager/backend/internal/handlers"
	"task-manager/backend/internal/services"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupTestRouter(db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	authService := services.NewAuthService()
	registerService := services.NewRegisterService()

	authHandler := handlers.NewAuthHandler(db, authService)
	refreshHandler := handlers.NewRefreshHandler(db, authService)
	registerHandler := handlers.NewRegisterHandler(db, registerService)

	v1 := r.Group("/api/v1/auth")
	v1.POST("/register", registerHandler.Registration)
	v1.POST("/login", authHandler.Token)
	v1.POST("/refresh", refreshHandler.Refresh)

	return r
}

func newTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn:                 sqlDB,
		PreferSimpleProtocol: true,
	}), &gorm.Config{})
	require.NoError(t, err)

	return gormDB, mock
}

// TestRefreshTokenLifecycle tests the full token rotation flow:
//  1. Register a user
//  2. Login to obtain an access + refresh token
//  3. Exchange the refresh token → get a new pair
//  4. Use the OLD refresh token again → must be rejected (401)
//  5. Use the NEW refresh token → must succeed (200)
func TestRefreshTokenLifecycle(t *testing.T) {
	t.Setenv("JWT_SECRET", "integration-test-secret")

	gormDB, mock := newTestDB(t)
	router := setupTestRouter(gormDB)

	userID := uuid.Must(uuid.NewV7())
	hashedPw, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	require.NoError(t, err)

	now := time.Now()

	userCols := []string{"id", "username", "email", "password", "created_at", "updated_at", "deleted_at"}
	tokenCols := []string{"id", "user_id", "refresh_token", "expires_at", "created_at", "updated_at", "deleted_at"}

	userRow := func() *sqlmock.Rows {
		return sqlmock.NewRows(userCols).
			AddRow(userID, "testuser", "test@example.com", string(hashedPw), now, now, nil)
	}

	// ── 1. Register ───────────────────────────────────────────────────────────
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "users"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(userID))
	mock.ExpectCommit()

	w := httptest.NewRecorder()
	body := mustMarshal(t, map[string]string{
		"username": "testuser",
		"email":    "test@example.com",
		"password": "password123",
	})
	req := newJSONRequest(t, http.MethodPost, "/api/v1/auth/register", body)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code, "register should return 201")

	// ── 2. Login → capture first refresh token ────────────────────────────────
	mock.ExpectQuery(`SELECT .* FROM "users" WHERE username`).WillReturnRows(userRow())
	mock.ExpectQuery(`SELECT .* FROM "users" WHERE id`).WillReturnRows(userRow())
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "tokens"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.Must(uuid.NewV7())))
	mock.ExpectCommit()

	w = httptest.NewRecorder()
	body = mustMarshal(t, map[string]string{"username": "testuser", "password": "password123"})
	req = newJSONRequest(t, http.MethodPost, "/api/v1/auth/login", body)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "login should return 200")

	var loginResp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &loginResp))
	oldRefreshToken, ok := loginResp["refresh_token"].(string)
	require.True(t, ok, "login response must contain refresh_token")

	// ── 3. Exchange old refresh token → get new pair ──────────────────────────
	oldRefreshUUID := mustParseUUID(t, oldRefreshToken)
	firstTokenID := uuid.Must(uuid.NewV7())

	mock.ExpectQuery(`SELECT .* FROM "tokens" WHERE refresh_token`).
		WillReturnRows(sqlmock.NewRows(tokenCols).
			AddRow(firstTokenID, userID, oldRefreshUUID, now.Add(24*time.Hour), now, now, nil))
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "tokens" SET`).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	mock.ExpectQuery(`SELECT .* FROM "users" WHERE id`).WillReturnRows(userRow())
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "tokens"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.Must(uuid.NewV7())))
	mock.ExpectCommit()

	w = httptest.NewRecorder()
	body = mustMarshal(t, map[string]string{"refresh_token": oldRefreshToken})
	req = newJSONRequest(t, http.MethodPost, "/api/v1/auth/refresh", body)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "refresh with valid token should return 200")

	var refreshResp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &refreshResp))
	newRefreshToken, ok := refreshResp["refresh_token"].(string)
	require.True(t, ok, "refresh response must contain refresh_token")

	// ── 4. Reuse old refresh token → must be rejected ─────────────────────────
	// The old token was soft-deleted; the DB returns no rows.
	mock.ExpectQuery(`SELECT .* FROM "tokens" WHERE refresh_token`).
		WillReturnRows(sqlmock.NewRows(tokenCols)) // empty result set → ErrRecordNotFound

	w = httptest.NewRecorder()
	body = mustMarshal(t, map[string]string{"refresh_token": oldRefreshToken})
	req = newJSONRequest(t, http.MethodPost, "/api/v1/auth/refresh", body)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code, "old refresh token must be rejected with 401")

	// ── 5. Use new refresh token → must succeed ───────────────────────────────
	newRefreshUUID := mustParseUUID(t, newRefreshToken)
	secondTokenID := uuid.Must(uuid.NewV7())

	mock.ExpectQuery(`SELECT .* FROM "tokens" WHERE refresh_token`).
		WillReturnRows(sqlmock.NewRows(tokenCols).
			AddRow(secondTokenID, userID, newRefreshUUID, now.Add(24*time.Hour), now, now, nil))
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "tokens" SET`).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	mock.ExpectQuery(`SELECT .* FROM "users" WHERE id`).WillReturnRows(userRow())
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "tokens"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.Must(uuid.NewV7())))
	mock.ExpectCommit()

	w = httptest.NewRecorder()
	body = mustMarshal(t, map[string]string{"refresh_token": newRefreshToken})
	req = newJSONRequest(t, http.MethodPost, "/api/v1/auth/refresh", body)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "new refresh token must be accepted with 200")

	require.NoError(t, mock.ExpectationsWereMet())
}


func mustMarshal(t *testing.T, v interface{}) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return b
}

func newJSONRequest(t *testing.T, method, path string, body []byte) *http.Request {
	t.Helper()
	req, err := http.NewRequest(method, path, bytes.NewBuffer(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	return req
}

func mustParseUUID(t *testing.T, s string) uuid.UUID {
	t.Helper()
	id, err := uuid.FromString(s)
	require.NoError(t, err, "value %q is not a valid UUID refresh token", s)
	return id
}
