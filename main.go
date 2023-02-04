/*
package setiap program harus ada minimal satu package, dan itu package main
isi dari package bisa berbagai macam fungsi
*/
package main

/*import() digunakan untuk memasukan file package lain ke program
untuk memanfaatkan isi package yang di import
contoh nya untuk import fungsi yang ada di package lain
*/
import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

var Data = map[string]interface{}{
	"Title": "Personal Web",
}

/*
fungsi main() adalah fungsi yang pertama kali di panggil saat program running atau
di eksekusi
*/
func main() {
	//handler itu adalah alat komunikasi antara server dan client
	//deklarasi new router
	router := mux.NewRouter()

	//create static folder
	router.PathPrefix("/public").Handler(http.StripPrefix("/public", http.FileServer(http.Dir("./public"))))

	http.Handle("/public", http.StripPrefix("/public", http.FileServer(http.Dir("public"))))

	// create handling URL
	/*
		router.HandleFunc(endpoint, function)
		fungsi ini digunakan untuk menghandle sebuah function yang akan di jalankan
		di web server

		jadi untuk pemanggilan HandleFunc,
		di url browser cukup pakai localhost:port/endPoint
	*/
	router.HandleFunc("/hello", helloWorld).Methods("GET")
	router.HandleFunc("/", home).Methods("GET")
	router.HandleFunc("/add-project", projectForm).Methods("GET")
	router.HandleFunc("/contact-me", contactMe).Methods("GET")
	router.HandleFunc("/getDataProject", getDataProject).Methods("POST")
	/*
		fmt.Println berfungsi sebagai console log. fungsi ini di import dari "fmt"
	*/
	fmt.Println("Server running smoothly on port 5000")
	http.ListenAndServe("localhost:5000", router)
}

func helloWorld(w http.ResponseWriter, r *http.Request) {
	/*
		parameter ke dua itu adalah tipe data yang di parsing
	*/
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	/*
		Memakai []byte untuk digunakan enkripsi data string menjadi byte
		untuk di tampilkan di browser dengan web server
	*/
	w.Write([]byte("Hello World"))
}

func home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/html; charset=utf-8")

	//parsing template html
	var tmpl, err = template.ParseFiles("views/index.html")

	//error handling
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
	}
	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, Data)
}

// function handling form add-project.html
func projectForm(w http.ResponseWriter, r *http.Request) {
	/*
		parameter ke dua itu adalah tipe data yang di parsing
	*/
	w.Header().Set("Content-type", "text/html; charset=utf-8")

	//parsing template html
	//untuk parsing file yang ada di local server kita
	var tmpl, err = template.ParseFiles("views/my-project.html")

	//error handling
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, Data)
}

// function handling form contact-me.html
func contactMe(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/html; charset=utf-8")

	//parsing template html
	var tmpl, err = template.ParseFiles("views/contact-me.html")

	//error handling
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, Data)
}

func getDataProject(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	projectName := r.PostForm.Get("projectName")
	startDate := r.PostForm.Get("startDate")
	endDate := r.PostForm.Get("endDate")
	description := r.PostForm.Get("description")
	nodeJs := r.PostForm.Get("nodeJs")
	vueJs := r.PostForm.Get("vueJs")
	reactJs := r.PostForm.Get("reactJs")
	javaScript := r.PostForm.Get("javaScript")

	fmt.Println("Project Name: " + projectName)
	fmt.Println("Start Date: " + startDate)
	fmt.Println("End Date: " + endDate)
	fmt.Println("Description: " + description)
	fmt.Println("Node JS: " + nodeJs)
	fmt.Println("Vue JS: " + vueJs)
	fmt.Println("React JS: " + reactJs)
	fmt.Println("JavaScript: " + javaScript)
	http.Redirect(w, r, "/add-project", http.StatusMovedPermanently)
}
