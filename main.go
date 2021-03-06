package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

// struct of the model of the db
type Msg struct {
	gorm.Model
	Channel string `json:"channel"`
	User    string `json:"user"`
	Content string `json:"content"`
}

// This is a struct for us to store the message that is posted into
type Message struct {
	Channel string `json:"channel"`
	User    string `json:"user"`
	Content string `json:"content"`
}

type Response struct {
	Name string
	Body []string
}

func initDb() *gorm.DB {
	db, err := gorm.Open("sqlite3", "msg.db")
	if err != nil {
		panic("failed to connect database")
	}
	return db
}

func migrateDB(db *gorm.DB) {
	db.AutoMigrate(&Msg{})
}

func postMessage(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	// get the form data
	r.ParseForm()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println("Ooops! somethings gone wrong")
		os.Exit(1)
	}

	// change the json data in the body into a message struct
	var message Message
	err = json.Unmarshal(body, &message)

	fmt.Println("Params: ")
	// this deals with post params if we send any through - we're just printing them here
	for k, v := range r.Form {
		fmt.Println("key:", k)
		fmt.Println("val:", strings.Join(v, ""))
	}

	// send a response echoing the values sent in - just for demonstration
	response := Response{"Message", []string{message.Channel, message.User, message.Content}}

	msg := Msg{Channel: message.Channel, User: message.User, Content: message.Content}
	db.NewRecord(msg)
	db.Create(&msg)

	js, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func welcome(w http.ResponseWriter, r *http.Request) {
	profile := Response{"OK", []string{"Ok", "OK"}}

	js, err := json.Marshal(profile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func getMessages(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	r.ParseForm()
	// set up an array od Msg
	var msgs []Msg
	// get the result
	db.Where("channel = ?", r.Form["channel"]).Find(&msgs)
	//cycle over the results
	for _, msg := range msgs {
		fmt.Println(msg)
		//fmt.Println("val:", strings.Join(v, ""))
	}
	js, err := json.Marshal(msgs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// spit out the values in the channel
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func main() {

	db := initDb()
	migrateDB(db)
	//get the local IP
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println("Ooops! somethings gone wrong")
		os.Exit(1)
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				fmt.Println("Send Requests to: http://" + ipnet.IP.String() + ":8666")
				//os.Stdout.WriteString(ipnet.IP.String() + "\n")
			}
		}
	}
	fmt.Println("Ctrl-C exit!")

	http.HandleFunc("/", welcome)
	http.HandleFunc("/message", func(w http.ResponseWriter, r *http.Request) {
		postMessage(w, r, db)
	})
	http.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		getMessages(w, r, db)
	})
	err = http.ListenAndServe(":8666", nil)
	if err != nil {
		fmt.Println("ListenAndServe: ", err)
	}
}
