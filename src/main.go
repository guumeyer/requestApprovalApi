//mongodb://localhost:27017/foodTracker
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"goji.io"
	"goji.io/pat"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func ErrorWithJSON(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	fmt.Fprintf(w, "{message: %q}", message)
}

func ResponseWithJSON(w http.ResponseWriter, json []byte, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	w.Write(json)
}

type Request struct {
  Id      bson.ObjectId `json:"id" bson:"_id,omitempty"`
	Title   string   `json:"title"`
  Description   string   `json:"description"`
}

func main() {
	session, err := mgo.Dial("<url mongo database>")
	if err != nil {
		panic(err)
	}
	defer session.Close()

	session.SetMode(mgo.Monotonic, true)
	//ensureIndex(session)

	mux := goji.NewMux()
	mux.HandleFunc(pat.Get("/request"), allRequest(session))
	mux.HandleFunc(pat.Post("/request"), addRequest(session))
  mux.HandleFunc(pat.Get("/request/:id"), RequestById(session))
  mux.HandleFunc(pat.Put("/request/:id"), updateRequest(session))
  mux.HandleFunc(pat.Delete("/request/:title"), deleteRequest(session))
	http.ListenAndServe(":8080", mux)
}


func allRequest(s *mgo.Session) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		session := s.Copy()
		defer session.Close()

		c := session.DB("requestsapproval").C("requests")

		var request []Request
		err := c.Find(bson.M{}).All(&request)
		if err != nil {
			ErrorWithJSON(w, "Database error", http.StatusInternalServerError)
			log.Println("Failed get all request: ", err)
			return
		}

		respBody, err := json.MarshalIndent(request, "", "  ")
		
		defer session.Close()

		if err != nil {
			log.Fatal(err)
		}
		ResponseWithJSON(w, respBody, http.StatusOK)
	}
}

func addRequest(s *mgo.Session) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		session := s.Copy()
		defer session.Close()

		var request Request
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&request)
		if err != nil {
			ErrorWithJSON(w, "Incorrect body", http.StatusBadRequest)
			return
		}

		c := session.DB("requestsapproval").C("requests")

		err = c.Insert(request)

		defer session.Close()

		if err != nil {
			if mgo.IsDup(err) {
				ErrorWithJSON(w, "Request with this ISBN already exists", http.StatusBadRequest)
				return
			}

			ErrorWithJSON(w, "Database error", http.StatusInternalServerError)
			log.Println("Failed insert Request: ", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		// w.Header().Set("Location", r.URL.Path+"/"+Request.ISBN)
		w.WriteHeader(http.StatusCreated)
	}
}
//
// func ensureIndex(s *mgo.Session) {
// 	session := s.Copy()
// 	defer session.Close()
//
// 	c := session.DB("requestsapproval").C("requests")
//
// 	index := mgo.Index{
// 		Key:        []string{"title"},
// 		Unique:     true,
// 		DropDups:   true,
// 		Background: true,
// 		Sparse:     true,
// 	}
// 	err := c.EnsureIndex(index)
// 	if err != nil {
// 		panic(err)
// 	}
// }
//
//
func RequestById(s *mgo.Session) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

    log.Println("RequestById")
		session := s.Copy()
		defer session.Close()

		requestId := pat.Param(r, "id")
    log.Println(requestId)

		c := session.DB("requestsapproval").C("requests")

		var request Request
		err := c.Find(bson.M{"_id": bson.ObjectIdHex(requestId)}).One(&request)

		defer session.Close()

		if err != nil {
			ErrorWithJSON(w, "Database error", http.StatusInternalServerError)
			log.Println("Failed find Request: ", err)
			return
		}

		if request.Title == "" {
			ErrorWithJSON(w, "Request not found", http.StatusNotFound)
			return
		}

		respBody, err := json.MarshalIndent(request, "", "  ")
		if err != nil {
			log.Fatal(err)
		}

		ResponseWithJSON(w, respBody, http.StatusOK)
	}
}
//
func updateRequest(s *mgo.Session) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		session := s.Copy()
		defer session.Close()

		requestId := pat.Param(r, "id")

		var request Request
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&request)
		if err != nil {
			ErrorWithJSON(w, "Incorrect body", http.StatusBadRequest)
			defer session.Close()
			return
		}

		c := session.DB("requestsapproval").C("requests")

		err = c.Update(bson.M{"_id": bson.ObjectIdHex(requestId)}, &request)
		defer session.Close()
		if err != nil {
			switch err {
			default:
				ErrorWithJSON(w, "Database error", http.StatusInternalServerError)
				log.Println("Failed update Request: ", err)
				return
			case mgo.ErrNotFound:
				ErrorWithJSON(w, "Request not found", http.StatusNotFound)
				return
			}
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func deleteRequest(s *mgo.Session) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		session := s.Copy()
		defer session.Close()

		requestId := pat.Param(r, "id")

		c := session.DB("requestsapproval").C("requests")

		err := c.Remove(bson.M{"_id": bson.ObjectIdHex(requestId)})
		defer session.Close()
		if err != nil {
			switch err {
			default:
				ErrorWithJSON(w, "Database error", http.StatusInternalServerError)
				log.Println("Failed delete Request: ", err)
				return
			case mgo.ErrNotFound:
				ErrorWithJSON(w, "Request not found", http.StatusNotFound)
				return
			}
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
