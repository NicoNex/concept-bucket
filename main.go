package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/NicoNex/echotron"
	"github.com/teris-io/shortid"
)

type Bucket struct {
	Name     string
	Concepts map[string]Concept
}

type Concept struct {
	Title string
	Body  string
	Date  int64
}

// Struct representing the data cached on the disk for each bot.
type Data struct {
	Buckets []string
	Curid   string
}

type bot struct {
	chatId int64
	data   Data
	state  stateFn
	bucket *Bucket
	tmp    string
	echotron.Api
}

// Recursive definition of the state-function type.
type stateFn func(*echotron.Update) stateFn

var cc Cache
var ar Archive
var sid *shortid.Shortid

func newBot(api echotron.Api, chatId int64) echotron.Bot {
	data, err := cc.Get(chatId)
	if err != nil {
		log.Println("newBot", err)
	}

	b := &bot{
		chatId: chatId,
		data:   data,
		Api:    api,
	}
	b.state = b.handleMessage
	b.loadBucket()
	return b
}

func (b *bot) loadBucket() {
	// holds the id of the bucket currently used
	var id = b.data.Curid

	if id != "" {
		bk, err := ar.Get(id)
		if err != nil {
			log.Println("loadBucket", err)
			return
		}
		b.bucket = &bk
	}
}

// Returns the message from the given update.
func extractMessage(update *echotron.Update) string {
	// Some defensive programming here...
	if update.Message != nil {
		return update.Message.Text
	} else if update.EditedMessage != nil {
		return update.EditedMessage.Text
	} else {
		return ""
	}
}

// Sends a confirmation message for the newly created bucket.
func (b bot) confirmBucket(id, name string) {
	b.SendMessageOptions(
		fmt.Sprintf("New bucket created!\n\nName: *%s*\nID: *%s*", name, id),
		b.chatId,
		echotron.PARSE_MARKDOWN,
	)
}

// Sends an overview of the bucket associated with the provided id.
func (b bot) sendBucketOverview(id string) {
	bk, err := ar.Get(id)
	if err != nil {
		log.Println("sendBucketOverview", err)
		b.SendMessage("Something went wrong...", b.chatId)
		return
	}

	b.SendMessageOptions(
		fmt.Sprintf("Name: *%s*\nID: `%s`", bk.Name, id),
		b.chatId,
		echotron.PARSE_MARKDOWN,
	)
}

func (b bot) sendConcept(c Concept) {
	b.SendMessageOptions(
		fmt.Sprintf("*%s*\n\n%s", c.Title, c.Body),
		b.chatId,
		echotron.PARSE_MARKDOWN,
	)
}

// Creates a new bucket and saves the relative data onto the disk.
func (b *bot) newBucket(update *echotron.Update) stateFn {
	var name = extractMessage(update)

	// Generate the id of the bucket.
	id, err := sid.Generate()
	if err != nil {
		b.SendMessage("Something went wrong...", b.chatId)
		log.Println("newBucket", err)
		return b.handleMessage
	}
	b.data.Buckets = append(b.data.Buckets, id)
	b.data.Curid = id
	go b.updateData()
	b.bucket = &Bucket{Name: name}

	// Save the bucket into the db.
	if err := b.updateBucket(); err == nil {
		b.confirmBucket(id, name)
	} else {
		b.SendMessage("Something went wrong...", b.chatId)
	}
	return b.handleMessage
}

// Adds a bucket from its ID.
func (b *bot) addBucket(update *echotron.Update) stateFn {
	var id = extractMessage(update)

	if b.isExistingId(id) {
		b.data.Buckets = append(b.data.Buckets, id)
		go b.updateData()
		b.SendMessage("Bucket added successfully", b.chatId)
	}
	return b.handleMessage
}

// Sets the currently-in-use bucket.
func (b *bot) setBucket(update *echotron.Update) stateFn {
	var id = extractMessage(update)

	if b.isValidId(id) {
		bk, err := ar.Get(id)
		if err != nil {
			log.Println("setBucket", err)
			b.SendMessage("Something went wrong...", b.chatId)
			return b.handleMessage
		}
		b.bucket = &bk
		b.data.Curid = id
		go b.updateData()
		b.SendMessage("Bucket set successfully", b.chatId)
	} else {
		b.SendMessage("Invalid ID", b.chatId)
	}

	return b.handleMessage
}

func (b *bot) newConceptTitle(update *echotron.Update) stateFn {
	b.tmp = extractMessage(update)
	b.SendMessage("What's the new concept?", b.chatId)
	return b.newConceptBody
}

func (b *bot) newConceptBody(update *echotron.Update) stateFn {
	var body = extractMessage(update)

	if b.bucket.Concepts == nil {
		b.bucket.Concepts = make(map[string]Concept)
	}

	b.bucket.Concepts[b.tmp] = Concept{
		Title: b.tmp,
		Body:  body,
		Date:  time.Now().Unix(),
	}
	go b.updateBucket()
	b.SendMessage("Concept added successfully", b.chatId)
	return b.handleMessage
}

// Handles the messages when the bot is in its normal state.
func (b *bot) handleMessage(update *echotron.Update) stateFn {
	switch extractMessage(update) {

	case "/start":
		b.welcomeMessage()

	case "🆕 New bucket":
		b.SendMessage("What's the name of the bucket?", b.chatId)
		return b.newBucket

	case "🗑 My buckets":
		if b.data.Buckets != nil && len(b.data.Buckets) > 0 {
			for _, bk := range b.data.Buckets {
				b.sendBucketOverview(bk)
			}
		} else {
			b.SendMessage("You have no bucket", b.chatId)
		}

	case "➕ Add bucket":
		b.SendMessage("What's the ID of the bucket you want to add?", b.chatId)
		return b.addBucket

	case "☑️ Set bucket":
		b.SendMessage("What's the ID of the bucket you want to use?", b.chatId)
		return b.setBucket

	case "💡 New concept":
		if b.bucket != nil {
			b.SendMessage("What's the title of the new concept?", b.chatId)
			return b.newConceptTitle
		}
		b.SendMessage("No bucket selected, please select or create one first", b.chatId)

	case "📝 My concepts":
		b.loadBucket()
		if b.bucket != nil {
			if len(b.bucket.Concepts) > 0 {
				for _, c := range b.bucket.Concepts {
					b.sendConcept(c)
				}
			} else {
				b.SendMessage("The bucket is empty", b.chatId)
			}
		} else {
			b.SendMessage("No bucket selected, please select or create one first", b.chatId)
		}

	case "❓ Which bucket":
		if id := b.data.Curid; id != "" {
			b.sendBucketOverview(id)
		} else {
			b.SendMessage("No bucket selected, please select or create one first", b.chatId)
		}
	}

	return b.handleMessage
}

// Sends the welcome message and the bot keyboard.
func (b bot) welcomeMessage() {
	// Generate the keyboard
	kbd := b.KeyboardMarkup(false, true, false,
		b.KeyboardRow(
			b.KeyboardButton("🆕 New bucket", false, false),
			b.KeyboardButton("🗑 My buckets", false, false),
		),
		b.KeyboardRow(
			b.KeyboardButton("➕ Add bucket", false, false),
			b.KeyboardButton("☑️ Set bucket", false, false),
		),
		b.KeyboardRow(
			b.KeyboardButton("💡 New concept", false, false),
			b.KeyboardButton("📝 My concepts", false, false),
		),
		b.KeyboardRow(
			b.KeyboardButton("❓ Which bucket", false, false),
			b.KeyboardButton("❌ Cancel", false, false),
		),
	)
	b.SendMessageWithKeyboard("Welcome to Concept Bucket!", b.chatId, kbd)
}

// Updates the cache db on the disk.
func (b bot) updateData() {
	if err := cc.Put(b.chatId, b.data); err != nil {
		b.SendMessage("Something went wrong...", b.chatId)
		log.Println("updateData", err)
	}
}

// Syncs the current bucket in use with the database.
func (b bot) updateBucket() error {
	return ar.Put(b.data.Curid, *b.bucket)
}

// Returns true if the given id exists in the buckets db.
func (b bot) isExistingId(id string) bool {
	kch, err := ar.Keys()
	if err != nil {
		log.Println("isExistingId", err)
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
	for _, i := range b.data.Buckets {
		if id == i {
			return true
		}
	}
	return false
}

// I think we all know what this function does.
func (b *bot) Update(update *echotron.Update) {
	// The command '❌ Cancel' needs to take precedence over everything.
	if msg := extractMessage(update); msg == "❌ Cancel" {
		b.SendMessage("Action cancelled", b.chatId)
		b.state = b.handleMessage
		return
	}
	b.state = b.state(update)
}

func main() {
	var home = os.Getenv("HOME")

	cc = Cache(fmt.Sprintf("%s/.cache/concept-bucket/cache", home))
	ar = Archive(fmt.Sprintf("%s/.cache/concept-bucket/buckets", home))
	sid = shortid.MustNew(0, shortid.DefaultABC, uint64(time.Now().Unix()))
	dsp := echotron.NewDispatcher(
		readToken(fmt.Sprintf("%s/.config/concept-bucket", home)),
		newBot,
	)
	dsp.Run()
}
