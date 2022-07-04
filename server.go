package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"regexp"

	_ "github.com/go-sql-driver/mysql"
)

//http:communicates with other apps (standardGoLibrary) to create servers
type Student struct {
	Name    string  `json:"name" validate:"required"`
	Mail    string  `json:"mail" validate:"required,email"`
	ClassId string  `json:"id" validate:"required"`
	Score   float64 `json:"score" validate:"required,gte=0,lte=20"`
}
type Class struct {
	ClassId    string `json:"id" validate:"required"`
	Instructor string
	Lecture    string
}

func server() {
	studentMap := map[string]Student{}
	classMap := map[string]Class{}
	//db
	//sql.open  returns a *sql.DB
	//no driver connection ,no database connection establishment just abstraction
	//user:password[@protocol](localhost adress:port number) []->optional
	//based on http://go-database-sql.org/index.html
	db, err := sql.Open("mysql",
		"root:@tcp(127.0.0.1:3306)/web_project_back")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	//check if database is available and accessible
	err = db.Ping()
	if err != nil {
		fmt.Println(err)
	}
	// An resstudent slice to hold data from returned rows (student)
	rows, err := db.Query("SELECT * FROM student")
	if err != nil {
		log.Fatal(err)
	}
	//executed when the function exits
	defer rows.Close()
	// Loop through rows, using Scan to assign column data to struct fields
	for rows.Next() {
		var student Student
		err := rows.Scan(&student.Name, &student.ClassId, &student.Mail, &student.Score)
		studentMap[student.Name] = student

		if err != nil {
			log.Fatal(err)
		}
	}
	// An resclass slice to hold data from returned rows (class)
	rowc, err := db.Query("SELECT * FROM class")
	if err != nil {
		log.Fatal(err)
	}
	//executed when the function exits
	defer rowc.Close()
	// Loop through rows, using Scan to assign column data to struct fields
	for rowc.Next() {
		var class Class
		//rowc.Scan returns rows
		err := rowc.Scan(&class.ClassId, &class.Instructor, &class.Lecture)
		classMap[class.ClassId] = class

		if err != nil {
			log.Fatal(err)
		}
	}
	http.HandleFunc("/make_student", func(w http.ResponseWriter, r *http.Request) {
		//Create a slice with space preallocated
		body := make([]byte, 256)

		defer r.Body.Close()

		n, _ := r.Body.Read(body)
		body = body[:n]
		//map[string]interface{} type(keys)=string ,type(values)=interface
		var bodyJson map[string]interface{}
		fmt.Printf("%#v\n\n", bodyJson)
		if err := json.Unmarshal(body, &bodyJson); err != nil {
			//status code bad request
			w.WriteHeader(400)
			w.Write([]byte("try"))
			return
		} else {
			if isEmailValid(bodyJson["mail"].(string)) && isScoreValid(bodyJson["score"].(float64)) {
				appendStudent(bodyJson, studentMap, db)
			} else {
				w.WriteHeader(400)
				w.Write([]byte("email format or score is not valid"))
			}

		}
		for key, value := range studentMap {
			fmt.Println(key, value)
		}
		//status code hameh chi khoobe :)
		w.WriteHeader(200)
	})
	http.HandleFunc("/make_class", func(w http.ResponseWriter, r *http.Request) {
		body := make([]byte, 256)
		defer r.Body.Close()
		n, _ := r.Body.Read(body)
		body = body[:n]
		if !checkApiKey(r.Header["Apikey"][0]) {
			w.WriteHeader(400)
			w.Write([]byte("incorrect api key"))
			return
		}
		var bodyJson map[string]interface{}
		fmt.Printf("%#v\n\n", bodyJson)
		if err := json.Unmarshal(body, &bodyJson); err != nil {
			fmt.Println(err)
			w.WriteHeader(400)

		} else {
			fmt.Println(bodyJson["id"])
			appendClass(bodyJson, classMap, db)
		}
		for key, value := range classMap {
			fmt.Println(key, value)
		}
		w.WriteHeader(200)

	})
	http.HandleFunc("/take_class", func(w http.ResponseWriter, r *http.Request) {
		body := make([]byte, 256)
		defer r.Body.Close()
		n, _ := r.Body.Read(body)
		body = body[:n]
		if !checkApiKey(r.Header["Apikey"][0]) {
			w.WriteHeader(400)
			w.Write([]byte("incorrect api key"))
			return
		}
		var bodyJson map[string]interface{}
		fmt.Printf("%#v\n\n", bodyJson)
		if err := json.Unmarshal(body, &bodyJson); err != nil {
			fmt.Println(err)
			w.WriteHeader(400)

		} else {
			takeClass(bodyJson, classMap, db)
		}
		for key, value := range classMap {
			fmt.Println(key, value)
		}
		w.WriteHeader(200)
	})
	http.HandleFunc("/take_student", func(w http.ResponseWriter, r *http.Request) {
		body := make([]byte, 256)
		defer r.Body.Close()
		n, _ := r.Body.Read(body)
		body = body[:n]
		if !checkApiKey(r.Header["Apikey"][0]) {
			w.WriteHeader(400)
			w.Write([]byte("incorrect api key"))
			return
		}
		var bodyJson map[string]interface{}
		fmt.Printf("%#v\n\n", bodyJson)
		if err := json.Unmarshal(body, &bodyJson); err != nil {
			fmt.Println(err)
			w.WriteHeader(400)

		} else {
			takeStudent(bodyJson, studentMap, db)
		}
		for key, value := range studentMap {
			fmt.Println(key, value)
		}
		//status code hameh chi khoobe :)
		w.WriteHeader(200)
	})
	//Send_mail
	http.HandleFunc("/send_mail", func(w http.ResponseWriter, r *http.Request) {
		body := make([]byte, 256)
		defer r.Body.Close()
		n, _ := r.Body.Read(body)
		body = body[:n]
		if !checkApiKey(r.Header["Apikey"][0]) {
			w.WriteHeader(400)
			w.Write([]byte("incorrect api key"))
			return
		}
		var bodyJson map[string]interface{}
		fmt.Printf("%#v\n\n", bodyJson)
		if err := json.Unmarshal(body, &bodyJson); err != nil {
			fmt.Println(err)
			w.WriteHeader(400)

		} else {
			response := send_allmail(bodyJson, studentMap, classMap)
			response_string := "mail result:\n"
			for i := 0; i < len(response); i++ {
				fmt.Printf("%x ", response[i])
				response_string += response[i] + "\n"
			}
			w.WriteHeader(200)
			w.Write([]byte(response_string))
		}
		//status code hameh chi khoobe :)
		w.WriteHeader(200)
	})

	fmt.Printf("%v\n", studentMap)
	fmt.Printf("%v\n", classMap)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func checkApiKey(s string) bool {
	if s == "54321" {
		return true
	}
	return false
}

func isEmailValid(e string) bool {
	//^ start
	//. match any single character
	//\ escape
	// $ end
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	return emailRegex.MatchString(e)
}
func isScoreValid(score float64) bool {
	if score >= 0 && score <= 20 {
		return true
	}
	return false
}

//func make_s(bodyJson map[string]interface{},
// studentMap interface{}, db interface{}) {

//}
func appendStudent(bodyJson map[string]interface{}, studentMap map[string]Student, db *sql.DB) int {
	//if exist
	if _, ok := studentMap[bodyJson["name"].(string)]; ok {
		takeStudent(bodyJson, studentMap, db)
	}
	studentMap[bodyJson["name"].(string)] =
		Student{bodyJson["name"].(string),
			bodyJson["mail"].(string),
			bodyJson["id"].(string),
			bodyJson["score"].(float64)}
	insert := "INSERT INTO student (name, id, email, score) VALUES ('" + bodyJson["name"].(string) + "','" + bodyJson["mail"].(string) + "','" + bodyJson["id"].(string) + "','" + fmt.Sprintf("%v", bodyJson["score"].(float64)) + "');"
	rows, err := db.Query(insert)
	if err != nil {
		fmt.Println("student insertion Error")
	} else {
		fmt.Println(rows)
	}
	defer rows.Close()
	return 0
}

func appendClass(bodyJson map[string]interface{}, classMap map[string]Class, db *sql.DB) int {
	if _, ok := classMap[bodyJson["id"].(string)]; ok {
		takeClass(bodyJson, classMap, db)
	}
	classMap[bodyJson["id"].(string)] =
		Class{bodyJson["id"].(string),
			bodyJson["instructor"].(string),
			bodyJson["lecture"].(string)}
	insert := "INSERT INTO class (id,instructor,lecture) VALUES ('" + bodyJson["id"].(string) + "','" + bodyJson["instructor"].(string) + "','" + bodyJson["lecture"].(string) + "');"
	rows, err := db.Query(insert)
	if err != nil {
		fmt.Println("class insertion Error")
	} else {
		fmt.Println(rows)
	}
	defer rows.Close()
	return 0
}
func takeClass(bodyJson map[string]interface{}, classMap map[string]Class, db *sql.DB) {
	//clear string from map of classes with id
	delete(classMap, bodyJson["id"].(string))
	delete := "DELETE FROM class WHERE class.id = \"" + bodyJson["id"].(string) + "\""
	rows, err := db.Query(delete)
	if err != nil {
		fmt.Println("class deletion Error")
	}
	defer rows.Close()
}

func takeStudent(bodyJson map[string]interface{}, studentMap map[string]Student, db *sql.DB) {
	//clear string from map of students with name
	delete(studentMap, bodyJson["name"].(string))
	delete := "DELETE FROM student WHERE student.name = \"" + bodyJson["name"].(string) + "\""
	rows, err := db.Query(delete)
	if err != nil {
		fmt.Println("class deletion Error")
	}
	defer rows.Close()
}

//send_mail
func send_email(sName string, smail string, sscore float64, lec string, ins string, res chan string) {
	from := "mahlbash79@gmail.com"
	to := []string{smail}
	password := ""
	smtpHost := "smtp.gmail.com"
	//465 (SSL)/587 (TLS)
	smtpPort := "587"
	//smtp.PlainAuth    ->returns an Auth

	//should be sprintf
	auth := smtp.PlainAuth("mahlbash79@gmail.com", from, password, smtpHost)
	msg := []byte("From:\r\n" + "To:" + smail + "\r\n" + "Subject: message for you from instructor ( hw3 go program ) \r\n\r\n" + "hello ms/mrs: " + sName + "\n" + " your score for " + lec + " lecture is : " + fmt.Sprintf("%v", sscore) + "\n\n" + "teacher: " + ins + "\r\n")
	//send message
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, msg)
	if err != nil {
		log.Fatal(err)
		res <- smail + "failed"
		return
	}
	res <- smail + "sent"
	return
}
func send_allmail(bodyJson map[string]interface{}, studentMap map[string]Student, classMap map[string]Class) []string {
	id := bodyJson["id"].(string)
	class := classMap[id]
	lecture := class.Lecture
	instructor := class.Instructor
	response := []string{}
	for key, val := range studentMap {
		if val.ClassId == id {
			//allocate send channel result
			result := make(chan string, 1)
			fmt.Println("top " + val.Mail)
			go send_email(key, val.Mail, val.Score, lecture, instructor, result)
			res := <-result
			fmt.Printf("result: %d\n", res)
			response = append(response, res)
			//close(res)

			fmt.Println(res)
		}
	}
	return response
}

func main() {
	server()
}
