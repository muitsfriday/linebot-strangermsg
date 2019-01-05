package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/line/line-bot-sdk-go/linebot"
	mgo "gopkg.in/mgo.v2"
)

// MsgRecord struct represent one record
type MsgRecord struct {
	message linebot.Message        `bson:"message"`
	users   []*linebot.EventSource `bson:"users"`
}

func main() {

	fmt.Println("enter main")
	bot, err := linebot.New(os.Getenv("SECRET"), os.Getenv("TOKEN"))
	if err != nil {
		fmt.Println(err)
	}

	ss := strings.Split(os.Getenv("MONGODB_URI"), "/")
	dbname := ss[len(ss)-1]

	session, errMgo := mgo.Dial(os.Getenv("MONGODB_URI"))
	if errMgo != nil {
		fmt.Print(os.Getenv("MONGODB_URI"))
		panic("end")
	}
	defer session.Close()

	msgColl := session.DB(dbname).C("messages")

	// Setup HTTP Server for receiving requests from LINE platform
	http.HandleFunc("/callback", func(w http.ResponseWriter, req *http.Request) {
		events, err := bot.ParseRequest(req)
		if err != nil {
			if err == linebot.ErrInvalidSignature {
				w.WriteHeader(400)
			} else {
				w.WriteHeader(500)
			}
			return
		}

		for _, event := range events {
			fmt.Println("events: ", events)

			if event.Type == linebot.EventTypeMessage {
				switch message := event.Message.(type) {
				case *linebot.TextMessage:
					fmt.Println("text received: ", message.Text)
										
					msg := MsgRecord{
						event.Message,
						event.Members,
					}
					errInsert := msgColl.Insert(event.Message)

					fmt.Println("msg", msg)

					if errInsert != nil {
						fmt.Println(errInsert.Error())
					}
					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(message.Text)).Do(); err != nil {
						fmt.Println(err)
					}
				}
			}
		}
	})

	fmt.Println("server start")

	// This is just sample code.
	// For actual use, you must support HTTPS by using `ListenAndServeTLS`, a reverse proxy or something else.
	if err := http.ListenAndServe(":"+os.Getenv("PORT"), nil); err != nil {
		fmt.Println(err.Error())
	}

}
