package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand/v2"
	"os"
	"strconv"
	"strings"

	telego "github.com/mymmrac/telego"
)

type Player struct {
	ID        int64 `json:"id"`
	HasPlayed bool  `json:"hasPlayed"`
	Score     int   `json:"score"`
}

type PlayerStorage struct {
	Players  []Player `json:"players"`
	FilePath string
}

func initPlayers(filePath string) *PlayerStorage {
	storage := &PlayerStorage{[]Player{}, filePath}

	if _, err := os.Stat(filePath); err == nil {
		data, err := os.ReadFile(filePath)
		if err != nil {
			log.Printf("Ошибка при чтении файла: %v", err)
			return storage
		}

		err = json.Unmarshal(data, &storage.Players)
		if err != nil {
			log.Printf("Ошибка при разборе JSON: %v", err)
			return storage
		}
	}

	return storage
}

func loadRoors(filename string) []bool {
	data, err := os.ReadFile(filename)
	if err != nil {
		log.Panic("Ошибка загрузки руров: ", err)
	}

	s := strings.TrimSpace(string(data))
	parts := strings.Split(s, ",")

	roors := make([]bool, len(parts))
	for i, p := range parts {
		p = strings.TrimSpace(p)
		n, err := strconv.Atoi(p)
		if err != nil {
			log.Printf("Ошибка преобразования %q в число: %v", p, err)
			continue
		}
		roors[i] = n != 0
	}
	return roors
}

func saveRoors(filepath string, r []bool) {
	var builder strings.Builder
	for i, b := range r {
		if b {
			builder.WriteString("1")
		} else {
			builder.WriteString("0")
		}

		if i < len(r)-1 {
			builder.WriteString(", ")
		}
	}

	err := os.WriteFile(filepath, []byte(builder.String()), 0644)
	if err != nil {
		log.Println("Ошибка сохранения руров:", err)
	}
}

func (s *PlayerStorage) Save() error {
	data, err := json.MarshalIndent(s.Players, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.FilePath, data, 0644)
}

func (s *PlayerStorage) FindPlayer(id int64) (Player, int, bool) {
	for i, player := range s.Players {
		if player.ID == id {
			return player, i, true
		}
	}
	return Player{}, -1, false
}

func (s *PlayerStorage) AddPlayer(id int64) {
	_, _, found := s.FindPlayer(id)
	if found {
		log.Printf("Игрок с ID %d уже существует", id)
		return
	}

	s.Players = append(s.Players, Player{
		ID:        id,
		HasPlayed: false,
		Score:     0,
	})

	if err := s.Save(); err != nil {
		log.Printf("Ошибка при сохранении: %v", err)
	} else {
		log.Printf("Новый игрок с ID %d добавлен", id)
	}
}

func (s *PlayerStorage) UpdatePlayer(id int64, hasPlayed bool, newScore int) {
	_, index, found := s.FindPlayer(id)
	if !found {
		log.Printf("Игрок с ID %d не найден", id)
		return
	}

	s.Players[index].HasPlayed = hasPlayed
	s.Players[index].Score = newScore

	if err := s.Save(); err != nil {
		log.Printf("Ошибка при сохранении: %v", err)
	} else {
		log.Printf("Игрок с ID %d обновлен (Score: %d, HasPlayed: %v)", id, newScore, hasPlayed)
	}
}

func (s *PlayerStorage) GetScore(id int64) int {
	_, index, found := s.FindPlayer(id)
	if !found {
		log.Printf("Игрок с ID %d не найден", id)
		return -int(^uint(0)>>1) - 1
	}

	return s.Players[index].Score
}

func (s *PlayerStorage) CheckPlayer(id int64) (exists bool, hasPlayed bool) {
	player, _, found := s.FindPlayer(id)
	if !found {
		return false, false
	}

	return true, player.HasPlayed
}

func baH(update *telego.Update, bot *telego.Bot) {
	if update.Message.From == nil {
		log.Printf("Нет конкретного пользователя у сообщения с ID %d", update.Message.MessageID)
		return
	}

	var name string
	if update.Message.From.Username == "" {
		name = update.Message.From.FirstName
	} else {
		name = "@" + update.Message.From.Username
	}

	index := rand.IntN(len(roors))
	playerId := update.Message.From.ID

	exists, played := players.CheckPlayer(playerId)

	if !exists {
		players.AddPlayer(playerId)
	}

	if !roors[index] && !played {
		players.UpdatePlayer(playerId, !played, players.GetScore(playerId)+1)
		msg := fmt.Sprintf("%s, Tbl <b><i>3abaHuJl</i></b> pypa №%d. TBou c4eT: <b>%d</b> (+1)", name, index, players.GetScore(playerId))
		sendFormattedText(bot, update.Message.Chat.ID, msg)
	} else if roors[index] && !played {
		players.UpdatePlayer(playerId, !played, players.GetScore(playerId))
		msg := fmt.Sprintf("%s, Tbl <b><i>pa3abaHuJl</i></b> pypa №%d. TBou c4eT: <b>%d</b> (+0)", name, index, players.GetScore(playerId))
		sendFormattedText(bot, update.Message.Chat.ID, msg)
	} else {
		msg := fmt.Sprintf("%s, Tbl <b><i>y>l<e baHuJl</i></b> pypa. TBou c4eT: <b>%d</b>", name, players.GetScore(playerId))
		sendFormattedText(bot, update.Message.Chat.ID, msg)
	}

	roors[index] = !roors[index]
	saveRoors(".\\roors.txt", roors)
}
