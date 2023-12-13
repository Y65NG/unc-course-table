package main

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMain(t *testing.T) {
	assert := assert.New(t)
	arr := []int{1, 2, 3, 4, 5}
	for i := 0; i < len(arr); i++ {
		log.Println(i, arr[i])
		arr = arr[:len(arr)-1]

	}
	assert.Equal(1, 2)

}

func TestConnection(t *testing.T) {
	courses := courseSystem.GetCourses("COMP 21")
	for _, course := range courses {
		log.Println(course.GetName())
	}
}
