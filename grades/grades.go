package grades

import (
	"fmt"
	"sync"
)

type Student struct { //学生的个人信息
	ID        int
	FirstName string
	LastName  string
	Grades    []Grade
}

func (s Student) Average() float32 { //计算学生的平均成绩
	var result float32
	for _, grade := range s.Grades {
		result += grade.Score
	}

	return result / float32(len(s.Grades))
}

type Students []Student //学生的集合

var (
	students      Students
	studentsMutex sync.Mutex
)

func (ss Students) GetByID(id int) (*Student, error) {
	for i := range ss {
		if ss[i].ID == id {
			return &ss[i], nil
		}
	}

	return nil, fmt.Errorf("Student with ID %d not found", id)
}

type GradeType string

const (
	GradeQuiz = GradeType("Quiz")
	GradeTest = GradeType("Test")
	GradeExam = GradeType("Exam")
)

type Grade struct {
	Title string
	Type  GradeType
	Score float32
}
