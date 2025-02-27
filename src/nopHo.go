package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"

	telego "github.com/mymmrac/telego"
)

type CircularBuffer struct {
	data  [12]int
	index int
	size  int
}

func (cb *CircularBuffer) Add(num int) {
	cb.data[cb.index] = num
	cb.index = (cb.index + 1) % len(cb.data)
	if cb.size < len(cb.data) {
		cb.size++
	}
}

func (cb *CircularBuffer) Contains(num int) bool {
	for i := 0; i < cb.size; i++ {
		if cb.data[i] == num {
			return true
		}
	}
	return false
}

type GelbooruPost struct {
	ID      int    `json:"id"`
	FileURL string `json:"file_url"`
}

type GelbooruResponse struct {
	Posts []GelbooruPost `json:"post"`
}

var alreadyWas CircularBuffer

func getUniquePost(posts []GelbooruPost) GelbooruPost {
	for {
		post := posts[rand.Intn(len(posts))]

		if alreadyWas.Contains(post.ID) {
			continue
		}

		alreadyWas.Add(post.ID)
		return post
	}
}

const maxPages = 62

func nopHo(chatId int64, bot *telego.Bot) {
	baseURL := fmt.Sprintf("https://gelbooru.com/index.php?pid=%d&json=1&limit=42&page=dapi&q=index&s=post&tags=cat_ears+yaoi", rand.Intn(maxPages))

	fmt.Println(baseURL)
	response, err := http.Get(baseURL)
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
		log.Println("Ошибка получения постов")
		return
	}

	post := getUniquePost(result.Posts)
	_, err = bot.SendPhoto(&telego.SendPhotoParams{
		ChatID: telego.ChatID{ID: chatId},
		Photo:  telego.InputFile{URL: post.FileURL},
	})
	if err != nil {
		log.Println("Ошибка отправки:", err)
	}
}
