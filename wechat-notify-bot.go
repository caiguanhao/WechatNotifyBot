package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

// github.com/caiguanhao/wechat-notify
type Input struct {
	Timestamp   int64
	Service     string
	Event       string
	Action      string
	Host        string
	Description string
	URL         string
}

const tpl = "`{{.Host}}`\n\n```text\n{{.Description | html}}\n```\n{{if .URL}}[{{.Action}}]({{.URL}}) / {{end}}{{if .Timestamp}}{{.Timestamp | format}}{{end}}"

func (input Input) String() string {
	t := template.Must(template.New("content").Funcs(template.FuncMap{
		"format": func(sec int64) string {
			return time.Unix(sec, 0).Format("2006-01-02 15:04:05")
		},
		"html": func(input string) template.HTML {
			return template.HTML(input)
		},
	}).Parse(tpl))
	var b bytes.Buffer
	err := t.Execute(&b, input)
	if err == nil {
		return b.String()
	}
	return err.Error()
}

// github.com/caiguanhao/wechat-notify
func parse(input []byte) *Input {
	scanner := bufio.NewScanner(bytes.NewReader(bytes.TrimSpace(input)))
	var ret Input
	isDesc := false
	for scanner.Scan() {
		line := scanner.Text()
		if !isDesc && len(line) == 0 {
			isDesc = true
			continue
		}
		if isDesc {
			ret.Description = ret.Description + line + "\n"
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			ret.Description = ret.Description + line + "\n"
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		switch key {
		case "timestamp":
			ret.Timestamp, _ = strconv.ParseInt(value, 10, 64)
		case "service":
			ret.Service = value
		case "event":
			ret.Event = value
		case "action":
			ret.Action = value
		case "host":
			ret.Host = value
		case "url":
			ret.URL = value
		}
		ret.Description = ""
	}
	ret.Description = strings.TrimSpace(ret.Description)
	return &ret
}

func main() {
	flag.Parse()
	if flag.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "Please provide at least one Telegram chat ID.")
		os.Exit(1)
	}
	if PROXY != "" {
		err := os.Setenv("HTTP_PROXY", PROXY)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	stdin, stdinErr := ioutil.ReadAll(os.Stdin)
	if stdinErr != nil {
		fmt.Fprintln(os.Stderr, stdinErr)
		os.Exit(1)
	}
	input := parse(stdin)

	bot, err := tgbotapi.NewBotAPI(BOTAPI)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	bot.Debug = false
	hasError := false
	for _, arg := range flag.Args() {
		parts := strings.SplitN(arg, "@", 2)
		chatId, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			hasError = true
			continue
		}
		msg := tgbotapi.NewMessage(chatId, input.String())
		msg.ParseMode = "Markdown"
		_, err = bot.Send(msg)
		if err != nil {
			fmt.Fprintln(os.Stderr, chatId, err)
			hasError = true
		}
	}

	if hasError {
		os.Exit(1)
	}
}
