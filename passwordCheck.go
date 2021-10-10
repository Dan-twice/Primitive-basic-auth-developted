package main

import (
	"context"
	"crypto/sha256"
	_ "crypto/subtle"
	"fmt"
	"net/http"
)

type User struct {
	Name     string `json:"username"`
	Password string `json:"password"`
}

type ContextKey string

const ContextUserKey ContextKey = "user"

func checkNamePassword(user string, password string) bool {
	////////// smth better then hold in memory
	users := []User{}
	results, err := db.Queryx("SELECT username, password FROM users")
	if err != nil {
		panic(err.Error())
	}
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

			passwordMatch := false
			if passwordHash == v.Password {
				passwordMatch = true
			}
			return passwordMatch
		}
	}
	return false
}

func BasicAuth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		passwordMatch := checkNamePassword(username, password)
		fmt.Println("ok is", ok)
		if ok {
			if passwordMatch {
				fmt.Println("Password — matched for", username)
				// ctx, cancel := context.WithTimeout(r.Context(), time.Duration(60*time.Second))
				// defer cancel()
				ctx := context.WithValue(r.Context(), ContextUserKey, username)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}

		w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
		// StatusUnauthorized = 401
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	})
}
