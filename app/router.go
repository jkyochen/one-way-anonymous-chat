package app

import (
	"net/http"
	"os"
	"strconv"
	"text/template"

	faker "github.com/bxcodec/faker/v3"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

var homeTemplate = template.Must(template.ParseFiles("assets/index.html"))

var upgrader = websocket.Upgrader{}

var bc *botClient

func homeRouter(w http.ResponseWriter, r *http.Request) {
	err := homeTemplate.Execute(w, "ws://"+r.Host+"/chat")
	if err != nil {
		logrus.Error("homeRouter:", err)
	}
}

func chatRouter(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logrus.Error("chatRouter:", err)
		return
	}
	defer conn.Close()

	name := faker.Name()
	defer bc.cleanCache(name)

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			logrus.Error("read:", err)
			return
		}
		if messageType == websocket.TextMessage {
			err := bc.sendMsg(message, name, conn)
			if err != nil {
				return
			}
		}
	}
}

func load() http.Handler {

	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	chatID, err := strconv.Atoi(os.Getenv("TELEGRAM_CHAT_ID"))
	if err != nil {
		logrus.Fatal("input chatID is error")
	}
	bc = newBotClient(botToken, int64(chatID))

	router := mux.NewRouter()
	router.HandleFunc("/", homeRouter)
	router.HandleFunc("/chat", chatRouter)
	return router
}
