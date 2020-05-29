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
    ADD_BUCKET
    SET_BUCKET
)

type bot struct {
	chatId int64
	buckets []string
    state State
    bucket *Bucket
    curid string
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

    case "/add_bucket":
        b.SendMessage("What's the ID of the bucket you want to add?", b.chatId)
        b.state = ADD_BUCKET

    case "/set_bucket":
        b.SendMessage("What's the ID of the bucket you want to use?", b.chatId)
        b.state = SET_BUCKET
    }
}

// Updates the cache db on the disk.
func (b bot) updateCache() {
    if err := cc.Put(b.chatId, b.buckets); err != nil {
        b.SendMessage("Something went wrong...", b.chatId)
        log.Println(err)
    }
}

func (b bot) updateBucket() error {
    return ar.Put(b.curid, *b.bucket)
}

// Returns true if the given id exists in the buckets db.
func (b bot) isExistingId(id string) bool {
    kch, err := ar.Keys()
    if err != nil {
        log.Println(err)
        b.SendMessage("Something went wrong...", b.chatId)
        return false
    }

    for k := range kch {
        if id == string(k) {
            return true
        }
    }
    return false
}

// Returns true if the bucket id is associated with the current bot.
func (b bot) isValidId(id string) bool {
    for _, i := range b.buckets {
        if id == i {
            return true
        }
    }
    return false
}

// TODO: convert this switch case into a functional approach.
func (b *bot) Update(update *echotron.Update) {
    var msg = b.extractMessage(update)

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
        go b.updateCache()
        b.bucket = &Bucket{Name: msg}
        b.curid = id

        // Save the bucket into the db.
        if err := b.updateBucket(); err == nil {
            b.confirmBucket(id, msg)
        } else {
            b.SendMessage("Something went wrong...", b.chatId)
        }
        b.state = NO_OP

    case ADD_BUCKET:
        if b.isExistingId(msg) {
            b.buckets = append(b.buckets, msg)
            go b.updateCache()
            b.SendMessage("Bucket added successfully", b.chatId)
        }
        b.state = NO_OP

    case SET_BUCKET:
        if b.isValidId(msg) {
            bk, err := ar.Get(msg)
            if err != nil {
                log.Println(err)
                b.SendMessage("Something went wrong...", b.chatId)
                b.state = NO_OP
                return
            }
            b.bucket = &bk
            b.SendMessage("Bucket set successfully", b.chatId)
        } else {
            b.SendMessage("Invalid ID", b.chatId)
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
