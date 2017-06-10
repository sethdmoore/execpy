package main

import (
	//"github.com/davecgh/go-spew/spew"
	"bytes"
	"fmt"
	"errors"
	"github.com/sethdmoore/execpy/config"
	"github.com/sethdmoore/execpy/globals"
	"github.com/tucnak/telebot"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"time"
)

func trimString(s string, length int, maxlength int) string{
	if len(s) <= max {
		return s
	}

	return s[:
}


func isAuthorized(id int, list *[]int) bool {
	// no authorized users means all users are authorized
	if len(*list) == 0 {
		return true
	}

	for _, item := range *list {
		if id == item {
			return true
		}
	}
	return false
}

func execScript(path string, timeout int, bin string) (string, error) {
	var result string
	// need to outer-scope this var so that the err var can be set (and checked
	// inside of time.AfterFunc)
	var err error
	cmd := exec.Command(bin, path)

	var stdout, stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	cmd.Start()
	timer := time.AfterFunc(time.Duration(timeout)*time.Second, func() {
		err = cmd.Process.Kill()
	})

	if err != nil {
		// Can't imagine this happening unless there's a system-level issue
		e := errors.New(fmt.Sprintf("Problem killing proc: %s", err))
		log.Printf(fmt.Sprintf("%s", e))
		return fmt.Sprintf("Internal Error: %s", e), e
	}

	err = cmd.Wait()
	if stderr.String() != "" {
		log.Printf("There was stderr!")
		err = errors.New(stderr.String())
	}

	result = stdout.String() + stderr.String()
	if err != nil && result == "" {
		log.Println(err)
		result = fmt.Sprintf("Command timed out after maximum %d seconds", timeout)
	}
	timer.Stop()

	return result, err
}

func evalScript(c *config.Config, content *string) (string, error) {
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
	return execScript(f.Name(), c.Timeout, c.Binary)
}

func evalPreformattedScript(c *config.Config, content *string, message telebot.Message, bot *telebot.Bot) {
	output, _ := evalScript(c, content)
	result := fmt.Sprintf("```\n%s```", output)

	opts := telebot.SendOptions{ParseMode: telebot.ModeMarkdown}
	bot.SendMessage(message.Chat, result, &opts)
	// bot.SendMessage(message.Chat, result, nil)
}

func messages(c *config.Config, bot *telebot.Bot) {
	for message := range bot.Messages {
		var script string

		//spew.Dump(message)
		if isAuthorized(message.Sender.ID, &c.AuthorizedUsers) {
			log.Printf("Authenticated message from %s", message.Sender.Username)
		} else {
			log.Printf("==Unauthorized message from ID %d @%s==", message.Sender.ID, message.Sender.Username)
			bot.SendMessage(message.Chat, globals.Responses["unauthorized"], nil)
			continue
		}

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
				"You have activated, " + message.Sender.FirstName + "!", nil)
		} else {

		}
	}

	for message := range bot.Messages {
		log.Printf("%s sent text %s\n",
			message.Sender.Username, message.Text)
	}
}

func processQuery(c *config.Config, query telebot.Query, bot *telebot.Bot) {
	var title string
	var desc string
	var output string

	result, err := evalScript(c, &query.Text)

	if err == nil {
		title = "Execution"
	} else {
		title = "Error"
	}

	if result != "" {
		output = fmt.Sprintf("Python3\n```\n%s\n```\nOutput\n```\n%s\n```\n", query.Text, result)
	} else {
		output = fmt.Sprintf("Python3\n```\n%s\n```\n", query.Text)
		//preview := "Execution"
	}


	article := &telebot.InlineQueryResultArticle{
		Title: title,
		Description:    "foo...",
		//URL:   "http://google.com",
		InputMessageContent: &telebot.InputTextMessageContent{
			Text:           output,
			ParseMode:      "markdown",
			DisablePreview: false,
		},
	}

	results := []telebot.InlineQueryResult{article}
	response := telebot.QueryResponse{
		Results:    results,
		IsPersonal: false,
	}

	if err := bot.AnswerInlineQuery(&query, &response); err != nil {
		log.Println("Failed to respond to query", err)
	}

}

func queries(c *config.Config, bot *telebot.Bot) {
	for query := range bot.Queries {
		if isAuthorized(query.From.ID, &c.AuthorizedUsers) {
			if query.Text != "" || query.Text != " " {
				go processQuery(c, query, bot)
			}
		} else {
			log.Printf("Unauthorized query from ID %d, @%s\n", query.From.ID, query.From.Username)
		}
	}
}

func main() {
	c, err := config.New(globals.AppPrefix)
	if err != nil {
		log.Fatalln(err)
	}

	if len(c.AuthorizedUsers) == 0 {
		log.Printf("Warning: No authorized users specified. This makes the bot open to the world (dangerous)")
	}


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
