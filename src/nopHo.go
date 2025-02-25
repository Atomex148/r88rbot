package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"time"

	telego "github.com/mymmrac/telego"
)

type GelbooruPost struct {
	ID      int    `json:"id"`
	FileURL string `json:"file_url"`
}

type GelbooruResponse struct {
	Posts []GelbooruPost `json:"post"`
}

func nopHo(chatId int64, bot *telego.Bot) {
	baseURL := "https://gelbooru.com/index.php"
	params := url.Values{}
	params.Set("page", "dapi")
	params.Set("s", "post")
	params.Set("q", "index")
	params.Set("tags", "cat_ears 1boy -1girl -hetero boys_only")
	params.Set("sort", "random")
	params.Set("limit", "100")
	params.Set("json", "1")

	rand.New(rand.NewSource(time.Now().UnixNano()))
	params.Set("after_id", strconv.Itoa(rand.Intn(5000000)+1))

	response, err := http.Get(baseURL + "?" + params.Encode())
	if err != nil {
		log.Println("Ошибка запроса:", err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusTooManyRequests {
		log.Println("Слишком много запросов")
		return
	}

	var result GelbooruResponse
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		log.Println("Ошибка парсинга JSON:", err)
		return
	}

	if len(result.Posts) == 0 {
		log.Println("Нет результатов, пробуем без after_id")
		params.Del("after_id")
		nopHo(chatId, bot)
		return
	}

	post := result.Posts[rand.Intn(len(result.Posts))]

	_, err = bot.SendPhoto(&telego.SendPhotoParams{
		ChatID: telego.ChatID{ID: chatId},
		Photo:  telego.InputFile{URL: post.FileURL},
	})
	if err != nil {
		log.Println("Ошибка отправки:", err)
	}
}
