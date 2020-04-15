package app

import (
	"os"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

type clientMessageRelation struct {
	msgIds []int
	client *websocket.Conn
}

type botClient struct {
	api    *tgbotapi.BotAPI
	chatID int64
	cache  map[string]clientMessageRelation
}

func newBotClient(token string, chatID int64) *botClient {

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		logrus.Fatal("input chatID is error: ", err)
	}

	isDebug, err := strconv.ParseBool(os.Getenv("SERVER_DEBUG"))
	if err != nil {
		logrus.Fatal("input chatID is error: ", err)
	}

	if isDebug {
		bot.Debug = true
	}

	logrus.Infof("Authorized on account: %s", bot.Self.UserName)

	bc := &botClient{
		api:    bot,
		chatID: chatID,
		cache:  make(map[string]clientMessageRelation, 5),
	}
	go bc.recvMsg()

	return bc
}

func (b *botClient) recvMsg() {

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := b.api.GetUpdatesChan(u)
	if err != nil {
		logrus.Fatal("GetUpdatesChan:", err)
		return
	}

	isContainMsgId := func(msgIds []int, mId int) bool {
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

		logrus.Infof("[%s] %s", update.Message.From.UserName, update.Message.Text)

		for _, v := range b.cache {
			if isContainMsgId(v.msgIds, replyToMessage.MessageID) {
				err := v.client.WriteMessage(websocket.TextMessage, []byte(update.Message.Text))
				if err != nil {
					logrus.Error("recvMsg:", err)
					break
				}
			}
		}
	}
}

func (b *botClient) sendMsg(msg []byte, name string, conn *websocket.Conn) error {

	nm := tgbotapi.NewMessage(b.chatID, name+"\n\n"+string(msg))
	m, err := b.api.Send(nm)
	if err != nil {
		logrus.Error("sendMsg:", err)
		return err
	}

	if v, ok := b.cache[name]; ok {
		v.msgIds = append(v.msgIds, m.MessageID)
		b.cache[name] = clientMessageRelation{
			msgIds: v.msgIds,
			client: v.client,
		}
	} else {
		b.cache[name] = clientMessageRelation{
			msgIds: []int{m.MessageID},
			client: conn,
		}
	}
	return nil
}

func (b *botClient) cleanCache(name string) {
	delete(b.cache, name)
}
