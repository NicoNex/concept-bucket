package main

import (
	"log"
	"fmt"
	"time"
	"os"

	"github.com/teris-io/shortid"
	"github.com/NicoNex/echotron"
)

type Bucket struct {
	Name string
	Concepts map[string]Concept
}

type Concept struct {
	Title string
	Body  string
	Date  int64
}

type bot struct {
	chatId int64
	buckets []string
	echotron.Api
}

var cc Cache
var sid *shortid.Shortid

func die(a ...interface{}) {
	log.Println(a...)
	os.Exit(1)
}

func newBot(api echotron.Api, chatId int64) echotron.Bot {
	bs, err := cc.Get(chatId)
	if err != nil {
		log.Println(err)
	}

	return &bot{
		chatId: chatId,
		buckets: bs,
		Api: api,
	}
}

func (b bot) sendBucketId(id string) {
	b.SendMessageOptions(
		fmt.Sprintf("New bucket created with ID:\n*%s*", id),
		b.chatId,
		echotron.PARSE_MARKDOWN,
	)
}

func (b *bot) Update(update *echotron.Update) {
	var text string

    // Some defensive programming...
	if update.Message != nil {
		text = update.Message.Text
	} else if update.EditedMessage != nil {
		text = update.EditedMessage.Text
	} else {
		return
	}

	switch text {

	case "/start":
		b.SendMessage("I'm alive", b.chatId)
		fmt.Println(b.buckets)

	case "/new_bucket":
		// Generate the id of the bucket.
		id, err := sid.Generate()
		if err != nil {
			b.SendMessage("Something went wrong...", b.chatId)
			log.Println(err)
			return
		}
		b.buckets = append(b.buckets, id)

		// Add the id to the cache and associate it with the bot chatId.
		err = cc.Put(b.chatId, b.buckets)
		if err != nil {
			b.SendMessage("Something went wrong...", b.chatId)
			log.Println(err)
			return
		}
		// TODO: before sending the confirmation ask for the bucket name
		// then save the bucket in the archive.
		b.sendBucketId(id)
	}
}

func main() {
	cc = Cache("./cache")
	sid = shortid.MustNew(0, shortid.DefaultABC, uint64(time.Now().Unix()))
	dsp := echotron.NewDispatcher(
        "568059758:AAG32AudAzQyh_KEDqOsMSECbOgXY5fyu6U",
        newBot,
    )
	dsp.Run()
}
