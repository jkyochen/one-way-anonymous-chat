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

	name := faker.Name()
	defer cleanCache(name)
	defer conn.Close()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			logrus.Error("read:", err)
			break
		}
		bc.sendMsg(message, name, conn)
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
	router.Handle("/", http.HandlerFunc(homeRouter))
	router.Handle("/chat", http.HandlerFunc(chatRouter))
	return router
}
