package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	cors "github.com/rs/cors/wrapper/gin"
)

const Debug = false

func DPrintf(format string, a ...any) {
	if Debug {
		log.Printf(format, a...)
	}
}

func DPrintln(a ...any) {
	if Debug {
		log.Println(a...)
	}
}

var courseSystem *CourseSystem

func init() {
	courseSystem = NewCourseSystem()
	// courseSystem.ConnectDB()
	if len(os.Args) > 1 && os.Args[1] == "recache" {
		courseSystem.Recache()
		// courseSystem.Store()
	}
	courseSystem.LoadCache("courses.json")

}

func main() {

	if !Debug {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.Default()
	// router.Use(CORSMiddleware())
	router.Use(cors.Default())

	router.GET("", GetCourses)
	// router.GET("/info", GetCourse)

	router.Run()
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST, HEAD, PATCH, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func GetCourses(ctx *gin.Context) {
	courseStr := ctx.Query("course")
	if len(courseStr) == 0 {
		courseStr = "AA"
	}

	if len(courseStr) > 4 && courseStr[4] != ' ' {
		courseStr = courseStr[:4] + " " + courseStr[4:]
	}

	courses := courseSystem.GetCourses(courseStr)
	if len(courses) == 1 {
		course := courses[0]
		course.CourseTree = Interpret(*course)
		ctx.JSON(200, courses)
		return
	}

	ctx.JSON(200, courses)
}
