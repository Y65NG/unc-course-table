package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCourseAPI(t *testing.T) {
	gin.SetMode(gin.TestMode)
	courses := []Course{
		Course{CourseCategory: "COMP", CourseCode: "210", CourseTitle: "Data Structures", CoursePrerequisites: "Prerequisite, COMP 110."},
		Course{CourseCategory: "COMP", CourseCode: "110", CourseTitle: "Introduction to Programming"},
	}
	for index := 0; index < 63; index++ {
		courses = append(courses, Course{CourseCategory: "COMP", CourseCode: fmt.Sprintf("9%02d", index), CourseTitle: "Pagination fixture"})
	}
	catalog := testCatalog(courses...)
	router := NewRouter(catalog, "http://localhost:5173")

	t.Run("search", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/api/courses?q=comp&limit=30&page=2", nil)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		if response.Code != http.StatusOK {
			t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
		}
		var body struct {
			Results    []CourseSummary `json:"results"`
			Page       int             `json:"page"`
			PageSize   int             `json:"page_size"`
			Total      int             `json:"total"`
			TotalPages int             `json:"total_pages"`
		}
		if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		if len(body.Results) != 30 || body.Page != 2 || body.PageSize != 30 || body.Total != 65 || body.TotalPages != 3 {
			t.Fatalf("unexpected pagination response: %#v", body)
		}
	})

	t.Run("invalid page", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/api/courses?q=comp&page=0", nil)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		if response.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
		}
	})

	t.Run("exact course includes parse and tree", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/api/courses/COMP/210?depth=3", nil)
		request.Header.Set("Origin", "http://localhost:5173")
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		if response.Code != http.StatusOK {
			t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
		}
		if response.Header().Get("Access-Control-Allow-Origin") != "http://localhost:5173" {
			t.Error("expected allowed CORS origin")
		}
		var body CourseDetailResponse
		if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		if body.Parse == nil || body.Parse.Status != ParseComplete || body.Tree == nil {
			t.Fatalf("unexpected detail response: %#v", body)
		}
	})

	t.Run("invalid depth", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/api/courses/COMP/210?depth=99", nil)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		if response.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
		}
	})

	t.Run("not found", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/api/courses/COMP/999", nil)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		if response.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want %d", response.Code, http.StatusNotFound)
		}
	})
}
