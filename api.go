package main

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type API struct {
	catalog        *Catalog
	allowedOrigins map[string]bool
}

type CourseDetailResponse struct {
	Course    Course          `json:"course"`
	Parse     *ParseResult    `json:"parse,omitempty"`
	Tree      *DependencyTree `json:"tree,omitempty"`
	DataNotes string          `json:"data_notes,omitempty"`
}

func NewRouter(catalog *Catalog, allowedOrigin string) *gin.Engine {
	allowedOrigins := make(map[string]bool)
	for _, origin := range strings.Split(allowedOrigin, ",") {
		origin = strings.TrimRight(strings.TrimSpace(origin), "/")
		if origin != "" {
			allowedOrigins[origin] = true
		}
	}
	api := &API{catalog: catalog, allowedOrigins: allowedOrigins}
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery(), api.corsMiddleware())
	router.GET("/api/health", api.health)
	router.GET("/api/courses", api.searchCourses)
	router.GET("/api/courses/:subject/:code", api.getCourse)
	return router
}

func (a *API) corsMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		origin := strings.TrimRight(ctx.GetHeader("Origin"), "/")
		if origin != "" && a.allowedOrigins[origin] {
			ctx.Header("Access-Control-Allow-Origin", origin)
			ctx.Header("Vary", "Origin")
			ctx.Header("Access-Control-Allow-Headers", "Content-Type")
			ctx.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
		}
		if ctx.Request.Method == http.MethodOptions {
			ctx.AbortWithStatus(http.StatusNoContent)
			return
		}
		ctx.Next()
	}
}

func (a *API) health(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"status":         "ok",
		"courses":        a.catalog.Size(),
		"subject_count":  a.catalog.SubjectCount(),
		"parser_version": 1,
	})
}

func (a *API) searchCourses(ctx *gin.Context) {
	query := strings.TrimSpace(ctx.Query("q"))
	limit := 25
	if rawLimit := ctx.Query("limit"); rawLimit != "" {
		parsed, err := strconv.Atoi(rawLimit)
		if err != nil || parsed < 1 || parsed > 100 {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "limit must be an integer between 1 and 100"})
			return
		}
		limit = parsed
	}
	page := 1
	if rawPage := ctx.Query("page"); rawPage != "" {
		parsed, err := strconv.Atoi(rawPage)
		if err != nil || parsed < 1 {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "page must be a positive integer"})
			return
		}
		page = parsed
	}
	maxInt := int(^uint(0) >> 1)
	if page-1 > maxInt/limit {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "page is too large"})
		return
	}
	results, total := a.catalog.SearchPage(query, (page-1)*limit, limit)
	totalPages := 0
	if total > 0 {
		totalPages = (total + limit - 1) / limit
	}
	ctx.JSON(http.StatusOK, gin.H{
		"query":       query,
		"results":     results,
		"page":        page,
		"page_size":   limit,
		"total":       total,
		"total_pages": totalPages,
	})
}

func (a *API) getCourse(ctx *gin.Context) {
	courseName := normalizeCourseName(ctx.Param("subject") + " " + ctx.Param("code"))
	course, ok := a.catalog.GetCourse(courseName)
	if !ok {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "course not found", "course": courseName})
		return
	}

	depth := 6
	if rawDepth := ctx.Query("depth"); rawDepth != "" {
		parsed, err := strconv.Atoi(rawDepth)
		if err != nil || parsed < 0 || parsed > 12 {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "depth must be an integer between 0 and 12"})
			return
		}
		depth = parsed
	}

	response := CourseDetailResponse{Course: *course}
	if course.CoursePrerequisites != "" {
		parsed := ParseRequisites(course.CoursePrerequisites)
		response.Parse = &parsed
		tree, _ := a.catalog.BuildDependencyTree(courseName, depth)
		response.Tree = tree
	}
	ctx.JSON(http.StatusOK, response)
}
