package controller

import (
	"context"
	"encoding/json"
	"github.com/gorilla/mux"
	uuid "github.com/satori/go.uuid"
	"github.com/umerm-work/arcTest/data"
	"github.com/umerm-work/arcTest/db"
	"github.com/umerm-work/arcTest/util"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

type App struct {
	Router *mux.Router
	DB     db.Repository
}

func (a *App) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, a.Router))
	//go func() {
	//	if err := http.ListenAndServe(addr, a.Router); err != nil {
	//		log.Println(err)
	//	}
	//}()
}

func (a *App) InitializeRoutes() {
	a.Router.HandleFunc("/access-tokens", a.Login).Methods("POST")
	a.Router.HandleFunc("/access-tokens/refresh", a.RefreshToken).Methods("POST")
	a.Router.HandleFunc("/access-tokens", a.Logout).Methods("DELETE")

	a.Router.HandleFunc("/ideas", a.CreateIdea).Methods("POST")
	a.Router.HandleFunc("/ideas", a.ListIdeas).Methods("GET")
	a.Router.HandleFunc("/ideas/{id}", a.DeleteIdea).Methods("DELETE")
	a.Router.HandleFunc("/ideas/{id}", a.UpdateIdea).Methods("PUT")

	a.Router.HandleFunc("/users", a.Signup).Methods("POST")

	a.Router.HandleFunc("/me", a.GetUser).Methods("GET")

	m := authenticationMiddleware{}
	a.Router.Use(m.Middleware)
}

func (a *App) Login(w http.ResponseWriter, r *http.Request) {
	var u data.User
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&u); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()
	if err := a.DB.Login(context.Background(), &u); err != nil {
		log.Printf("db error:%v", err)
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	td, err := util.CreateToken(u.ID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	u.RefreshToken = td.RefreshToken
	if err := a.DB.UpdateToken(context.Background(), &u); err != nil {
		log.Printf("db error:%v", err)
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	rs := make(map[string]string)
	rs["jwt"] = td.AccessToken
	rs["refresh_token"] = td.RefreshToken
	respondWithJSON(w, http.StatusCreated, rs)

}
func (a *App) Logout(w http.ResponseWriter, r *http.Request) {
	var u data.User
	token := r.Header.Get("X-Access-Token")
	value, err := util.ParseToken(token)
	if err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	u.ID = value["user_id"].(string)
	if err := a.DB.GetUser(context.Background(), &u); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	u.RefreshToken = "-"
	log.Printf("%v", u)
	if err := a.DB.UpdateToken(context.Background(), &u); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusNoContent, nil)

}
func (a *App) GetUser(w http.ResponseWriter, r *http.Request) {
	var u data.User
	token := r.Header.Get("X-Access-Token")
	value, err := util.ParseToken(token)
	if err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	u.ID = value["user_id"].(string)
	log.Printf("User ID : %v", u.ID)
	if err := a.DB.GetUser(context.Background(), &u); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	rs := make(map[string]string)
	rs["name"] = u.Name
	rs["email"] = u.Email
	rs["avatar_url"] = "https://www.gravatar.com/avatar/b36aafe03e05a85031fd8c411b69f792?d=mm&s=200"
	respondWithJSON(w, http.StatusOK, rs)

}
func (a *App) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var u data.User
	//Read all the data in r.Body from a byte[], convert it to a string, and assign store it in 's'.
	s, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err) // This would normally be a normal Error http response but I've put this here so it's easy for you to test.
	}
	var m map[string]interface{}
	// use the built in Unmarshal function to put the string we got above into the empty page we created at the top.  Notice the &p.  The & is important, if you don't understand it go and do the 'Tour of Go' again.
	err = json.Unmarshal(s, &m)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}
	token := m["refresh_token"].(string)
	value, err := util.ParseToken(token)
	if err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	u.ID = value["user_id"].(string)
	defer r.Body.Close()

	if err := a.DB.GetUser(context.Background(), &u); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	td, err := util.CreateToken(u.ID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	rs := make(map[string]string)
	rs["jwt"] = td.AccessToken
	respondWithJSON(w, http.StatusCreated, rs)

}
func (a *App) Signup(w http.ResponseWriter, r *http.Request) {
	var u data.User
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&u); err != nil {
		log.Printf("decode error:%v", err)
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	if err := u.Validate(); err != nil {
		log.Printf("validation error:%v", err)
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := a.DB.GetUserByEmail(context.Background(), &u); err == nil {
		log.Printf("validation error:%v", "email already exist")
		respondWithError(w, http.StatusBadRequest, "email already exist")
		return
	}
	u.ID = uuid.NewV4().String()
	td, err := util.CreateToken(u.ID)
	if err != nil {
		log.Printf("create token error:%v", err)
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	u.RefreshToken = td.RefreshToken
	if err := a.DB.CreateUser(context.Background(), u); err != nil {
		log.Printf("db error:%v", err)
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	rs := make(map[string]string)
	rs["jwt"] = td.AccessToken
	rs["access_token"] = td.RefreshToken
	respondWithJSON(w, http.StatusCreated, rs)
}

func (a *App) CreateIdea(w http.ResponseWriter, r *http.Request) {
	var i data.Idea
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&i); err != nil {
		log.Printf("decode error:%v", err)
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	if err := i.Validate(); err != nil {
		log.Printf("validation error:%v", err)
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	i.ID = uuid.NewV4().String()
	if err := a.DB.CreateIdea(context.Background(), i); err != nil {
		log.Printf("db error:%v", err)
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, i)
}
func (a *App) UpdateIdea(w http.ResponseWriter, r *http.Request) {
	param := mux.Vars(r)
	var i data.Idea
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&i); err != nil {
		log.Printf("decode error:%v", err)
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()
	i.ID = param["id"]

	if err := i.Validate(); err != nil {
		log.Printf("validation error:%v", err)
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := a.DB.UpdateIdea(context.Background(), i); err != nil {
		log.Printf("db error:%v", err)
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, i)
}

func (a *App) ListIdeas(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	param := r.URL.Query().Get("page")
	i, err := strconv.Atoi(param)
	if err != nil {
		log.Printf("parsing error:%v", err)
		respondWithError(w, http.StatusInternalServerError, "invalid query param")
	}
	ideas, err := a.DB.FindIdeas(context.Background(), int64(i))

	if err != nil {
		log.Printf("db error:%v", err)
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusCreated, ideas)
}
func (a *App) DeleteIdea(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	param := mux.Vars(r)

	if err := a.DB.DeleteIdea(context.Background(), param["id"]); err != nil {
		log.Printf("db error:%v", err)
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusNoContent, nil)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
