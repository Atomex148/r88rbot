package main

import (
	"bufio"
	"log"
	"os"
	"strings"

	telego "github.com/mymmrac/telego"
)

const (
	helpNote = `KoMaHDbI:
    /pyp_xeJln, /roor_help - onucaHue KoMaHD
    /pyp_nopHo, /roor_porno - oTnpaBTb KapTuHKy c KoLLIKoMaJlm4uKoM
    /pyp_pJlaKaT, /roor_plakat - noHbITb
	/pyp_baH, /roor_ban - 3abaHuTb pypa`

	version = "1.0"
)

var (
	players PlayerStorage
	roors   []bool
)

func hasArg(target string) bool {
	for _, arg := range os.Args[1:] {
		if arg == target {
			return true
		}
	}
	return false
}

func main() {
	debugMode := false
	if hasArg("-debug") {
		debugMode = true
	}

	data, err := os.ReadFile("./token.txt")
	if err != nil {
		log.Fatal("Ошибка чтения файла с токеном: ", err)
	}

	roors = loadRoors("./roors.txt")
	defer saveRoors("./roors.txt", roors)

	players = *initPlayers("./players.json")
	defer players.Save()

	var bot *telego.Bot
	if debugMode {
		bot, err = telego.NewBot(strings.TrimSpace(string(data)), telego.WithDefaultDebugLogger())
	} else {
		bot, err = telego.NewBot(strings.TrimSpace(string(data)))
	}

	if err != nil {
		log.Fatalf("Бота с токеном %s не найдено", string(data))
	}

	params := &telego.GetUpdatesParams{
		Timeout: 30,
	}

	commands := []telego.BotCommand{
		{
			Command:     "roor_help",
			Description: "onucaHue KoMaHD",
		},
		{
			Command:     "roor_porno",
			Description: "oTnpaBTb KapTuHKy c KoLLIKoMaJlm4uKoM",
		},
		{
			Command:     "roor_plakat",
			Description: "noHbITb",
		},
		{
			Command:     "roor_ban",
			Description: "3abaHuTb pypa",
		},
	}

	err = bot.SetMyCommands(&telego.SetMyCommandsParams{Commands: commands})
	if err != nil {
		log.Printf("Ошибка при установке команд: %v", err)
	}

	updates, _ := bot.UpdatesViaLongPolling(params)
	defer bot.StopLongPolling()
	log.Printf("Рурбот v%s запущен. \"exit\" чтоб завершить", version)

	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			cmd := strings.TrimSpace(scanner.Text())
			if cmd == "exit" {
				log.Println("Завершение работы бота...")
				bot.StopLongPolling()
				return
			}
		}
	}()

	for update := range updates {
		processing(&update, bot)
	}

	log.Println("Бот остановлен")
}

func processing(update *telego.Update, bot *telego.Bot) {
	if update.Message == nil || update.Message.Text == "" || !strings.HasPrefix(update.Message.Text, "/") {
		return
	}

	isCommand := false
	if update.Message.Entities != nil {
		for _, entity := range update.Message.Entities {
			if entity.Type == "bot_command" && entity.Offset == 0 {
				isCommand = true
				break
			}
		}
	}

	if !isCommand {
		return
	}

	chatID := update.Message.Chat.ID

	command := strings.Split(update.Message.Text, " ")[0][1:]
	if atIndex := strings.Index(command, "@"); atIndex != -1 {
		command = command[:atIndex]
	}

	switch command {
	case "pyp_nopHo", "roor_porno":
		nopHo(chatID, bot)
	case "roor_help", "pyp_xeJln":
		sendText(bot, chatID, helpNote)
	case "roor_plakat", "pyp_pJlaKaT":
		if update.Message.ReplyToMessage != nil {
			plakat(chatID, update.Message.ReplyToMessage.MessageID, bot)
		} else {
			plakat(chatID, update.Message.MessageID, bot)
		}
	case "roor_ban", "pyp_baH":
		baH(update, bot)
	default:
		return
	}
}

func plakat(chatID int64, messageID int, bot *telego.Bot) {
	bot.SetMessageReaction(&telego.SetMessageReactionParams{
		ChatID:    telego.ChatID{ID: chatID},
		MessageID: int(messageID),
		Reaction:  []telego.ReactionType{&telego.ReactionTypeEmoji{Type: "emoji", Emoji: "😭"}},
	})
}
