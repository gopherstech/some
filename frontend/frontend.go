package main

import (
	"fmt"
	"log"
	"net/http"
	"net/rpc"
	"runtime"
	"time"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

var (
	DEFAULT_PORT = 3410
	DEFAULT_HOST = "localhost"
)
var send bool
var msg string = "welcome"

type Nothing bool

type Message struct {
	User   string
	Target string
	Msg    string
}

type ChatClient struct {
	Username string
	Address  string
	Client   *rpc.Client
}

var name = "Nicko"
var messages = make([]string, 0, 10)

func (c *ChatClient) CheckMessages() {
	var reply []string
	c.Client = c.getClientConnection()

	for {
		err := c.Client.Call("ChatServer.CheckMessages", c.Username, &reply)
		if err != nil {
			log.Fatalln("Chat has been shutdown. Goodbye.")
		}

		messages = append(messages, reply...)

		time.Sleep(10 * time.Second)
	}
}

func (c *ChatClient) Say(msg string) {
	var reply Nothing
	c.Client = c.getClientConnection()

	message := Message{
		User:   c.Username,
		Target: c.Address,
		Msg:    msg,
	}

	err := c.Client.Call("ChatServer.Say", message, &reply)
	if err != nil {
		log.Printf("Error saying something: %q", err)
	}

}

func (c *ChatClient) getClientConnection() *rpc.Client {
	var err error

	if c.Client == nil {
		c.Client, err = rpc.DialHTTP("tcp", c.Address)
		if err != nil {
			log.Panicf("Error establishing connection with host: %q", err)
		}
	}

	return c.Client
}

func createClient() (*ChatClient, error) {
	var c *ChatClient = &ChatClient{}

	c.Username = name
	c.Address = "localhost:3410"

	return c, nil
}

// Register takes a username and registers it with the server
func (c *ChatClient) Register() {
	var reply string
	c.Client = c.getClientConnection()

	err := c.Client.Call("ChatServer.Register", c.Username, &reply)
	if err != nil {
		log.Printf("Error registering user: %q", err)
	} else {
		log.Printf("Reply: %s", reply)
	}
}

var user string

type welcome struct {
	app.Compo
}

func (w *welcome) Render() app.UI {
	return app.Div().Class("grid").Class("grid-pad").Body(
		app.Div().Class("col-1-1").Body(
			app.Div().Class("content").Body(
				app.Input().Type("text").Value(user).Placeholder("Введите имя!").
					AutoFocus(true).OnChange(w.ValueTo(&user)),
				app.A().Href("/chat"),
			),
		),
	)
}

type index struct {
	app.Compo
	Messages []string
}

func (i *index) Render() app.UI {
	i.Messages = append(i.Messages, messages...)
	return app.Div().Class("grid").Class("grid-pad").Body(
		app.Div().Class("col-1-1").Body(
			app.Div().Class("content").Class("header").Body(
				app.H1().Text("Warg Messenger"),
			),
		),
		app.Range(i.Messages).Slice(func(c int) app.UI {
			return app.Li().Text(i.Messages[c])
		}),
		app.Div().Class("col-1-1").Body(
			app.Div().Class("content").Class("cmain").Body(),
		),
		app.Div().Class("col-1-1").Class("footer").Body(
			app.Div().Class("content").Body(
				app.Input().Type("text").Placeholder("Напишите сообщение!").OnChange(i.ValueTo(&msg)),
				app.Button().Text("Отправить").OnClick(
					func(ctx app.Context, e app.Event) {
						send = true
						ctx.Update()
					}),
			),
		),
	)
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	client, err := createClient()
	if err != nil {
		log.Panicf("Error creating client from flags: %q", err)
	}

	client.Register()

	go func() {
		for {
			go fmt.Println(messages)
			time.Sleep(5 * time.Second)
		}
	}()

	go client.CheckMessages()
	client.Say(msg)
	app.Route("/chat", func() app.Composer { return &index{} })
	app.Route("/", func() app.Composer { return &welcome{} })
	app.RunWhenOnBrowser()

	http.Handle("/", &app.Handler{
		Name:        "Hello",
		Description: "An Hello World! example",
		Styles: []string{
			"/web/styles.css",
		},
	})

	if err := http.ListenAndServe(":8000", nil); err != nil {
		log.Fatal(err)
	}
}
