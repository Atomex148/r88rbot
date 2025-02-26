package main

import (
	"bufio"
	"log"
	"os"
	"strings"

	telego "github.com/mymmrac/telego"
)

const cmdHelpNote = `
"exit" - выключить бота
"reset_players_timer" - принудительно сбросить таймер игроков 
"reset_players" - сбросить ВСЕ статусы игроков
`

func cmdInput(bot *telego.Bot) {
	scanner := bufio.NewScanner(os.Stdin)
	var pendingResetTimer, pendingResetPlayers bool

	for {
		if !scanner.Scan() {
			return
		}

		cmd := strings.TrimSpace(scanner.Text())

		if pendingResetTimer {
			switch strings.ToLower(cmd) {
			case "y", "yes", "д", "да":
				log.Println("Подтверждение получено. Сбрасываю состояния...")
				players.resetPlayers()
			default:
				log.Println("Сброс отменён")
			}
			pendingResetTimer = false
			continue
		}

		if pendingResetPlayers {
			switch strings.ToLower(cmd) {
			case "y", "yes", "д", "да":
				log.Println("Подтверждение получено. Полный сброс состояния...")
				players.resetAllPlayers()
			default:
				log.Println("Сброс отменён")
			}
			pendingResetPlayers = false
			continue
		}

		switch cmd {
		case "exit":
			log.Println("Завершение работы бота...")
			bot.StopLongPolling()
			return

		case "reset_players_timer":
			log.Print("Подтвердите сброс [y/yes/д/да]: ")
			pendingResetTimer = true

		case "reset_players":
			log.Print("Подтвердите сброс [y/yes/д/да]: ")
			pendingResetPlayers = true

		case "help":
			log.Print(cmdHelpNote)

		default:
			log.Println("Неизвестная команда:", cmd)
		}
	}
}
