package controller

import (
	"context"
	"github.com/umerm-work/arcTest/util"
	"log"
	"net/http"
	"strings"
	"time"
)

// Define our struct
type authenticationMiddleware struct {
	routes map[string]string
}

// Initialize it somewhere
func (amw *authenticationMiddleware) Populate() {
	amw.routes = make(map[string]string)
	amw.routes["/access-tokens/refresh"] = "/access-tokens/refresh"
	amw.routes["/access-tokens"] = "/access-tokens"
	amw.routes["/me"] = "/me"
	//amw.routes["users"] = "users"
	amw.routes["/ideas"] = "/ideas"

}

// Middleware function, which will be called for each request
func (amw *authenticationMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("X-Access-Token")
		amw.Populate()
		user, found := amw.routes[r.RequestURI]
		if r.RequestURI == "/users" || r.RequestURI == "/access-tokens/refresh" || (r.RequestURI == "/access-tokens" && r.Method == "POST") {
			next.ServeHTTP(w, r)
		} else if found || strings.Contains(r.RequestURI, "ideas") {
			value, err := util.ParseToken(token)
			if err != nil {
				log.Print("error ", err)
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
			timpstamp := int64(value["exp"].(float64))
			log.Printf("Time :%v data: %v Now: %v", timpstamp, time.Unix(timpstamp, 0), time.Now())
			if time.Unix(timpstamp, 0).Before(time.Now()) {
				http.Error(w, "token expired", http.StatusForbidden)
				return
			}
			ctx := r.Context()
			ctx = context.WithValue(ctx, "uid", value)
			log.Println("Value is : ", value)
			// We found the token in our map
			log.Printf("Authenticated user %s\n", user)
			// Pass down the request to the next middleware (or final handler)
			next.ServeHTTP(w, r)
		} else {
			log.Printf("Request URL %s\n", r.RequestURI)
			// Write an error and stop the handler chain
			http.Error(w, "Forbidden", http.StatusForbidden)
		}
	})
}
