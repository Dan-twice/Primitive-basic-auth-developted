package main

import (
	"crypto/sha256"
	_ "crypto/subtle"
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/cockroachdb/cockroach-go/v2/crdb"
	_ "github.com/lib/pq"
)

type application struct {
	auth struct {
		username string
	}
}

type User struct {
	Name     string `json:"username"`
	Password string `json:"password"`
}

type User2 struct {
	Name string `json:"username"`
}

var db *sql.DB
var err error

func checkNamePassword(user string, password string) bool {
	////////// smth better then hold in memory
	results, err := db.Query("SELECT username, password FROM users")
	if err != nil {
		panic(err.Error())
	}
	users := []User{}
	for results.Next() {
		var user User
		results.Scan(&user.Name, &user.Password)
		users = append(users, user)
	}
	for _, v := range users {
		if user == v.Name {
			sha := sha256.New()
			sha.Write([]byte(password))
			passwordHash := fmt.Sprintf("%x", sha.Sum(nil))

			// alternative
			// passwordHash := sha256.Sum256([]byte(password))
			// passwordMatch := (subtle.ConstantTimeCompare(passwordHash[:], expectedPasswordHash[:]) == 1)

			passwordMatch := false
			if passwordHash == v.Password {
				passwordMatch = true
			}
			return passwordMatch
		}
	}
	return false
}

func (app *application) logOut(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusUnauthorized)
	w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
	tmpl, _ := template.New("").ParseFiles("logout.html")
	err := tmpl.ExecuteTemplate(w, "logout.html", struct{}{})
	if err != nil {
		panic(err)
	}
	// http.Error(w, "Unauthorized", http.StatusUnauthorized)
}

func (app *application) registrationHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.New("").ParseFiles("registration.html")
	err := tmpl.ExecuteTemplate(w, "registration.html", struct{}{})
	if err != nil {
		panic(err)
	}

	if r.Method == http.MethodPost {
		nickNote, nameNote, birthNote := r.FormValue("nick"), r.FormValue("username"), r.FormValue("birth")
		sha := sha256.New()
		sha.Write([]byte(r.FormValue("password")))
		passwordNote := fmt.Sprintf("%x", sha.Sum(nil))
		res := fmt.Sprintf(`INSERT INTO users (username, password, birth_date, full_name)
							VALUES('%s', '%s', '%s', '%s');`, nickNote, passwordNote, birthNote, nameNote)
		_, err := db.Exec(res)
		if err != nil {
			panic(err)
		}
		log.Println("Registration complete for ", nickNote)
		// StatusSeeOther = 302 — redirect as a result of POST
		// http.Redirect(w, r, "/protected", http.StatusSeeOther)
	}
}

// Killing session is not enough, since, once user is authenticated, each request contains login info, so user is automatically logged in
// next time he/she access the site using the same credentials.
// The only solution so far is to close browser, but that's not acceptable from the usability standpoint.
func (app *application) basicAuth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		passwordMatch := checkNamePassword(username, password)
		fmt.Println("ok is ", ok)
		if ok {
			if passwordMatch {
				app.auth.username = username
				fmt.Println("Password — matched", app.auth.username)
				next.ServeHTTP(w, r)
				return
			}
		}

		w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
		// StatusUnauthorized = 401
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	})
}

func (app *application) showUsers(w http.ResponseWriter, r *http.Request) {
	results, err := db.Query("SELECT username FROM users")
	if err != nil {
		panic(err.Error())
	}
	tmpl, err := template.New("").ParseFiles("show_users.html")
	userss := []User2{}
	for results.Next() {
		var user User2
		results.Scan(&user.Name)
		userss = append(userss, user)
	}

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

func (app *application) protectedHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.New("").ParseFiles("greetings.html")
	err := tmpl.ExecuteTemplate(w, "greetings.html",
		struct {
			Username string
		}{
			app.auth.username,
		})
	if err != nil {
		panic(err)
	}
	fmt.Fprintln(w, "This is the protected handler")
}

func (app *application) unprotectedHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "This is the unprotected handler")
}

func main() {
	app := new(application)

	pgConnString := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=disable",
		os.Getenv("PGHOST"),
		os.Getenv("PGPORT"),
		os.Getenv("PGDATABASE"),
		os.Getenv("PGUSER"),
		os.Getenv("PGPASSWORD"),
	)
	db, err = sql.Open("postgres", pgConnString) // init db
	if err != nil {
		log.Println("Open")
		log.Fatal(err)
	}

	// retry logic
	retries := 5
	for retries > 0 {
		pingErr := db.Ping() // really connect to database
		if pingErr != nil {
			retries -= 1
			fmt.Println("Retries left ", retries)
			log.Println(pingErr)
			time.Sleep(time.Second)
		} else {
			log.Printf("Left Ping")
			break
		}
	}

	// fast router without parsing
	mux := http.NewServeMux()
	mux.HandleFunc("/protected", app.basicAuth(app.protectedHandler))
	mux.HandleFunc("/", app.registrationHandler)
	mux.HandleFunc("/unprotected", app.unprotectedHandler)
	mux.HandleFunc("/show", app.showUsers)
	mux.HandleFunc("/log", app.logOut)

	srv := &http.Server{
		Addr:         ":4000",
		Handler:      mux,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// Use https with «SSL certificate». mkcert localhost creates these files
	err = srv.ListenAndServeTLS("./localhost.pem", "./localhost-key.pem")
	log.Printf("starting server on %s", srv.Addr)
}

// docker tag local-image:tagname new-repo:tagname
// docker push new-repo:tagname

// func basicAuth(next http.HandlerFunc) http.HandlerFunc {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//         // Extract the username and password from the request
//         // Authorization header. If no Authentication header is present
//         // or the header value is invalid, then the 'ok' return value
//         // will be false.
// 		username, password, ok := r.BasicAuth()
// 		if ok {
//             // Calculate SHA-256 hashes for the provided and expected
//             // usernames and passwords.
// 			usernameHash := sha256.Sum256([]byte(username))
// 			passwordHash := sha256.Sum256([]byte(password))
// 			expectedUsernameHash := sha256.Sum256([]byte("your expected username"))
// 			expectedPasswordHash := sha256.Sum256([]byte("your expected password"))

//             // Use the subtle.ConstantTimeCompare() function to check if
//             // the provided username and password hashes equal the
//             // expected username and password hashes. ConstantTimeCompare
//             // will return 1 if the values are equal, or 0 otherwise.
//             // Importantly, we should to do the work to evaluate both the
//             // username and password before checking the return values to
//             // avoid leaking information.
// 			usernameMatch := (subtle.ConstantTimeCompare(usernameHash[:], expectedUsernameHash[:]) == 1)
// 			passwordMatch := (subtle.ConstantTimeCompare(passwordHash[:], expectedPasswordHash[:]) == 1)

//             // If the username and password are correct, then call
//             // the next handler in the chain. Make sure to return
//             // afterwards, so that none of the code below is run.
// 			if usernameMatch && passwordMatch {
// 				next.ServeHTTP(w, r)
// 				return
// 			}
// 		}

//         // If the Authentication header is not present, is invalid, or the
//         // username or password is wrong, then set a WWW-Authenticate
//         // header to inform the client that we expect them to use basic
//         // authentication and send a 401 Unauthorized response.
// 		// Basic realm — the realm value is a string which allows you
// 		// to create partitions of protected space in your application.
// 		//  So, for example, an application could have a "documents" realm
// 		// and an "admin area" realm, which require different credentials.
// 		// A web browser (or other type of client) can cache and automatically
// 		// reuse the same username and password for any requests within the same
// 		//  realm, so that the prompt doesn't need to be shown for every single request.
// 		w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
// 		http.Error(w, "Unauthorized", http.StatusUnauthorized)
// 	})
// }
