package main

import (
	"Personal-Web/connection"
	"Personal-Web/middleware"
	"context"
	"fmt"
	"html/template"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

type MetaData struct {
	Title     string
	IsLogin   bool
	Username  string
	FlashData string
	Email     string
	Id        int
}

var Data = MetaData{
	Title: "Personal Web",
}

type User struct {
	Id       int
	Name     string
	Email    string
	Password string
}

type Project struct {
	Id                int
	Project_name      string
	Start_date        time.Time
	End_date          time.Time
	Description       string
	Technologies      []string
	Duration          string
	Author_id         int
	Author_name       string
	Start_date_format string
	End_date_format   string
	Image             string
}

var Projects []Project
var ProjectDetail = Project{}
var store = sessions.NewCookieStore([]byte("SESSION_ID"))

func main() {

	router := mux.NewRouter()

	router.PathPrefix("/public").Handler(http.StripPrefix("/public/", http.FileServer(http.Dir("./public"))))
	router.PathPrefix("/uploads").Handler(http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))
	connection.DatabaseConnect()
	router.HandleFunc("/", home).Methods("GET")
	router.HandleFunc("/add-project", projectForm).Methods("GET")
	router.HandleFunc("/contact-me", contactMe).Methods("GET")
	router.HandleFunc("/detail-project", detailProject).Methods("GET")
	router.HandleFunc("/get-data-project", middleware.UploadFile(getDataProject)).Methods("POST")
	router.HandleFunc("/delete-project/{id}", deleteProject).Methods("GET")
	router.HandleFunc("/detail-project/{id}", detailProject).Methods("GET")
	router.HandleFunc("/show-data-project/{id}", showDataProject).Methods("GET")
	router.HandleFunc("/update-project/{id}", middleware.UploadFile(updateProject)).Methods("POST")
	router.HandleFunc("/form-register", formRegister).Methods("GET")
	router.HandleFunc("/register", register).Methods("POST")
	router.HandleFunc("/form-login", formLogin).Methods("GET")
	router.HandleFunc("/login", login).Methods("POST")
	router.HandleFunc("/logout", logout).Methods("GET")
	router.HandleFunc("/update-password", updatePassword).Methods("POST")

	fmt.Println("Server running smoothly on port 5000")
	http.ListenAndServe("localhost:5000", router)

}

func home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/html; charset=utf-8")

	var tmpl, err = template.ParseFiles("views/index.html")

	ProjectDetail = Project{}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))

	}

	session, _ := store.Get(r, "SESSION_ID")

	if session.Values["IsLogin"] != true {
		Data.IsLogin = false
	} else {
		Data.IsLogin = session.Values["IsLogin"].(bool)
		Data.Username = session.Values["Name"].(string)
		Data.Id = session.Values["Id"].(int)
		fetchDataProjects(Data.Id)
	}

	fm := session.Flashes("message")

	var flashes []string

	if len(fm) > 0 {
		session.Save(r, w)

		for _, fl := range fm {
			flashes = append(flashes, fl.(string))
		}
	}

	Data.FlashData = strings.Join(flashes, "")

	resp := map[string]interface{}{
		"Data":     Data,
		"Projects": Projects,
	}

	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, resp)
}

func updatePassword(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	id := Data.Id
	password := r.PostForm.Get("password")

	passwordHash, _ := bcrypt.GenerateFromPassword([]byte(password), 10)
	_, err = connection.Conn.Exec(context.Background(), "UPDATE public.tb_user set password = $1 where id = $2", passwordHash, id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Message : " + err.Error()))
		return
	}

	http.Redirect(w, r, "/form-login", http.StatusMovedPermanently)
}

func register(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	name := r.PostForm.Get("name")
	email := r.PostForm.Get("email")
	password := r.PostForm.Get("password")

	passwordHash, _ := bcrypt.GenerateFromPassword([]byte(password), 10)

	_, err = connection.Conn.Exec(context.Background(), "INSERT INTO public.tb_user(name, email, password) VALUES ($1, $2, $3);", name, email, passwordHash)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Message : " + err.Error()))
		return
	}

	http.Redirect(w, r, "/form-login", http.StatusMovedPermanently)
}
func logout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache,no-store,must-revalidate")
	Projects = nil
	session, _ := store.Get(r, "SESSION_ID")

	session.Values["IsLogin"] = false
	session.Options.MaxAge = -1
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusMovedPermanently)

}

func login(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	session, _ := store.Get(r, "SESSION_ID")

	email := r.PostForm.Get("email")
	password := r.PostForm.Get("password")

	var each = User{}
	err = connection.Conn.QueryRow(context.Background(), "SELECT Id, email, name, password FROM tb_user WHERE email=$1", email).Scan(&each.Id, &each.Email, &each.Name, &each.Password)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Message : " + err.Error()))
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(each.Password), []byte(password))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Message : " + err.Error()))
		return
	}

	session.Values["IsLogin"] = true
	session.Values["Name"] = each.Name
	session.Values["Email"] = each.Email
	session.Values["Id"] = each.Id
	session.Options.MaxAge = 10800

	session.AddFlash("Login success", "message")
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

func formRegister(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/html; charset=utf-8")

	var tmpl, err = template.ParseFiles("views/register.html")

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))

	}

	session, _ := store.Get(r, "SESSION_ID")

	if session.Values["IsLogin"] != true {
		Data.IsLogin = false
	} else {
		Data.IsLogin = session.Values["IsLogin"].(bool)
		Data.Username = session.Values["Name"].(string)
		Data.Email = session.Values["Email"].(string)
		Data.Id = session.Values["Id"].(int)
	}

	resp := map[string]interface{}{
		"Data": Data,
	}

	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, resp)
}

func formLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/html; charset=utf-8")

	session, _ := store.Get(r, "SESSION_ID")
	session.Options.MaxAge = -1

	var tmpl, err = template.ParseFiles("views/login.html")

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))

	}

	resp := map[string]interface{}{
		"Data": Data,
	}

	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, resp)
}

func detailProject(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/html; charset=utf-8")
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	fetchDataProjectDetail(id)
	var tmpl, err = template.ParseFiles("views/detail-project.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + (err.Error())))
		fmt.Println(err.Error())
		return
	}

	strFormatDate := "2006-Jan-02"
	ProjectDetail.Start_date_format = ProjectDetail.Start_date.Format(strFormatDate)
	ProjectDetail.End_date_format = ProjectDetail.End_date.Format(strFormatDate)
	resp := map[string]interface{}{
		"Data":          Data,
		"ProjectDetail": ProjectDetail,
	}
	fmt.Println(resp["ProjectDetail"])
	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, resp)
}

func showDataProject(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/html; charset=utf-8")
	var tmpl, err = template.ParseFiles("views/my-project.html")
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	fetchDataProjectDetail(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + (err.Error())))
		fmt.Println(err.Error())
		return
	}

	strFormatDate := "2006-01-02"
	ProjectDetail.Start_date_format = ProjectDetail.Start_date.Format(strFormatDate)
	ProjectDetail.End_date_format = ProjectDetail.End_date.Format(strFormatDate)
	resp := map[string]interface{}{
		"Data":          Data,
		"ProjectDetail": ProjectDetail,
	}

	w.WriteHeader(http.StatusOK)

	tmpl.Execute(w, resp)

}

func contactMe(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/html; charset=utf-8")
	var tmpl, err = template.ParseFiles("views/contact-me.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, Data)
}

func updateProject(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-type", "text/html; charset=utf-8")

	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
		fmt.Println(err.Error())
		return
	}
	id := ProjectDetail.Id
	dataContext := r.Context().Value("dataFile")
	projectName := r.PostForm.Get("projectName")
	startDate := r.PostForm.Get("startDate")
	endDate := r.PostForm.Get("endDate")
	description := r.PostForm.Get("description")
	technologies := r.Form["technologies"]
	image := dataContext.(string)
	_, err = connection.Conn.Exec(context.Background(),
		"UPDATE public.tb_projects SET project_name = $1, start_date = $2, end_date = $3, description = $4, technologies = $5,image_project = $7 where id = $6", projectName, startDate, endDate, description, technologies, id, image)
	if err != nil {
		fmt.Println("Message : " + err.Error())
		return
	}

	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

func deleteProject(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-cache,no-store,must-revalidate")
	w.Header().Set("Content-type", "text/html; charset=utf-8")

	id := ProjectDetail.Id
	fmt.Println(id)

	_, err := connection.Conn.Exec(context.Background(),
		"DELETE FROM public.tb_projects where id = $1", id)
	if err != nil {
		fmt.Println("Message : " + err.Error())
		return
	}

	http.Redirect(w, r, "/", http.StatusMovedPermanently)

}

func projectForm(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/html; charset=utf-8")
	var tmpl, err = template.ParseFiles("views/my-project.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}

	resp := map[string]interface{}{
		"Data":          Data,
		"ProjectDetail": ProjectDetail,
	}
	ProjectDetail = Project{}
	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, resp)
}

func getDataProject(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-cache,no-store,must-revalidate")
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}
	session, _ := store.Get(r, "SESSION_ID")
	if session.Values["IsLogin"] != true {
		Data.IsLogin = false
	} else {
		Data.Id = session.Values["Id"].(int)
	}
	dataContext := r.Context().Value("dataFile")

	projectName := r.PostForm.Get("projectName")
	strLayoutFormat := "2006-01-02"
	startDate, _ := time.Parse(strLayoutFormat, r.PostForm.Get("startDate"))
	endDate, _ := time.Parse(strLayoutFormat, r.PostForm.Get("endDate"))
	description := r.PostForm.Get("description")
	technologies := r.Form["technologies"]
	authorId := Data.Id
	image := dataContext.(string)
	fmt.Println(startDate)
	fmt.Println(endDate)
	_, err = connection.Conn.Exec(context.Background(),
		"INSERT INTO public.tb_projects(project_name,start_date,end_date,description,technologies,author_id,image_project) VALUES($1, $2, $3, $4, $5, $6,$7)", projectName, startDate, endDate, description, technologies, authorId, image)
	if err != nil {
		fmt.Println("Message : " + err.Error())
		return
	}

	http.Redirect(w, r, "/", http.StatusMovedPermanently)

}

func countDuration(startDate time.Time, endDate time.Time) string {
	var timeDuration string

	distance := endDate.Sub(startDate).Hours()
	dayDistance := math.Floor(distance / 24)
	monthDistance := math.Floor(dayDistance / 30)
	yearDistance := math.Floor(monthDistance / 12)

	strDayDistance := strconv.FormatFloat(dayDistance, 'f', 0, 64)
	strMonthDistance := strconv.FormatFloat(monthDistance, 'f', 0, 64)
	strYearDistance := strconv.FormatFloat(yearDistance, 'f', 0, 64)

	if yearDistance >= 1 {
		timeDuration = strYearDistance + " Year "
	} else if monthDistance >= 1 {
		timeDuration = strMonthDistance + " Month "
	} else if dayDistance >= 1 {
		timeDuration = strDayDistance + " Day"
	} else {
		timeDuration = "No Duration"
	}

	return timeDuration

}

func fetchDataProjects(authorId int) {
	Projects = nil

	rows, _ := connection.Conn.Query(context.Background(), `SELECT t1.id, project_name, start_date, end_date, description, technologies,author_id , t2.name author_name,image_project
	FROM public.tb_projects t1 join tb_user t2 on
	t1.author_id = t2.id
	where t2.id = $1`, authorId)
	for rows.Next() {
		var each = Project{}
		var err = rows.Scan(&each.Id, &each.Project_name, &each.Start_date, &each.End_date, &each.Description, &each.Technologies, &each.Author_id, &each.Author_name, &each.Image)

		if err != nil {
			fmt.Println(err.Error())
			return
		}

		each.Duration = countDuration(each.Start_date, each.End_date)
		Projects = append(Projects, each)
	}
}

func fetchDataProjectDetail(id int) {

	err := connection.Conn.QueryRow(context.Background(), `SELECT t1.id, project_name, start_date, end_date, description, technologies,author_id , t2.name author_name,image_project
	FROM public.tb_projects t1 join tb_user t2 on
	t1.author_id = t2.id where t1.id = $1`, id).
		Scan(
			&ProjectDetail.Id,
			&ProjectDetail.Project_name,
			&ProjectDetail.Start_date,
			&ProjectDetail.End_date,
			&ProjectDetail.Description,
			&ProjectDetail.Technologies,
			&ProjectDetail.Author_id,
			&ProjectDetail.Author_name,
			&ProjectDetail.Image)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	ProjectDetail.Duration = countDuration(ProjectDetail.Start_date, ProjectDetail.End_date)

}
