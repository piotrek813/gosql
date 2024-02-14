package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx"
	"github.com/joho/godotenv"
)

func init() {

	godotenv.Load(".env")
}

type User struct {
	ID       int    `json:"user_id,omitempty" db:"user_id"`
	Username string `json:"username,omitempty" db:"username"`
	Email    string `json:"email,omitempty" db:"email"`
	Password string `json:"password,omitempty" db:"password"`
	// CreatedAt time.Time `json:"created_at" db:"created_at"`
}

func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.URL)
		next.ServeHTTP(w, r)
	})
}

func redirectSlash(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rp := r.URL.JoinPath().String()
		if rp[len(rp)-1] == '/' {
			log.Printf("redirecting: %v\n", rp)
			http.Redirect(w, r, rp[:len(rp)-1], http.StatusMovedPermanently)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	cc := pgx.ConnConfig{
		Host:     "localhost",
		User:     "dupa",
		Password: "dupa",
		Database: "dupa",
		Port:     5432,
	}
	conn, err := pgx.Connect(cc)

	defer conn.Close()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(0)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("POST /user/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		fmt.Printf("id: %v\n", id)

		var user User
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&user); err != nil {
			http.Error(w, "Bad request", http.StatusBadGateway)
			return
		}

		_, err := conn.Exec("INSERT INTO users (user_id, username, password, email) VALUES ($1, $2, $3, $4);", id, user.Username, user.Password, user.Email)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("err: %v\n", err)
			return
		}

		w.Header().Add("content-type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(user)
	})

	mux.HandleFunc("GET /user/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		var user User

		err := conn.QueryRow("SELECT user_id,username from users WHERE user_id = $1", id).Scan(&user.ID, &user.Username)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, err)
			return
		}
		fmt.Printf("user: %v\n", user)

		w.Header().Add("content-type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(user)
	})

	http.ListenAndServe(":8080", redirectSlash(logging(mux)))

	// var user User
	// userType := reflect.TypeOf(user)

	// for row.Next() {
	// 	for range row.FieldDescriptions() {
	// 		var user User
	// 		userType := reflect.TypeOf(user)
	// 		if val, ok := userType.FieldByName("Username"); ok == true {
	// 			row.Scan(userType.FieldByIndex(0))
	// 			fmt.Printf("val: %v\n", val)
	// 		}
	// 		fmt.Printf("user: %v\n", user)
	// 	}
	// }
}
