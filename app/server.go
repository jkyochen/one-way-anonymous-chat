package app

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"

	faker "github.com/bxcodec/faker/v3"
	"github.com/gorilla/websocket"

	// dotenv load .env config to current environment variables
	_ "github.com/joho/godotenv/autoload"
)

// Run server main run
func Run() {

	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	chatID, err := strconv.Atoi(os.Getenv("TELEGRAM_CHAT_ID"))
	if err != nil {
		log.Fatal("input chatID is error")
	}

	homeTemplate := template.Must(template.ParseFiles("assets/index.html"))
	upgrader := websocket.Upgrader{}
	botClient := newBotClient(botToken, int64(chatID))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		err := homeTemplate.Execute(w, "ws://"+r.Host+"/chat")
		if err != nil {
			log.Panic(err)
		}
	})

	http.HandleFunc("/chat", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("upgrade:", err)
			return
		}

		name := faker.Name()
		defer cleanCache(name)
		defer conn.Close()

		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				break
			}
			botClient.sendMsg(message, name, conn)
		}
	})

	addr := os.Getenv("SERVER_ADDR")
	log.Fatal(http.ListenAndServe(addr, nil))
}
