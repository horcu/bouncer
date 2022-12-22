package main

import (
	"context"
	"encoding/json"
	firebase "firebase.google.com/go"
	"firebase.google.com/go/db"
	"fmt"
	models "github.com/horcu/peez_me_models"
	"html/template"
	"log"
	"net/http"
	"os"
)

// templateData provides template parameters.
type templateData struct {
	Service  string
	Revision string
}

// Variables used to generate the HTML page.
var (
	data     templateData
	lobbyRef *db.Ref
	tmpl     *template.Template
)

func main() {
	setup()

	// Initialize template parameters.
	service := os.Getenv("K_SERVICE")
	if service == "" {
		service = "???"
	}

	revision := os.Getenv("K_REVISION")
	if revision == "" {
		revision = "???"
	}

	// Prepare template for execution.
	tmpl = template.Must(template.ParseFiles("index.html"))
	data = templateData{
		Service:  service,
		Revision: revision,
	}

	// Define HTTP server.
	mux := http.NewServeMux()
	mux.HandleFunc("/", defHandler)
	mux.HandleFunc("/join", joinHandler)
	mux.HandleFunc("/leave", leaveHandler)

	fs := http.FileServer(http.Dir("./assets"))
	http.Handle("/assets/", http.StripPrefix("/assets/", fs))

	// PORT environment variable is provided by Cloud Run.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Print("Hello from Cloud Run! The container started successfully and is listening for HTTP requests on $PORT")
	log.Printf("Listening on port %s", port)
	err := http.ListenAndServe(":"+port, mux)
	if err != nil {
		log.Fatal(err)
	}
}

func setup() {

	// set context
	_ = context.Background()

	// global variables here like firebase client etc...
	conf := &firebase.Config{
		ProjectID:   "peezme",
		DatabaseURL: "https://peezme-default-rtdb.firebaseio.com/",
	}

	// new firebase app instance
	app, err := firebase.NewApp(context.Background(), conf)
	if err != nil {
		_ = fmt.Errorf("error initializing firebase database app: %v", err)
	}

	// connect to the database
	database, err := app.Database(context.Background())

	if err != nil {
		_ = fmt.Errorf("error connecting to the database app: %v", err)
	}

	// Get a database reference to our game details.
	lobbyRef = database.NewRef("lobby/")
}

// joinHandler responds to requests by rendering an HTML page.
func defHandler(w http.ResponseWriter, r *http.Request) {
	if err := tmpl.Execute(w, data); err != nil {
		msg := http.StatusText(http.StatusInternalServerError)
		log.Printf("template.Execute: %v", err)
		http.Error(w, msg, http.StatusInternalServerError)
	}
}

func joinHandler(w http.ResponseWriter, r *http.Request) {
	// join room logic

	req := models.LobbyRoomRequest{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return
	}

	rUser := req.User
	player := lobbyRef.Child("private/").Child(req.RoomId).Child("players")
	player.Set(context.Background(), &rUser)
}

func leaveHandler(w http.ResponseWriter, r *http.Request) {
	//todo leave room logic
	req := models.LobbyRoomRequest{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return
	}

	player := lobbyRef.Child("private/").Child(req.RoomId).Child("players")
	player.Child(req.User.ID).Delete(context.Background())
}
