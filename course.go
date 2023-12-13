package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Course struct {
	CourseCategory      string      `json:"course_category" bson:"course_category"`
	CourseCode          string      `json:"course_code" bson:"course_code"`
	CourseTitle         string      `json:"course_title" bson:"course_title"`
	CourseCredit        string      `json:"course_credit" bson:"course_credit"`
	CourseDescription   string      `json:"course_description" bson:"course_description"`
	CoursePrerequisites string      `json:"course_prerequisites" bson:"course_prerequisites"`
	CourseTree          *CourseTree `json:"course_tree"`
}

func NewCourse(courseCategory, courseCode, courseTitle, courseCredit, courseDescription, coursePrerequisites string) Course {
	return Course{
		CourseCategory:      courseCategory,
		CourseCode:          courseCode,
		CourseTitle:         courseTitle,
		CourseCredit:        courseCredit,
		CourseDescription:   courseDescription,
		CoursePrerequisites: coursePrerequisites,
	}
}

func (self Course) GetName() string {
	return self.CourseCategory + " " + self.CourseCode
}

func (self Course) ToString() string {
	prerequisite, ok := courseSystem.GetPrerequisites(self.GetName())
	if !ok {
		prerequisite = self.CoursePrerequisites
	}
	return "Course Name: " + self.GetName() + "\n" +
		"Course Title: " + self.CourseTitle + "\n" +
		"Course Credit: " + self.CourseCredit + "\n" +
		"Course Description: " + self.CourseDescription + "\n" +
		self.CoursePrerequisites + "\n" +
		"Course Tree: \n" + prerequisite
}

type CourseSystem struct {
	Courses map[string][]Course
	db      *mongo.Client
}

func NewCourseSystem() *CourseSystem {
	return &CourseSystem{
		Courses: make(map[string][]Course),
	}
}

func (self *CourseSystem) ConnectDB() {
	// var mongoUri = os.Getenv("MONGO_URI")
	var mongoUri = "mongodb://localhost:27017"
	DPrintln(mongoUri)

	// CONNECT TO YOUR ATLAS CLUSTER:
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(
		mongoUri,
	))
	if err != nil {
		panic(err)
	}

	err = client.Ping(ctx, nil)

	if err != nil {
		fmt.Println("There was a problem connecting to your Atlas cluster. Check that the URI includes a valid username and password, and that your IP address has been added to the access list. Error: ")
		panic(err)
	}

	self.db = client
	fmt.Println("Connected to MongoDB!")
}

func (self *CourseSystem) Store() {
	client := self.db

	// Provide the name of the database and collection you want to use.
	// If they don't already exist, the driver and Atlas will create them
	// automatically when you first write data.
	var dbName = "CourseTable"
	var collectionName = "CourseData"
	collection := client.Database(dbName).Collection(collectionName)
	var allCourses []any
	for _, courses := range self.Courses {
		for _, course := range courses {
			course := course
			allCourses = append(allCourses, &course)
		}
	}
	insertResult, err := collection.InsertMany(context.Background(), allCourses)
	if err != nil {
		fmt.Println("Something went wrong trying to insert the new documents:")
		panic(err)
	}
	fmt.Println(len(insertResult.InsertedIDs), "new documents inserted to "+collectionName+" collection!")
}

func (self *CourseSystem) LoadCache(path string) {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	reader := json.NewDecoder(f)
	if err = reader.Decode(&self.Courses); err != nil {
		panic(err)
	}
}

func (self *CourseSystem) Size() int {
	var count int
	for _, courses := range self.Courses {
		count += len(courses)
	}
	return count
}

func (self *CourseSystem) AllCourses() []*Course {
	var result []*Course
	for _, courses := range self.Courses {
		for _, course := range courses {
			result = append(result, &course)
		}
	}
	return result
}

func (self *CourseSystem) AddCourse(course Course) {
	self.Courses[course.CourseCategory] = append(self.Courses[course.CourseCategory], course)
}

func (self *CourseSystem) GetCourses(courseStr string) []*Course {
	// data := strings.Split(courseStr, " ")
	// var category, code string
	// if len(data) == 2 {
	// 	category, code = data[0], data[1]
	// } else if len(data) == 1 {
	// 	category, code = data[0], ""
	// } else {
	// 	return nil
	// }
	// var result []*Course
	// var filter bson.M
	// if len(data[0]) == 4 {
	// 	filter = bson.M{"course_category": category, "course_code": bson.M{"$regex": "^" + code}}
	// } else {
	// 	filter = bson.M{"course_category": bson.M{"$regex": "^" + category}, "course_code": bson.M{"$regex": "^" + code}}
	// }
	// opts := options.Find()
	// opts.SetSort(bson.D{{Key: "course_code", Value: 1}})
	// opts.SetLimit(50)

	// cursor, err := self.db.Database("CourseTable").Collection("CourseData").Find(context.Background(), filter, opts)
	// if err != nil {
	// 	return nil
	// }

	// defer cursor.Close(context.Background())

	// for cursor.Next(context.Background()) {
	// 	var course Course
	// 	if err = cursor.Decode(&course); err != nil {
	// 		return nil
	// 	}
	// 	result = append(result, &course)
	// }

	var result []*Course
	var count int
L:
	for _, courses := range self.Courses {
		for _, course := range courses {
			if count > 100 {
				break L
			}
			course := course
			if strings.HasPrefix(course.GetName(), courseStr) {
				result = append(result, &course)
				count++
			}
		}
	}
	return result
}

func (self *CourseSystem) GetCourse(courseStr string) (*Course, bool) {
	data := strings.Split(courseStr, " ")
	var category, code string
	if len(data) == 2 {
		category, code = data[0], data[1]
	} else if len(data) == 1 {
		category, code = data[0], ""
	} else {
		return nil, false
	}

	// var course Course
	// filter := bson.M{"course_category": category, "course_code": code}
	// e := self.db.Database("CourseTable").Collection("CourseData").FindOne(context.Background(), filter).Decode(&course)
	// if e != nil {
	// 	// log.Println(e)
	// 	return nil, false
	// }
	// return &course, true

	for _, course := range self.Courses[category] {
		if course.CourseCode == code {
			return &course, true
		}
	}
	return nil, false
}

func (self *CourseSystem) Recache() {
	var URL = "https://catalog.unc.edu/courses"
	collector := colly.NewCollector(
		colly.CacheDir("./courses_cache"),
		colly.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36"),
	)

	detailCollector := collector.Clone()
	collector.OnHTML("a[href]", func(h *colly.HTMLElement) {
		if strings.Contains(h.Attr("href"), "courses/") {
			log.Println("Visiting", h.Request.AbsoluteURL(h.Attr("href")))
			detailCollector.Visit(h.Request.AbsoluteURL(h.Attr("href")))
		}
	})

	detailCollector.OnHTML("div.courseblock", func(h *colly.HTMLElement) {
		data := strings.Split(h.ChildText("span.detail-code"), " ")
		courseCategory := data[0]
		courseCode := data[1]
		courseCode = strings.ReplaceAll(courseCode, ".", "")
		courseTitle := h.ChildText("span.detail-title")
		courseCredit := strings.Split(h.ChildText("span.detail-hours"), " ")[0]
		courseDescription := h.ChildText("p.courseblockextra")
		coursePrerequisitesStr := h.ChildText("span.detail-requisites")

		course := NewCourse(
			courseCategory,
			courseCode,
			courseTitle,
			courseCredit,
			courseDescription,
			coursePrerequisitesStr,
		)
		// course.CoursePrerequisites.ParseText()
		self.AddCourse(course)
	})
	if err := collector.Visit(URL); err != nil {
		log.Println(err)
	}
	// if err := self.ToJSON("./courses.json"); err != nil {
	// 	panic(err)
	// }

	self.Store()
}

func (self *CourseSystem) GetPrerequisites(courseStr string) (string, bool) {
	course, ok := self.GetCourse(courseStr)
	if !ok {
		DPrintln("Course not found!")
		return "", false
	}
	if len(course.CoursePrerequisites) == 0 {
		return "", false
	}
	tree := Interpret(*course)
	if tree == nil {
		return "", false
	}

	return tree.ToString(), true
}
func (self *CourseSystem) ToJSON(path string) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0777)
	if err != nil {
		return err
	}
	data, err := json.Marshal(self.Courses)
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	return nil
}
