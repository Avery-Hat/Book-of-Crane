package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/Avery-Hat/Book-of-Crane/internal/handler"
	"github.com/Avery-Hat/Book-of-Crane/internal/model"
	"github.com/Avery-Hat/Book-of-Crane/internal/store"
	"github.com/jackc/pgx/v5/pgxpool"
)

func testDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://campaign:campaign@localhost:5433/campaign_notes?sslmode=disable"
	}
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		t.Fatalf("connect to test db: %v", err)
	}
	if err := pool.Ping(context.Background()); err != nil {
		t.Fatalf("ping test db: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func resetCampaigns(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	_, err := pool.Exec(context.Background(), "DELETE FROM campaigns")
	if err != nil {
		t.Fatalf("reset campaigns: %v", err)
	}
}

func newRouter(pool *pgxpool.Pool) http.Handler {
	s := store.NewCampaignStore(pool)
	h := handler.NewCampaignHandler(s)
	return h.Routes()
}

func parseBody(t *testing.T, res *httptest.ResponseRecorder, dst any) {
	t.Helper()
	if err := json.NewDecoder(res.Body).Decode(dst); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}

func mustCreateCampaign(t *testing.T, pool *pgxpool.Pool, name, description string) model.Campaign {
	t.Helper()
	s := store.NewCampaignStore(pool)
	desc := &description
	c, err := s.Create(context.Background(), model.CreateCampaignRequest{Name: name, Description: desc})
	if err != nil {
		t.Fatalf("seed campaign: %v", err)
	}
	return c
}

// --- tests ---

func TestCampaignList_Empty(t *testing.T) {
	pool := testDB(t)
	resetCampaigns(t, pool)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	newRouter(pool).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
	var got []model.Campaign
	parseBody(t, rec, &got)
	if len(got) != 0 {
		t.Fatalf("want empty list, got %d items", len(got))
	}
}

func TestCampaignList_ReturnsSeedData(t *testing.T) {
	pool := testDB(t)
	resetCampaigns(t, pool)
	mustCreateCampaign(t, pool, "Campaign A", "desc a")
	mustCreateCampaign(t, pool, "Campaign B", "desc b")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	newRouter(pool).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
	var got []model.Campaign
	parseBody(t, rec, &got)
	if len(got) != 2 {
		t.Fatalf("want 2 campaigns, got %d", len(got))
	}
}

func TestCampaignCreate_Success(t *testing.T) {
	pool := testDB(t)
	resetCampaigns(t, pool)

	body := `{"name":"New Campaign","description":"a great campaign"}`
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()
	newRouter(pool).ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("want 201, got %d — body: %s", rec.Code, rec.Body.String())
	}
	var got model.Campaign
	parseBody(t, rec, &got)
	if got.ID == 0 {
		t.Fatal("expected non-zero ID")
	}
	if got.Name != "New Campaign" {
		t.Fatalf("want name %q, got %q", "New Campaign", got.Name)
	}
}

func TestCampaignCreate_MissingName(t *testing.T) {
	pool := testDB(t)
	resetCampaigns(t, pool)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"description":"no name"}`))
	rec := httptest.NewRecorder()
	newRouter(pool).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rec.Code)
	}
}

func TestCampaignCreate_InvalidBody(t *testing.T) {
	pool := testDB(t)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`not json`))
	rec := httptest.NewRecorder()
	newRouter(pool).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rec.Code)
	}
}

func TestCampaignGet_Success(t *testing.T) {
	pool := testDB(t)
	resetCampaigns(t, pool)
	c := mustCreateCampaign(t, pool, "Found It", "details")

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/%d", c.ID), nil)
	rec := httptest.NewRecorder()
	newRouter(pool).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
	var got model.Campaign
	parseBody(t, rec, &got)
	if got.ID != c.ID {
		t.Fatalf("want id %d, got %d", c.ID, got.ID)
	}
}

func TestCampaignGet_NotFound(t *testing.T) {
	pool := testDB(t)
	resetCampaigns(t, pool)

	req := httptest.NewRequest(http.MethodGet, "/99999", nil)
	rec := httptest.NewRecorder()
	newRouter(pool).ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d", rec.Code)
	}
}

func TestCampaignGet_InvalidID(t *testing.T) {
	pool := testDB(t)

	req := httptest.NewRequest(http.MethodGet, "/notanid", nil)
	rec := httptest.NewRecorder()
	newRouter(pool).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rec.Code)
	}
}

func TestCampaignUpdate_Success(t *testing.T) {
	pool := testDB(t)
	resetCampaigns(t, pool)
	c := mustCreateCampaign(t, pool, "Old Name", "old desc")

	body := `{"name":"New Name"}`
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/%d", c.ID), bytes.NewBufferString(body))
	rec := httptest.NewRecorder()
	newRouter(pool).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d — body: %s", rec.Code, rec.Body.String())
	}
	var got model.Campaign
	parseBody(t, rec, &got)
	if got.Name != "New Name" {
		t.Fatalf("want name %q, got %q", "New Name", got.Name)
	}
}

func TestCampaignUpdate_NotFound(t *testing.T) {
	pool := testDB(t)
	resetCampaigns(t, pool)

	req := httptest.NewRequest(http.MethodPut, "/99999", bytes.NewBufferString(`{"name":"x"}`))
	rec := httptest.NewRecorder()
	newRouter(pool).ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d", rec.Code)
	}
}

func TestCampaignDelete_Success(t *testing.T) {
	pool := testDB(t)
	resetCampaigns(t, pool)
	c := mustCreateCampaign(t, pool, "Delete Me", "")

	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/%d", c.ID), nil)
	rec := httptest.NewRecorder()
	newRouter(pool).ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("want 204, got %d", rec.Code)
	}

	// Confirm it's gone
	req2 := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/%d", c.ID), nil)
	rec2 := httptest.NewRecorder()
	newRouter(pool).ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusNotFound {
		t.Fatalf("want 404 after delete, got %d", rec2.Code)
	}
}

func TestCampaignDelete_NotFound(t *testing.T) {
	pool := testDB(t)
	resetCampaigns(t, pool)

	req := httptest.NewRequest(http.MethodDelete, "/99999", nil)
	rec := httptest.NewRecorder()
	newRouter(pool).ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d", rec.Code)
	}
}
