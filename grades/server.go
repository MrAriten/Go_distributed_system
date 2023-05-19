package grades

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func RegisterHandlers() {
	handler := new(studentsHandler)
	http.Handle("/students", handler)
	http.Handle("/students/", handler)
}

type studentsHandler struct{} //创建handler的空接口

// /students case 2
// /students/{id} case 3
// /students/{id}/grades case 4
func (sh studentsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	pathSegments := strings.Split(r.URL.Path, "/") //按照URL的斜杠数分类
	switch len(pathSegments) {
	case 2: //查询所有学生
		sh.getAll(w, r)
	case 3:
		id, err := strconv.Atoi(pathSegments[2])
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		sh.getOne(w, r, id)
	case 4:
		id, err := strconv.Atoi(pathSegments[2])
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		sh.addGrade(w, r, id)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func (sh studentsHandler) getAll(w http.ResponseWriter, r *http.Request) { //获取所有学生的信息
	studentsMutex.Lock()
	defer studentsMutex.Unlock()

	data, err := sh.toJSON(students)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.Write(data)
}

func (sh studentsHandler) getOne(w http.ResponseWriter, r *http.Request, id int) { //只获取一个学生的信息
	studentsMutex.Lock()
	defer studentsMutex.Unlock()

	student, err := students.GetByID(id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		log.Println(err)
		return
	}

	data, err := sh.toJSON(student)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Failed to serialize student: %q", err)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.Write(data)
}

func (sh studentsHandler) addGrade(w http.ResponseWriter, r *http.Request, id int) { //添加成绩
	studentsMutex.Lock()
	defer studentsMutex.Unlock()

	student, err := students.GetByID(id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		log.Println(err)
		return
	}
	var g Grade
	dec := json.NewDecoder(r.Body)
	err = dec.Decode(&g)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println(err)
		return
	}
	student.Grades = append(student.Grades, g)
	w.WriteHeader(http.StatusCreated)
	data, err := sh.toJSON(g)
	if err != nil {
		log.Println(err)
		return
	}
	w.Header().Add("Content-Type", "applicaiton/json")
	w.Write(data)
}

func (sh studentsHandler) toJSON(obj interface{}) ([]byte, error) { //将
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	err := enc.Encode(obj)
	if err != nil {
		return nil, fmt.Errorf("Failed to serialize students: %q", err)
	}
	return b.Bytes(), nil
}
