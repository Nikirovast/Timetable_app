package handler

import (
	"database/sql"
	"fmt"
	"github.com/labstack/echo"
	_ "github.com/lib/pq"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"time"
)

type (
	User struct {
		ID       bson.ObjectId `json:"id" bson:"_id,omitempty"`
		Email    string        `json:"email" bson:"email" form:"email" query:"email"`
		Password string        `json:"password,omitempty" bson:"password" form:"password" query:"password"`
	}
	Schedule struct {
		Id      int
		Faculty string
		Time    string
		Subject string
		Teacher string
	}
)

func Log(email, password string, c echo.Context) (bool, error) {
	u := new(User)
	// db connection
	db, err := mgo.Dial("localhost:27017")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	// check username and password against DB after hashing the password
	err = db.DB("AppUsers").C("users").Find(bson.M{"email": email, "password": password}).One(u);
	if err != nil {
		if err == mgo.ErrNotFound {
			return false, &echo.HTTPError{Code: http.StatusUnauthorized, Message: "invalid email or password"}
		}
		return false, err
	}
	emailcookie := new(http.Cookie)
	passwordcookie := new(http.Cookie)
	emailcookie.Name = "email"
	passwordcookie.Name = "password"
	emailcookie.Value = email
	passwordcookie.Value = password
	emailcookie.Expires = time.Now().Add(48 * time.Hour)
	emailcookie.Expires = time.Now().Add(48 * time.Hour)
	c.SetCookie(emailcookie)
	c.SetCookie(passwordcookie)
	return true, c.String(http.StatusOK, "You were logged in!")
}
func Login(c echo.Context) (err error) {
	u := new(User)
	err = c.Bind(u)
	fmt.Println("email \n", u.Email)
	fmt.Println("password \n", u.Password)
	if err != nil {
		return err
	}
	// db connection
	db, err := mgo.Dial("localhost:27017")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	// check username and password against DB after hashing the password
	err = db.DB("AppUsers").C("users").Find(bson.M{"email": u.Email, "password": u.Password}).One(u);
	if err != nil {
		if err == mgo.ErrNotFound {
			return &echo.HTTPError{Code: http.StatusUnauthorized, Message: "invalid email or password"}
		}
		return err
	}

	cookie := new(http.Cookie)
	cookie.Name = "sessionID"
	cookie.Value = "some_string"
	cookie.Expires = time.Now().Add(48 * time.Hour)

	c.SetCookie(cookie)

	return c.String(http.StatusOK, "You were logged in!")

}


func dbConn() (db *sql.DB) {
	// connecting to database psql
	connStr := "user=postgres password=Artem123 dbname=timetable sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	return db
}
func Show(c echo.Context) (err error) {
	// connecting to database psql
	db := dbConn()
	defer db.Close()
	a := c.QueryParam("faculty")
	b:= c.QueryParam("day")
	query := `SELECT schedule.id, intervals.start, intervals.finish, subjects.name, teachers.name FROM schedule
			  INNER JOIN faculties ON facultyid = faculties.id
		 	  INNER JOIN subjects ON subjectid = subjects.id
	 		  INNER JOIN intervals ON intervalid = intervals.id 
	   		  INNER JOIN teachers ON teacherid = teachers.id
			  INNER JOIN days ON dayid = days.id WHERE faculties.name = $1 AND days.name=$2;`
	selDB, err := db.Query(query, a,b)
	if err != nil {
		panic(err.Error())
	}
	sch := Schedule{}
	res := []Schedule{}
	for selDB.Next() {
		var Id int
		var Start, Finish, Subject, Teacher string
		err = selDB.Scan(&Id, &Start, &Finish, &Subject, &Teacher)
		if err != nil {
			fmt.Println(err)
			continue
		}
		sch.Id = Id
		sch.Faculty = a
		sch.Time = Start + " — " + Finish
		sch.Subject = Subject
		sch.Teacher = Teacher
		res = append(res, sch)
	}
	return c.Render(http.StatusOK, "student_timetable.html", res)
}
func Delete(c echo.Context) (err error) {
	db := dbConn()
	pid := c.QueryParam("id")
	fac := c.QueryParam("faculty")
	sqlStatement := `DELETE FROM schedule WHERE id = $1;`
	_, err = db.Exec(sqlStatement, pid)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	return c.String(http.StatusOK, `Was deleted from `+fac)
}
func Insert(c echo.Context) (err error) {
	db := dbConn()
	Faculty:=c.QueryParam("faculty")
	Day := c.QueryParam("day")
	Start := c.QueryParam("start")
	Finish := c.QueryParam("finish")
	Subject := c.QueryParam("subject")
	Teacher := c.QueryParam("teacher")
	var Iid int
	db.QueryRow(`SELECT intervals.id FROM intervals WHERE intervals.start= '`+Start+`' AND intervals.finish = '`+Finish+`';`).Scan(&Iid)
	if (Iid==0){
		db.QueryRow(`INSERT INTO intervals values(nextval('intervals_id_seq'),'`+ Start+`','`+Finish+`') returning id;`).Scan(&Iid)
	}
	var Sid int
	db.QueryRow(`SELECT subjects.id FROM subjects WHERE subjects.name = '`+Subject+`';`).Scan(&Sid)
	if (Sid==0){
		db.QueryRow(`INSERT INTO subjects values(nextval('subjects_id_seq'),'`+ Subject+`') returning id;`).Scan(&Sid)
	}
	var Tid int
	db.QueryRow(`SELECT teachers.id FROM teachers WHERE teachers.name = '`+Teacher+`';`).Scan(&Tid)
	if (Tid==0){
		db.QueryRow(`INSERT INTO teachers values(nextval('teachers_id_seq'),'`+ Teacher+`') returning id;`).Scan(&Tid)
	}
	var Did int
	db.QueryRow(`SELECT days.id FROM days WHERE days.name = '`+Day+`';`).Scan(&Did)
	var Fid int
	db.QueryRow(`SELECT faculties.id FROM faculties WHERE faculties.name = '`+Faculty+`';`).Scan(&Fid)
	insForm, err :=db.Prepare(`INSERT INTO schedule VALUES(nextval('schedule_id_seq'), $1,$2,$3,$4,$5)`)
	if err != nil {
		panic(err.Error())
	}
	insForm.Exec(Did, Iid, Fid, Sid,Tid)
	defer db.Close()
	return c.String(http.StatusOK, "Insert done!")
}

/*func Edit(c echo.Context) (err error) {
	db := dbConn()
	nId := c.QueryParam("id")
	a := c.QueryParam("faculty")
	query := `SELECT schedule.id, intervals.start, intervals.finish, subjects.name, teachers.name FROM schedule
			  INNER JOIN faculties ON facultyid = faculties.id
		 	  INNER JOIN subjects ON subjectid = subjects.id
	 		  INNER JOIN intervals ON intervalid = intervals.id
	   		  INNER JOIN teachers ON teacherid = teachers.id
			  INNER JOIN days ON dayid = days.id WHERE faculties.name = $1 AND days.name='Monday' AND schedule.id=$2`
	selDB, err := db.Query(query,a, nId)
	if err != nil {
		panic(err.Error())
	}
	sch := Schedule{}
	res := []Schedule{}
	for selDB.Next() {
		var Id int
		var Start, Finish, Subject, Teacher string
		err = selDB.Scan(&Id,&Start, &Finish, &Subject, &Teacher)
		if err != nil {
			panic(err.Error())
		}
		sch.Id = Id
		sch.Faculty = a
		sch.Time = Start +" — "+Finish
		sch.Subject = Subject
		sch.Teacher = Teacher
		res = append(res, sch)
	}
	tmpl.ExecuteTemplate(w, "Edit", emp)
	defer db.Close()
}
*/
