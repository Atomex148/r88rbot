package main

import (
	"log"
	"os"
	"strings"

	telego "github.com/mymmrac/telego"
	"slices"
)

const (
	helpNote = `KoMaHDbI:
    /pyp_xeJln, /roor_help - onucaHue KoMaHD
    /pyp_nopHo, /roor_porno - oTnpaBTb KapTuHKy c KoLLIKoMaJlm4uKoM
    /pyp_pJlaKaT, /roor_plakat - noHbITb
	/pyp_baH, /roor_ban - 3abaHuTb pypa
	/pyp_Ton, /roor_top - Ton baHepoB
	/pyp_bJIagocJIoBJIReT, /roor_bless - bJIagocJIoBuTb`

	version = "1.2"
)

var (
	players PlayerStorage
	roors   []bool
)

func hasArg(target string) bool {
	return slices.Contains(os.Args[1:], target)
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
	players.startDailyReset()
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
		{
			Command:     "roor_top",
			Description: "Ton baHepoB",
		},
		{
			Command:     "roor_bless",
			Description: "bJIagocJIoBuTb",
		},
	}

	err = bot.SetMyCommands(&telego.SetMyCommandsParams{Commands: commands})
	if err != nil {
		log.Printf("Ошибка при установке команд: %v", err)
	}

	updates, _ := bot.UpdatesViaLongPolling(params)
	defer bot.StopLongPolling()
	log.Printf("Рурбот v%s запущен. \"exit\" чтоб завершить, \"help\" для помощи", version)

	go cmdInput(bot)

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
	case "roor_top", "pyp_Ton":
		sendFormattedText(bot, chatID, players.getTop())
	case "roor_bless", "pyp_bJIagocJIoBJIReT":
		players.BlessPlayer(update, bot)
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
