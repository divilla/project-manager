package shared

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Client struct {
	baseURL string
	http    *http.Client
}

type cleanupTask struct {
	ID       int  `json:"id"`
	ParentID *int `json:"parent_id"`
}

func NewClient(t *testing.T) *Client {
	t.Helper()

	baseURL := strings.TrimRight(os.Getenv("AIPM_API_BASE_URL"), "/")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	client := &Client{
		baseURL: baseURL,
		http: &http.Client{
			Timeout: 5 * time.Second,
		},
	}

	req, err := http.NewRequest(http.MethodGet, client.baseURL+"/api/v1/health", nil)
	require.NoError(t, err)

	res, err := client.http.Do(req)
	require.NoErrorf(t, err, "backend is not available at %s", client.baseURL)
	defer res.Body.Close()

	return client
}

func NewDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	databaseURL := os.Getenv("AIPM_DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:postgres@localhost:5432/project_manager_test?sslmode=disable"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, databaseURL)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	require.NoError(t, pool.Ping(ctx))

	return pool
}

func AssertHistoryNotDeleted(t *testing.T, db *pgxpool.Pool, table string, id int) {
	t.Helper()

	AssertHistoryCountAtLeast(t, db, table, id, false, 1)
}

func AssertHistoryDeleted(t *testing.T, db *pgxpool.Pool, table string, id int) {
	t.Helper()

	AssertHistoryCountAtLeast(t, db, table, id, true, 1)
}

func AssertHistoryCountAtLeast(t *testing.T, db *pgxpool.Pool, table string, id int, deleted bool, minimum int) {
	t.Helper()

	var count int
	err := db.QueryRow(context.Background(), "select count(*) from "+table+" where id = $1 and deleted = $2", id, deleted).Scan(&count)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, count, minimum)
}

func AssertHistoryCount(t *testing.T, db *pgxpool.Pool, table string, id int, deleted bool, expected int) {
	t.Helper()
	var count int
	err := db.QueryRow(context.Background(), "select count(*) from "+table+" where id = $1 and deleted = $2", id, deleted).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, expected, count)
}

func CleanupProject(t *testing.T, client *Client, projectID int) {
	t.Helper()

	var tasks []cleanupTask
	status := client.Post(t, "/api/v1/task/list", map[string]any{"project_id": projectID}, &tasks)
	if status == http.StatusOK {
		for _, task := range tasks {
			if task.ParentID != nil {
				continue
			}
			status = client.Post(t, "/api/v1/task/delete", map[string]any{"id": task.ID}, nil)
			assert.Contains(t, []int{http.StatusNoContent, http.StatusNotFound}, status)
		}
		for _, task := range tasks {
			status = client.Post(t, "/api/v1/task/delete", map[string]any{"id": task.ID}, nil)
			assert.Contains(t, []int{http.StatusNoContent, http.StatusNotFound}, status)
		}
	}

	status = client.Post(t, "/api/v1/project/delete", map[string]any{"id": projectID}, nil)
	assert.Contains(t, []int{http.StatusNoContent, http.StatusNotFound}, status)
}

func (c *Client) Get(t *testing.T, path string, out any) int {
	t.Helper()

	req, err := http.NewRequest(http.MethodGet, c.baseURL+path, nil)
	require.NoError(t, err)

	res, err := c.http.Do(req)
	require.NoError(t, err)
	defer res.Body.Close()

	if out != nil {
		require.NoError(t, json.NewDecoder(res.Body).Decode(out))
	}

	return res.StatusCode
}

func (c *Client) Post(t *testing.T, path string, body any, out any) int {
	t.Helper()

	payload, err := json.Marshal(body)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, c.baseURL+path, bytes.NewReader(payload))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	res, err := c.http.Do(req)
	require.NoError(t, err)
	defer res.Body.Close()

	if out != nil {
		data, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(data, out), fmt.Sprintf("response body: %s", data))
	}

	return res.StatusCode
}
