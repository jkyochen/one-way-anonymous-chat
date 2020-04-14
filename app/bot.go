package app

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/gorilla/websocket"
)

type botClient struct {
	api    *tgbotapi.BotAPI
	chatID int64
}

func newBotClient(token string, chatID int64) *botClient {

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = false

	log.Printf("Authorized on account %s", bot.Self.UserName)

	return &botClient{bot, chatID}
}

type clientMessageRelation struct {
	msgIds []int
	client *websocket.Conn
}

var cache = make(map[string]clientMessageRelation)

func (b *botClient) recvMsg() {

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := b.api.GetUpdatesChan(u)
	if err != nil {
		log.Println("GetUpdatesChan:", err)
		return
	}

	contains := func(msgIds []int, mId int) bool {
		for _, id := range msgIds {
			if id == mId {
				return true
			}
		}
		return false
	}

	for update := range updates {
		if update.Message == nil {
			continue
		}
		if update.Message.Chat.ID != b.chatID {
			continue
		}

		replyToMessage := update.Message.ReplyToMessage
		if replyToMessage == nil {
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		for _, v := range cache {
			if contains(v.msgIds, replyToMessage.MessageID) {
				err := v.client.WriteMessage(websocket.TextMessage, []byte(update.Message.Text))
				if err != nil {
					log.Println("write:", err)
					break
				}
			}
		}
	}
}

func (b *botClient) sendMsg(msg []byte, name string, conn *websocket.Conn) {

	nm := tgbotapi.NewMessage(b.chatID, name+"\n\n"+string(msg))
	m, err := b.api.Send(nm)
	if err != nil {
		log.Println("sendMsg:", err)
		return
	}

	if v, ok := cache[name]; ok {
		v.msgIds = append(v.msgIds, m.MessageID)
		cache[name] = clientMessageRelation{
			msgIds: v.msgIds,
			client: v.client,
		}
		return
	}

	cache[name] = clientMessageRelation{
		msgIds: []int{m.MessageID},
		client: conn,
	}
}

func cleanCache(name string) {
	delete(cache, name)
}
