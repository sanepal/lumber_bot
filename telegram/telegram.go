/*
Package telegram is a client library for the telegram bot api: https://core.telegram.org/bots/api

Only a subset of API methods are implemented
*/
package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	_ "log"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Telegram struct {
	Client    *http.Client
	BotId     string
	UserAgent string
}

// Initialize a new telegram client that performs actions on behalf of the supplied
// bot
func New(botId string, userAgent string) *Telegram {
	tr := &http.Transport{
		MaxIdleConns:    10,
		IdleConnTimeout: 30 * time.Second,
	}
	client := &http.Client{Transport: tr}

	return &Telegram{
		Client:    client,
		BotId:     botId,
		UserAgent: userAgent,
	}
}

func (b *Telegram) GetUpdates(offset int) ([]Update, error) {
	request, _ := http.NewRequest("GET", fmt.Sprintf(GET_UPDATES, b.BotId, offset), nil)
	request.Header.Set("User-Agent", b.UserAgent)

	response, err := b.Client.Do(request)
	if err != nil {
		return nil, err
	}

	body, _ := ioutil.ReadAll(response.Body)
	var result Response
	if err = json.Unmarshal([]byte(body), &result); err != nil {
		return nil, err
	}
	if !result.Ok {
		return nil, fmt.Errorf("Could not get updates (%s): %s", result.Description, string(result.Result))
	}

	var updates []Update
	if err = json.Unmarshal([]byte(result.Result), &updates); err != nil {
		return nil, err
	}

	return updates, nil
}

func (b *Telegram) SendMessage(ChatId int64, Text string, ReplyToMessageId int) error {
	request, _ := http.NewRequest("GET", fmt.Sprintf(SEND_MESSAGE, b.BotId), nil)
	request.Header.Set("User-Agent", b.UserAgent)

	q := request.URL.Query()
	q.Add("chat_id", strconv.FormatInt(ChatId, 10))
	q.Add("text", Text)
	q.Add("reply_to_message_id", strconv.Itoa(ReplyToMessageId))
	request.URL.RawQuery = q.Encode()

	response, err := b.Client.Do(request)
	if err != nil {
		return err
	}

	body, _ := ioutil.ReadAll(response.Body)
	var result Response
	if err = json.Unmarshal([]byte(body), &result); err != nil {
		return fmt.Errorf("Could not unmarshall response %s: %s", string(body), err)
	}
	if !result.Ok {
		return fmt.Errorf("Could not send message (%s): %s", result.Description, string(result.Result))
	}

	return err
}

func (t *Telegram) SetWebhook(url string, certfile string) (err error) {
	// Prepare a form that you will submit to that URL.
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	// Add file
	f, err := os.Open(certfile)
	if err != nil {
		return
	}
	defer f.Close()
	fw, err := w.CreateFormFile("certificate", certfile)
	if err != nil {
		return
	}
	if _, err = io.Copy(fw, f); err != nil {
		return
	}
	// Add the other fields
	if fw, err = w.CreateFormField("url"); err != nil {
		return
	}
	if _, err = fw.Write([]byte(url)); err != nil {
		return
	}
	if fw, err = w.CreateFormField("allowed_updates"); err != nil {
		return
	}
	if _, err = fw.Write([]byte("message")); err != nil {
		return
	}
	// Don't forget to close the multipart writer.
	// If you don't close it, your request will be missing the terminating boundary.
	w.Close()

	// Now that you have a form, you can submit it to your handler.
	req, err := http.NewRequest("POST", fmt.Sprintf(SET_WEBHOOK, t.BotId), &b)
	if err != nil {
		return
	}
	// Don't forget to set the content type, this will contain the boundary.
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("User-Agent", t.UserAgent)

	// Submit the request
	res, err := t.Client.Do(req)
	if err != nil {
		return
	}

	body, _ := ioutil.ReadAll(res.Body)
	var result Response
	if err = json.Unmarshal([]byte(body), &result); err != nil {
		return err
	}
	if !result.Ok {
		return fmt.Errorf("Could not set webhook (%s): %s", result.Description, string(result.Result))
	}

	return err
}

type Chat struct {
	Id int64 `json:"id"`
}

type Message struct {
	MessageId int    `json:"message_id"`
	Chat      Chat   `json:"chat"`
	Text      string `json:"text"`
}

type Update struct {
	UpdateId int     `json:"update_id"`
	Message  Message `json:"message"`
}

type Response struct {
	Ok          bool   `json:"ok"`
	Description string `json:"description"`
	Result      json.RawMessage
}

const (
	SET_WEBHOOK  = "https://api.telegram.org/%s/setWebhook"
	GET_UPDATES  = "https://api.telegram.org/%s/getUpdates?offset=%d&timeout=20&allowed_updates=message"
	SEND_MESSAGE = "https://api.telegram.org/%s/sendMessage"
)
