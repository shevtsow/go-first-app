// https://youtu.be/IcEKRmzcXdU?list=PL0lO_mIqDDFXXqMzFOIGIb7FOprmUQ_tt

package main

import (
	"database/sql"
	"fmt"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"html/template"
	"net/http"
	"strings"
)

// Article это структура для хранения данных из формы в html
type Article struct {
	Id                     uint16
	Title, Anons, FullText string
}

// posts это список сохраняющий вышеуказанные структуры
var posts []Article // aka: var posts = []Article{}
// объект содержащий отображаемый пост
var showPost Article

// index метод который вызывается при открытии стартовой страницы, первоначально
// вызов происходит в handleFunc()
func index(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles(
		"templates/index.html",
		"templates/header.html",
		"templates/footer.html")
	if err != nil {
		fmt.Fprintf(w, err.Error()) // вывод ошибки на страницу
	}

	db := openSQLiteConn()
	defer db.Close()

	//--------------------------------------------------------------------------
	res, err := db.Query("SELECT * FROM article")
	if err != nil {
		panic(err)
	}

	posts = []Article{} // обнуление
	// добавление в список всех записей из базы в цыкле
	for res.Next() {
		var post Article
		err = res.Scan(&post.Id, &post.Title, &post.Anons, &post.FullText)
		if err != nil {
			panic(err)
		}
		posts = append(posts, post)
		//fmt.Println("Post Id:", post.Id, "Title -", post.Title, "Anons -", post.Anons)
	}
	//--------------------------------------------------------------------------

	// выполнить шаблон (с реализацией динамичного подключения)
	// здесь передача в ExecuteTemplate указанного списка, затем данные из этого
	// списка можно использовать в -> {{ . }}
	t.ExecuteTemplate(w, "index", posts) // см. index.html
}

func create(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles(
		"templates/create.html",
		"templates/header.html",
		"templates/footer.html")

	if err != nil {
		fmt.Fprintf(w, err.Error())
	}

	t.ExecuteTemplate(w, "create", nil)
}

func openSQLiteConn() *sql.DB {
	db, error := sql.Open("sqlite3", "store.db")
	if error != nil {
		panic(error)
	}
	return db
}

func insertToDb(db *sql.DB, title string, anons string, fullText string) {
	_, err := db.Exec("INSERT INTO article (title, anons, full_text) "+
		"VALUES ($1, $2, $3)", title, anons, fullText)

	if err != nil {
		panic(err)
	}
}

// в этом методе:
// получаются данные из формы в create.html
func save_article(w http.ResponseWriter, r *http.Request) {

	// 1. получаются данные из формы
	title := r.FormValue("title") // name="title" in form tag on create.html
	anons := r.FormValue("anons")
	full_text := r.FormValue("full_text")

	if len(strings.TrimSpace(title)) == 0 ||
		len(strings.TrimSpace(anons)) == 0 ||
		len(strings.TrimSpace(full_text)) == 0 {
		fmt.Fprintf(w, "Не все данные заполнены")
		return
	}

	// 2. открывается соединение с базой
	db := openSQLiteConn()
	defer db.Close()

	// 3. данные помещаются в базу
	insertToDb(db, title, anons, full_text)

	// 4. переадрисация:
	http.Redirect(w, r, "/", http.StatusSeeOther)

}

// отображение поста
func show_post(w http.ResponseWriter, r *http.Request) {
	//--------------------------------------------------------------------------
	vars := mux.Vars(r)
	w.WriteHeader(http.StatusOK)
	//--------------------------------------------------------------------------
	t, err := template.ParseFiles(
		"templates/show.html",
		"templates/header.html",
		"templates/footer.html")
	if err != nil {
		fmt.Fprintf(w, err.Error())
	}
	//--------------------------------------------------------------------------
	// fmt.Fprintf(w, "Id: %v\n", vars["id"]) // для теста
	db := openSQLiteConn()
	defer db.Close()
	//--------------------------------------------------------------------------
	query := fmt.Sprintf("SELECT * FROM article WHERE id = '%s'", vars["id"])
	res, err := db.Query(query)
	if err != nil {
		panic(err)
	}
	defer res.Close()
	//--------------------------------------------------------------------------
	if res.Next() {
		scnErr := res.Scan(&showPost.Id, &showPost.Title, &showPost.Anons, &showPost.FullText)
		if scnErr != nil {
			panic(scnErr)
		}
	} else {
		fmt.Fprintf(w, "Пост под номером: %v, не существует.", vars["id"])
		return
	}
	//--------------------------------------------------------------------------
	t.ExecuteTemplate(w, "show", showPost)
}

func handleFunc() {
	router := mux.NewRouter()

	//РОУТЕР.ОБРБ_ФУНКЦ(URL, НАЗВ_ФУНКЦ).МЕТОД_ПЕРЕДАЧИ_ДАННЫХ("GET")
	router.HandleFunc("/", index).Methods("GET")                     // ПОЛУЧЕНИЕ ИНФ. ПРОСТО ПУТЕМ ПЕРЕХОДА ПО ЭТОМУ АДРЕСУ
	router.HandleFunc("/create", create).Methods("GET")              // АНАЛОГИЧНО
	router.HandleFunc("/save_article", save_article).Methods("POST") // А ТУТ УЖЕ ПЕРЕДАЧА ДАННЫХ (А НЕ ПОЛУЧЕНИЕ) НА ЭТОТ АДРЕС
	router.HandleFunc("/post/{id:[0-9]+}", show_post).Methods("GET") // ТУТ ТОЖЕ ДАННЫЕ ПОЛУЧАЮТСЯ

	http.Handle("/", router) // УКАЗЫВАЕТ, ЧТО ОБРАБОТКА ВСЕХ URL ОСУЩЕСТВЛЯЕТСЯ ЧЕРЕЗ ОБЪЕКТ router As *Router

	handle := http.StripPrefix("/static/", http.FileServer(http.Dir("./static/")))
	http.Handle("/static/", handle)

	http.ListenAndServe(":8080", nil)

}

func main() {
	handleFunc()

}
