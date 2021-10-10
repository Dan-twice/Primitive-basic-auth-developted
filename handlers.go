package main

import (
	"crypto/sha256"
	_ "crypto/subtle"
	"fmt"
	"html/template"
	"log"
	"net/http"
)

type User2 struct {
	Name string `json:"username"`
}

func LogOut(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusUnauthorized)
	w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
	tmpl, _ := template.New("").ParseFiles("logout.html")
	err := tmpl.ExecuteTemplate(w, "logout.html", struct{}{})
	if err != nil {
		panic(err)
	}
}

func RegistrationHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Registar")
	// workdir, _ := os.Getwd()
	regPath := "registration.html"
	tmpl, _ := template.New("").ParseFiles(regPath)
	err := tmpl.ExecuteTemplate(w, regPath, struct{}{})
	if err != nil {
		log.Println("Html file problems")
		log.Println(err)
		panic(err)
	}
}

func PostHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("r.Method —", r.Method, " http.MethodPost —", http.MethodPost)
	if r.Method == http.MethodPost {
		nickNote, nameNote, birthNote := r.FormValue("nick"), r.FormValue("username"), r.FormValue("birth")
		sha := sha256.New()
		sha.Write([]byte(r.FormValue("password")))
		passwordNote := fmt.Sprintf("%x", sha.Sum(nil))
		res := fmt.Sprintf(`INSERT INTO users (username, password, birth_date, full_name)
							VALUES('%s', '%s', '%s', '%s');`, nickNote, passwordNote, birthNote, nameNote)
		_, err := db.Exec(res)
		if err != nil {
			log.Println("http.MethodPost Problem")
			panic(err)
		}
		log.Println("Registration complete for ", nickNote)

		// StatusSeeOther = 302 — redirect as a result of POST
		http.Redirect(w, r, "/protected", http.StatusSeeOther)
	}
}

func Unprotected(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Works"))
}

func ProtectedHandler(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(ContextUserKey)
	tmpl, _ := template.New("").ParseFiles("greetings.html")
	err := tmpl.ExecuteTemplate(w, "greetings.html",
		struct {
			Username string
		}{
			user.(string),
		})
	if err != nil {
		panic(err)
	}
	fmt.Fprintln(w, "This is the protected handler")
}

func ShowUsers(w http.ResponseWriter, r *http.Request) {
	results, err := db.Query("SELECT username FROM users")
	if err != nil {
		panic(err.Error())
	}

	userss := []User2{}
	for results.Next() {
		var user User2
		results.Scan(&user.Name)
		userss = append(userss, user)
	}

	tmpl, _ := template.New("").ParseFiles("show_users.html")
	err = tmpl.ExecuteTemplate(w, "show_users.html",
		struct {
			Users []User2
		}{
			userss,
		})
	if err != nil {
		panic(err)
	}
}
