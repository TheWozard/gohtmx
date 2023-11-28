package main

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/TheWozard/gohtmx"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

const (
	CardEvent = "card"
)

type Room struct {
	sync.Mutex
	Users       []*User
	Multiplexer *gohtmx.Multiplexer
}

func (r *Room) Join() string {
	r.Lock()
	defer r.Unlock()
	user := uuid.New().String()
	r.Users = append(r.Users, &User{
		id: user, card: &Card{Text: user},
	})
	r.UpdateCards()
	return user
}

func (r *Room) Leave(user string) {
	r.Lock()
	defer r.Unlock()
	offset := 0
	for i, u := range r.Users {
		if u.id == user {
			offset += 1
		} else if offset > 0 {
			r.Users[i-offset] = u
		}
	}
	r.Users = r.Users[:len(r.Users)-offset]
	r.UpdateCards()
}

func (r *Room) SetUserCard(user string, card *Card) {
	r.Lock()
	defer r.Unlock()
	for _, u := range r.Users {
		if u.id == user {
			u.card = card
		}
	}
	r.UpdateCards()
}

func (r *Room) UpdateCards() {
	r.Multiplexer.Send(gohtmx.SSEEvent{
		Event: CardEvent,
		Data:  r.Slots(),
	})
}

func (r *Room) Slots() gohtmx.Component {
	result := gohtmx.Fragment{}
	for _, user := range r.Users {
		result = append(result, user.card.Card())
	}
	return result
}

type User struct {
	id   string
	card *Card
}

type Card struct {
	Text string
}

func (c *Card) Body() gohtmx.Component {
	if c == nil {
		return gohtmx.Raw("")
	}
	return gohtmx.Raw(c.Text)
}

func (c *Card) Card() gohtmx.Component {
	return gohtmx.Div{
		Classes: []string{"card"},
		Content: gohtmx.Fragment{
			gohtmx.Div{
				Classes: []string{"top-left"},
				Content: gohtmx.Raw("1"),
			},
			gohtmx.Div{
				Classes: []string{"bottom-right"},
				Content: gohtmx.Raw("1"),
			},
			c.Body(),
		},
	}
}

func main() {
	room := &Room{
		Users:       []*User{},
		Multiplexer: &gohtmx.Multiplexer{},
	}

	mux := mux.NewRouter()
	mux.PathPrefix("/assets/").Handler(http.FileServer(http.Dir("./example")))
	gohtmx.Document{
		Header: gohtmx.Fragment{
			gohtmx.Raw(`<meta charset="utf-8">`),
			gohtmx.Raw(`<meta name="viewport" content="width=device-width, initial-scale=1">`),
			gohtmx.Raw(`<title>Cards In a Room</title>`),
			gohtmx.Raw(`<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/bulma@0.9.4/css/bulma.min.css">`),
			gohtmx.Raw(`<link rel="stylesheet" href="/assets/style.css">`),
			gohtmx.Raw(`<script src="https://unpkg.com/htmx.org@1.9.6/dist/htmx.min.js"></script>`),
			gohtmx.Raw(`<script src="https://unpkg.com/htmx.org@1.9.6/dist/ext/sse.js"></script>`),
			gohtmx.Raw(`<meta name="htmx-config" content='{"requestClass":"is-loading"}'>`),
		},
		Body: RenderRoom(room),
	}.Mount("/", mux)

	log.Default().Println("staring server at http://localhost:8080")
	log.Fatal((&http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}).ListenAndServe())
}

func RenderRoom(room *Room) gohtmx.Component {
	return gohtmx.Stream{
		ID: "stream",
		SSEEventGenerator: func(ctx context.Context, c chan gohtmx.SSEEvent) {
			ctx = room.Multiplexer.Subscribe(ctx, c)
			c <- gohtmx.SSEEvent{
				Event: CardEvent,
				Data:  room.Slots(),
			}
			user := room.Join()
			<-ctx.Done()
			room.Leave(user)
		},
		Content: gohtmx.Fragment{
			gohtmx.StreamTarget{
				Events:  []string{CardEvent},
				Classes: []string{"columns", "is-centered", "is-multiline", "deck"},
			},
		},
	}
}
