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

type State int
const (
    NO_OP State = iota
    BUCKET_NAME
)

type bot struct {
	chatId int64
	buckets []string
    state State
    // tmpid string
    tmpname string
	echotron.Api
}

var cc Cache
var ar Archive
var sid *shortid.Shortid

func die(a ...interface{}) {
	log.Println(a...)
	os.Exit(1)
}

func (b bot) extractMessage(update *echotron.Update) string {
    if update.Message != nil {
        return update.Message.Text
    } else if update.EditedMessage != nil {
        return update.EditedMessage.Text
    } else {
        return ""
    }
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

func (b bot) confirmBucket(id, name string) {
	b.SendMessageOptions(
		fmt.Sprintf("New bucket created!\n\nName: *%s*\nID: *%s*", name, id),
		b.chatId,
		echotron.PARSE_MARKDOWN,
	)
}

func (b bot) sendBucketOverview(id string) {
    bk, err := ar.Get(id)
    if err != nil {
        log.Println(err)
        b.SendMessage("Something went wrong...", b.chatId)
        return
    }

    b.SendMessageOptions(
        fmt.Sprintf("Name: *%s*\nID: *%s*", bk.Name, id),
        b.chatId,
        echotron.PARSE_MARKDOWN,
    )
}

func (b *bot) handleMessage(msg string) {
    switch msg {

    case "/start":
        b.SendMessage("I'm alive", b.chatId)
        fmt.Println(b.buckets)

    case "/new_bucket":
        b.SendMessage("What's the name of the bucket?", b.chatId)
        b.state = BUCKET_NAME

    case "/my_buckets":
        for _, bk := range b.buckets {
            b.sendBucketOverview(bk)
        }
    }
}

func (b *bot) Update(update *echotron.Update) {
    msg := b.extractMessage(update)

	switch b.state {
    case NO_OP:
        b.handleMessage(msg)

    case BUCKET_NAME:
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

        // Save the bucket into the db.
        if err := ar.Put(id, Bucket{Name: msg}); err == nil {
            b.confirmBucket(id, msg)
        } else {
            b.SendMessage("Something went wrong...", b.chatId)
        }
        b.state = NO_OP
    }
}

func main() {
	cc = Cache("./cache")
    ar = Archive("./buckets")
	sid = shortid.MustNew(0, shortid.DefaultABC, uint64(time.Now().Unix()))
	dsp := echotron.NewDispatcher(
        "568059758:AAG32AudAzQyh_KEDqOsMSECbOgXY5fyu6U",
        newBot,
    )
	dsp.Run()
}
