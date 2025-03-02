package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mymmrac/telego"
	"slices"
)

type Player struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	HasPlayed bool   `json:"hasPlayed"`
	Score     int    `json:"score"`
}

type PlayerStorage struct {
	Players     []Player  `json:"players"`
	Admins      []int64   `json:"adminsID"`
	LastUpdated time.Time `json:"lastUpdated"`
	BannedPypc  int       `json:"bannedPypc"`
	FilePath    string
	mu          sync.RWMutex
}

func initPlayers(filePath string) *PlayerStorage {
	storage := &PlayerStorage{
		Admins:      []int64{5724528655, 359844192},
		FilePath:    filePath,
		LastUpdated: time.Now(),
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		if err := storage.Save(); err != nil {
			log.Printf("Ошибка при создании файла игроков: %v", err)
		}
		log.Println("Создан новый файл конфигураций игроков")
		return storage
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("Ошибка при чтении файла: %v", err)
		return storage
	}

	if err = json.Unmarshal(data, storage); err != nil {
		log.Printf("Ошибка при разборе JSON: %v", err)
	}

	if time.Since(storage.LastUpdated) > 24*time.Hour {
		storage.resetPlayers()
	}

	return storage
}

func loadRoors(filename string) []bool {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		initialData := strings.Repeat("0, ", 99) + "0"
		if err := os.WriteFile(filename, []byte(initialData), 0644); err != nil {
			log.Panicf("Ошибка создания файла руров: %v", err)
		}
		log.Println("Создан новый файл 100шт руров")
	}

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

func updateBannedCount(r []bool) int {
	count := 0
	for _, b := range r {
		if b {
			count++
		}
	}
	return count
}

func (s *PlayerStorage) Save() error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.FilePath, data, 0644)
}

func (s *PlayerStorage) resetPlayers() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.Players {
		s.Players[i].HasPlayed = false
	}
	s.LastUpdated = time.Now()
	log.Println("Автоматический сброс статусов игроков")
	s.Save()
}

func (s *PlayerStorage) resetAllPlayers() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.Players {
		s.Players[i].HasPlayed = false
		s.Players[i].Score = 0
	}
	s.LastUpdated = time.Now()
	s.BannedPypc = 0

	if err := s.Save(); err != nil {
		log.Printf("Ошибка при полном сбросе: %v", err)
	} else {
		log.Println("Полный сброс игроков выполнен")
	}

	for i := 0; i < len(roors); i++ {
		roors[i] = false
	}
}

func (s *PlayerStorage) startDailyReset() {
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for range ticker.C {
			now := time.Now()
			if now.Hour() == 21 && now.Minute() == 0 {
				s.resetPlayers()
			}
		}
	}()
}

func (s *PlayerStorage) FindPlayer(id int64) (Player, int, bool) {
	for i, player := range s.Players {
		if player.ID == id {
			return player, i, true
		}
	}
	return Player{}, -1, false
}

func (s *PlayerStorage) AddPlayer(id int64, name string) *Player {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, _, found := s.FindPlayer(id)
	if found {
		log.Printf("Игрок с ID %d уже существует", id)
		return nil
	}

	newPlayer := Player{
		ID:        id,
		Name:      name,
		HasPlayed: false,
		Score:     0,
	}
	s.Players = append(s.Players, newPlayer)

	addedPlayer := &s.Players[len(s.Players)-1]

	if err := s.Save(); err != nil {
		log.Printf("Ошибка при сохранении: %v", err)
	} else {
		log.Printf("Новый игрок с ID %d добавлен", id)
	}

	return addedPlayer
}

func (s *PlayerStorage) UpdatePlayer(id int64, hasPlayed bool, newScore int) {
	s.mu.Lock()
	defer s.mu.Unlock()

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
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, index, found := s.FindPlayer(id)
	if !found {
		return -1
	}
	return s.Players[index].Score
}

func (s *PlayerStorage) getTop() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	playersTop := make([]Player, len(s.Players))
	copy(playersTop, s.Players)

	sort.Slice(playersTop, func(i, j int) bool {
		return playersTop[i].Score > playersTop[j].Score
	})

	if len(playersTop) == 0 {
		return "<b>PeuTuHG nycT</b>"
	}

	var msgBuilder strings.Builder
	msgBuilder.WriteString("Топ baHepoB:\n")
	for i, p := range playersTop {
		if p.Score != 0 {
			msgBuilder.WriteString(
				fmt.Sprintf("\n<b>%d</b>. <b>%s: %d o4KoB</b>", i+1, p.Name, p.Score),
			)
		}
	}
	msg := msgBuilder.String()
	return msg
}

func (s *PlayerStorage) getWinner() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	playersTop := make([]Player, len(s.Players))
	copy(playersTop, s.Players)

	sort.Slice(playersTop, func(i, j int) bool {
		return playersTop[i].Score > playersTop[j].Score
	})

	return fmt.Sprintf("<b>%s nobeDul, 3abanuB boJlbwe Bcego pypoB. OH 3abaHuJI %d pypoB!</b>", playersTop[0].Name, playersTop[0].Score)
}

func (s *PlayerStorage) CheckPlayer(id int64) (exists bool, hasPlayed bool) {
	s.mu.RLock()
	needsReset := time.Since(s.LastUpdated) > 24*time.Hour
	s.mu.RUnlock()

	if needsReset {
		s.mu.Lock()
		if time.Since(s.LastUpdated) > 24*time.Hour {
			s.resetPlayers()
		}
		s.mu.Unlock()
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	player, _, found := s.FindPlayer(id)
	return found, player.HasPlayed
}

func (s *PlayerStorage) BlessPlayer(update *telego.Update, bot *telego.Bot) {
	botUser, err := bot.GetMe()
	if err != nil {
		log.Printf("Ошибка получения информации о боте: %v", err)
		return
	}

	if update.Message == nil ||
		update.Message.ReplyToMessage == nil ||
		update.Message.ReplyToMessage.From == nil ||
		update.Message.ReplyToMessage.From.ID == botUser.ID {
		return
	}

	senderID := update.Message.From.ID
	s.mu.RLock()
	isAdmin := slices.Contains(s.Admins, senderID)
	s.mu.RUnlock()

	if !isAdmin {
		msg := "Bbl He uMeeTe npaBa bJIagocJIoBuTb"
		sendText(bot, update.Message.Chat.ID, msg)
		return
	}

	index := 1
	for i := range roors {
		if !roors[i] {
			roors[i] = true
			index = i + 1
			break
		}
	}
	s.BannedPypc = updateBannedCount(roors)

	playerId, playerName := update.Message.ReplyToMessage.From.ID, update.Message.ReplyToMessage.From.FirstName
	exist, hasPlayed := s.CheckPlayer(update.Message.ReplyToMessage.From.ID)

	var msg string
	if !exist {
		newPlayer := s.AddPlayer(playerId, playerName)
		newPlayer.Score = 1
		msg = fmt.Sprintf("<b>AmopaJl</b> bJIagoCJIoBuJI <b>%s</b>, <b><i>3abaHuB</i></b> pypa №%d. C4eT: <b>1</b>", playerName, index)
		sendFormattedText(bot, update.Message.Chat.ID, msg)
		return
	}

	currentScore := s.GetScore(playerId)
	s.UpdatePlayer(playerId, hasPlayed, currentScore+1)
	msg = fmt.Sprintf("<b>AmopaJl</b> bJIagoCJIoBuJI <b>%s</b>, <b><i>3abaHuB</i></b> pypa №%d. C4eT: <b>%d</b>", playerName, index, currentScore+1)
	sendFormattedText(bot, update.Message.Chat.ID, msg)

	if s.BannedPypc == len(roors) {
		msgWin := s.getWinner()
		msgWin += "\n--------------------\n" + s.getTop()
		s.resetAllPlayers()

		sendFormattedText(bot, update.Message.Chat.ID, msgWin)
	}

	log.Printf("Благословлен игрок %s, ID: %d", playerName, playerId)
	saveRoors(".\\roors.txt", roors)
}
