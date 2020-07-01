package main

import (
	//"fmt"
	"time"
	"log"
	"net/http"
	"context"
	"os"
	"cloud.google.com/go/firestore"
	"github.com/gorilla/mux"
	"html/template"
	"google.golang.org/api/iterator"
)

var tmplStr = `
<html><head>
<style>
*{
  margin:0;
  padding:0;
}
body{
  font-family:arial,sans-serif;
  font-size:100%;
  margin:3em;
  background:#666;
  color:#fff;
}
h2,p{
  font-size:100%;
  font-weight:normal;
}
ul,li{
  list-style:none;
}
ul{
  overflow:hidden;
  padding:3em;
}
ul li a{
  text-decoration:none;
  color:#000;
  background:#ffc;
  display:block;
  height:10em;
  width:10em;
  padding:1em;
  -moz-box-shadow:5px 5px 7px rgba(33,33,33,1);
  -webkit-box-shadow: 5px 5px 7px rgba(33,33,33,.7);
  box-shadow: 5px 5px 7px rgba(33,33,33,.7);
  -moz-transition:-moz-transform .15s linear;
  -o-transition:-o-transform .15s linear;
  -webkit-transition:-webkit-transform .15s linear;
}
ul li{
  margin:1em;
  float:left;
}
ul li h2{
  font-size:100%;
  font-weight:bold;
  padding-bottom:10px;
}
ul li p{
  font-family:"Reenie Beanie",arial,sans-serif;
  font-size:100%;
}
ul li a{
  -webkit-transform: rotate(-6deg);
  -o-transform: rotate(-6deg);
  -moz-transform:rotate(-6deg);
}
ul li:nth-child(even) a{
  -o-transform:rotate(4deg);
  -webkit-transform:rotate(4deg);
  -moz-transform:rotate(4deg);
  position:relative;
  top:5px;
  background:#cfc;
}
ul li:nth-child(3n) a{
  -o-transform:rotate(-3deg);
  -webkit-transform:rotate(-3deg);
  -moz-transform:rotate(-3deg);
  position:relative;
  top:-5px;
  background:#ccf;
}
ul li:nth-child(5n) a{
  -o-transform:rotate(5deg);
  -webkit-transform:rotate(5deg);
  -moz-transform:rotate(5deg);
  position:relative;
  top:-10px;
}
ul li a:hover,ul li a:focus{
  box-shadow:10px 10px 7px rgba(0,0,0,.7);
  -moz-box-shadow:10px 10px 7px rgba(0,0,0,.7);
  -webkit-box-shadow: 10px 10px 7px rgba(0,0,0,.7);
  -webkit-transform: scale(1.25);
  -moz-transform: scale(1.25);
  -o-transform: scale(1.25);
  position:relative;
  z-index:5;
}

.user {
  text-decoration:none;
  text-align: center;
  color:#000;
  background:#ffc;
  display:block;
  height:2em;
  width:10em;
  padding:0.5em;
  -moz-box-shadow:5px 5px 7px rgba(33,33,33,1);
  -webkit-box-shadow: 5px 5px 7px rgba(33,33,33,.7);
  box-shadow: 5px 5px 7px rgba(33,33,33,.7);
  -moz-transition:-moz-transform .15s linear;
  -o-transition:-o-transform .15s linear;
  -webkit-transition:-webkit-transform .15s linear;
}

.note {
  padding:0.5em;
  text-align: center;
  text-decoration:none;
  color:#000;
  background:#ffc;
  display:block;
  height:10em;
  width:10em;
  padding:1em;
  -moz-box-shadow:5px 5px 7px rgba(33,33,33,1);
  -webkit-box-shadow: 5px 5px 7px rgba(33,33,33,.7);
  box-shadow: 5px 5px 7px rgba(33,33,33,.7);
  -moz-transition:-moz-transform .15s linear;
  -o-transition:-o-transform .15s linear;
  -webkit-transition:-webkit-transform .15s linear;
}

ul li{
  margin:1em;
  float:left;
}
time {
 font-size: 60%;
}

input[type=submit] {
  background-color: #4CAF50;
  border: none;
  color: white;
  text-decoration: none;
  margin: 4px 2px;
  cursor: pointer;
}
h1 {
  text-align: center;
  color: white;
  font-size: 30px;
}
</style>
</head>

<body>
<h1>MMEC Notice Board</h1>
<ul>
{{range .}}
    <li>
    <a href="#">
        <time>{{.Timestamp.Format "Jan 02, 2006 15:04 IST"}}</time>
        <h2>{{.User}}</h2>
        <p>{{.Note}}</p>
    </a>
    </li>
{{end}}
</ul>
  <form action="/" method="POST" novalidate>
    <textarea placeholder='User' name='user' class="user"></textarea>
    <textarea placeholder='Note' name='note' class="note"></textarea>
    <input type="submit" value="Submit new note">
 </form>
</body>
</html>
`

var tmpl = template.Must(template.New("t").Parse(tmplStr))

type Document struct {
	User string
	Note string
	Timestamp time.Time
}

type getHandler struct {
	client *firestore.Client
}

func getDocuments(client *firestore.Client) []*Document {
	ctx := context.Background()
	iter := client.Collection("Wall").Query.OrderBy("Timestamp", firestore.Desc).Documents(ctx)
	defer iter.Stop()
	var docs []*Document
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		d := &Document{}
		doc.DataTo(d)
		docs = append(docs, d)
	}
	return docs
}

func (h *getHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	docs := getDocuments(h.client)
	if err := tmpl.Execute(w, docs); err != nil {
		log.Printf("got error with template: %v", err)
	}
}

type putHandler struct {
	client *firestore.Client
}

func (h *putHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	note := r.PostFormValue("note")
	user := r.PostFormValue("user")
	log.Printf("Got form: %v, %v", note, user)
	ref := h.client.Collection("Wall").NewDoc()
	d := &Document{}
	d.User = user
	d.Note = note
	d.Timestamp = time.Now()
	ctx := context.Background()
	if _, err := ref.Create(ctx, d); err != nil {
		log.Printf("ref.Create: %v", err)
	}
	docs := getDocuments(h.client)
	if err := tmpl.Execute(w, docs); err != nil {
		log.Printf("put got error with template: %v", err)
	}
}

func registerHandlers(h *getHandler, p *putHandler) {
	r := mux.NewRouter()
	r.Methods("GET").Path("/").Handler(h)
	r.Methods("POST").Path("/").Handler(p)
	http.Handle("/", r)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		log.Fatal("GOOGLE_CLOUD_PROJECT must be set")
	}

	ctx := context.Background()
	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("firestore.NewClient: %v", err)
	}
	h := &getHandler{
		client,
	}
	p := &putHandler {
		client,
	}
	registerHandlers(h, p)
	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
