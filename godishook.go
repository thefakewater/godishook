package godishook

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"time"
)

func NewEmbed() *messageEmbed {
	var e messageEmbed
	return &e
}

type messageEmbed struct {
	embed embed
}

func (me *messageEmbed) SetTitle(title string) *messageEmbed {
	me.embed.Title = title
	return me
}

func (me *messageEmbed) SetAuthor(name string, url string, iconUrl string) *messageEmbed {
	me.embed.Author.Name = name
	me.embed.Author.Url = url
	me.embed.Author.IconUrl = iconUrl
	return me
}

func (me *messageEmbed) SetUrl(url string) *messageEmbed {
	me.embed.Url = url
	return me
}

func (me *messageEmbed) AddField(name string, value string, inline bool) *messageEmbed {
	var field field
	field.Name = name
	field.Value = value
	field.Inline = inline
	me.embed.Fields = append(me.embed.Fields, field)
	return me
}

func (me *messageEmbed) SetColor(color uint32) *messageEmbed {
	me.embed.Color = color
	return me
}

func (me *messageEmbed) SetThumbnail(url string) *messageEmbed {
	me.embed.Thumbnail.Url = url
	return me
}

func (me *messageEmbed) SetDescription(description string) *messageEmbed {
	me.embed.Description = description
	return me
}

func (me *messageEmbed) SetImage(url string) *messageEmbed {
	me.embed.Image.Url = url
	return me
}

func (me *messageEmbed) SetFooter(text string, iconUrl string) *messageEmbed {
	me.embed.Footer.Text = text
	me.embed.Footer.IconUrl = iconUrl
	return me
}

func (me *messageEmbed) SetTimestamp(date *time.Time) *messageEmbed {
	me.embed.Timestamp = date.Format(time.RFC3339)
	return me
}

type author struct {
	Name    string `json:"name"`
	Url     string `json:"url"`
	IconUrl string `json:"icon_url"`
}
type footer struct {
	Text    string `json:"text"`
	IconUrl string `json:"icon_url"`
}
type image struct {
	Url string `json:"url"`
}
type thumbnail struct {
	Url string `json:"url"`
}
type field struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}
type embed struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Url         string    `json:"url"`
	Color       uint32    `json:"color"`
	Author      author    `json:"author"`
	Footer      footer    `json:"footer"`
	Timestamp   string    `json:"timestamp"`
	Image       image     `json:"image"`
	Thumbnail   thumbnail `json:"thumbnail"`
	Fields      []field   `json:"fields"`
}
type Payload struct {
	Content   string  `json:"content"`
	Embeds    []embed `json:"embeds"`
	Username  string  `json:"username"`
	AvatarUrl string  `json:"avatar_url"`
}

func NewWebhook(token string) (Webhook, error) {
	var wh Webhook
	wh.Token = token
	wh.Client = &http.Client{}
	return wh, nil
}

func isUrl(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func (wh *Webhook) SetUsername(username string) error {
	if len(username) >= 80 {
		return errors.New("exceeds maximum length of 80")
	}
	wh.Username = username
	return nil
}

func (wh *Webhook) SetAvatar(avatarUrl string) error {
	if !isUrl(avatarUrl) {
		return errors.New("not a valid url")
	}
	wh.Avatar = avatarUrl
	return nil
}

func (wh *Webhook) Delete() error {
	req, _ := http.NewRequest(http.MethodDelete, wh.Token, nil)
	resp, err := wh.Client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 && resp.StatusCode <= 599 {
		return errors.New((resp.Status))
	}
	return nil
}

func initPayload(wh *Webhook, payload *Payload) {
	*&payload.Username = wh.Username
	*&payload.AvatarUrl = wh.Avatar
}

func sendPayload(wh *Webhook, payload *Payload) error {
	b, err := json.Marshal(&payload)
	if err != nil {
		return err
	}
	req, _ := http.NewRequest(http.MethodPost, wh.Token, bytes.NewReader(b))
	req.Header.Add("content-type", "application/json")
	req.Header.Add("user-agent", "disgohook")
	resp, err := wh.Client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 && resp.StatusCode <= 599 {
		return errors.New((resp.Status))
	}
	return nil
}
func (wh *Webhook) Send(content string) error {
	var payload Payload
	initPayload(wh, &payload)
	payload.Content = content
	err := sendPayload(wh, &payload)
	if err != nil {
		return err
	}
	return nil
}

type Webhook struct {
	Token    string
	Username string
	Avatar   string
	Client   *http.Client
}

func (wh *Webhook) SendEmbed(embed *messageEmbed) error {
	var payload Payload
	initPayload(wh, &payload)
	payload.Embeds = append(payload.Embeds, embed.embed)
	err := sendPayload(wh, &payload)
	if err != nil {
		return err
	}
	return nil
}

func (wh *Webhook) SendFile(filePath string) error {
	fileDir, _ := os.Getwd()
	filePath = path.Join(fileDir, filePath)
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	writer.WriteField("avatar_url", wh.Avatar)
	writer.WriteField("username", wh.Username)
	filePart, _ := writer.CreateFormFile("file", filepath.Base(file.Name()))
	io.Copy(filePart, file)
	writer.Close()
	req, _ := http.NewRequest(http.MethodPost, wh.Token, body)
	req.Header.Add("content-type", writer.FormDataContentType())
	req.Header.Add("user-agent", "disgohook")
	resp, err := wh.Client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 && resp.StatusCode <= 599 {
		return errors.New((resp.Status))
	}
	return nil
}
