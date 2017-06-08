package main

import (
	//"github.com/davecgh/go-spew/spew"
	"bytes"
	"fmt"
	"github.com/sethdmoore/execpy/config"
	"github.com/sethdmoore/execpy/globals"
	"github.com/tucnak/telebot"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"time"
)

func execScript(path string, timeout int, bin string) (result string) {
	cmd := exec.Command(bin, path)
	var stdout, stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Start()
	timer := time.AfterFunc(time.Duration(timeout)*time.Second, func() {
		err := cmd.Process.Kill()
		if err != nil {
			panic(err)
		}
	})

	err := cmd.Wait()
	if err != nil {
		log.Println(err)
		result = fmt.Sprintf("Command timed out after maximum %d seconds", timeout)
	} else {
		result = stdout.String() + stderr.String()
	}
	timer.Stop()

	return
}

func evalScript(c *config.Config, content *string) string {
	f, err := ioutil.TempFile("", globals.AppPrefix+"_")
	if err != nil {
		log.Fatalf("Could create temporary file%s\n", f.Name())
	}
	bytecontent := []byte(*content)
	defer os.Remove(f.Name())

	if _, err := f.Write(bytecontent); err != nil {
		log.Fatal(err)
	}

	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
	result := execScript(f.Name(), c.Timeout, c.Binary)
	return result

}

func evalPreformattedScript(c *config.Config, content *string, message telebot.Message, bot *telebot.Bot) {
	result := evalScript(c, content)
	bot.SendMessage(message.Chat, result, nil)

}

func messages(c *config.Config, bot *telebot.Bot) {
	for message := range bot.Messages {
		var script string
		if message.Entities != nil {
			for _, e := range message.Entities {
				if e.Type == "pre" {
					script += message.Text[e.Offset:e.Length+e.Offset] + "\n"
				}
			}
		}

		if len(script) > 0 {
			go evalPreformattedScript(c, &script, message, bot)
		}

		if message.Text == "/start" {
			bot.SendMessage(message.Chat,
				"Hello, "+message.Sender.FirstName+"!", nil)
		} else {

		}
	}

	for message := range bot.Messages {
		log.Printf("Received a message from %s with text %s\n",
			message.Sender.Username, message.Text)
	}
}

func queries(c *config.Config, bot *telebot.Bot) {
	for query := range bot.Queries {
		log.Println("-- New Query --")
		log.Println("from: ", query.From.Username)
		log.Println("text: ", query.Text)

		//go evalScript(c, &script, message, bot)

		result := "Lorem ipsum"

		article := &telebot.InlineQueryResultArticle{
			Title: "Telebot",
			//URL:   "http://google.com",
			InputMessageContent: &telebot.InputTextMessageContent{
				Text:           result,
				DisablePreview: false,
			},
		}

		results := []telebot.InlineQueryResult{article}
		response := telebot.QueryResponse{
			Results:    results,
			IsPersonal: true,
		}

		if err := bot.AnswerInlineQuery(&query, &response); err != nil {
			log.Println("Failed to respond to query", err)
		}

	}
}

func main() {
	c := config.Get()

	bot, err := telebot.NewBot(c.Token)
	if err != nil {
		log.Fatalln(err)
	}
	bot.Messages = make(chan telebot.Message, 100)
	bot.Queries = make(chan telebot.Query, 1000)

	go messages(c, bot)
	go queries(c, bot)

	bot.Start(1 * time.Second)

}
