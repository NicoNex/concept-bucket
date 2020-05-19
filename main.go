package main

import (
    "fmt"
	"time"
)

type Concept struct {
    Title string
    Body string
    Date int64
}

func (c Concept) String() string {
    return fmt.Sprintf(
        "Title: %s\nBody: %s\nDate: %d",
        c.Title, c.Body, c.Date,
    )
}

func main() {
    ar := Archive("./test")

    test := Concept{
        Title: "Sas Mike",
        Body: "Concept di prova...",
        Date: time.Now().Unix(),
    }

    err := ar.Put("fif", test)
    if err != nil {
        fmt.Println(err)
        return
    }

    c, err := ar.Get("fif")
    if err != nil {
        fmt.Println(err)
        return
    }

    fmt.Println(c)
}
