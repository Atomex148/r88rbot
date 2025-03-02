package main

import (
	"fmt"
	"log"
	"math/rand/v2"

	telego "github.com/mymmrac/telego"
)

func play(displayName string, player *Player) string {
	var msg string
	log.Println("Всего забаненых руров: ", players.BannedPypc)
	attemps := (players.BannedPypc / (len(roors) / 5)) + 1
	log.Println("Попыток: ", attemps)

	index := rand.IntN(len(roors))
	for range attemps {
		if !roors[index] {
			players.UpdatePlayer(player.ID, !player.HasPlayed, players.GetScore(player.ID)+1)
			msg = fmt.Sprintf("%s, Tbl <b><i>3abaHuJl</i></b> pypa №%d. TBou c4eT: <b>%d</b> (+1)", displayName, index+1, players.GetScore(player.ID))
			roors[index] = !roors[index]
			return msg
		}
		index = rand.IntN(len(roors))
	}

	players.UpdatePlayer(player.ID, !player.HasPlayed, players.GetScore(player.ID))
	msg = fmt.Sprintf("%s, Tbl <b><i>pa3abaHuJl</i></b> pypa №%d. TBou c4eT: <b>%d</b> (+0)", displayName, index+1, players.GetScore(player.ID))
	roors[index] = !roors[index]
	return msg
}

func baH(update *telego.Update, bot *telego.Bot) {
	if update.Message.From == nil {
		log.Printf("Нет конкретного пользователя у сообщения с ID %d", update.Message.MessageID)
		return
	}

	user := update.Message.From
	name := user.FirstName
	playerId := user.ID

	var displayName string
	if user.Username != "" {
		displayName = "@" + user.Username
	} else {
		displayName = name
	}

	exists, played := players.CheckPlayer(playerId)

	var player Player
	if !exists {
		player = *players.AddPlayer(playerId, name)
	} else {
		player, _, _ = players.FindPlayer(playerId)
	}

	var msg string
	if played {
		msg = fmt.Sprintf("%s, Tbl <b><i>y}I{e baHuJl</i></b> pypa. TBou c4eT: <b>%d</b>",
			displayName, players.GetScore(playerId))
	} else {
		msg = play(displayName, &player)
	}

	sendFormattedText(bot, update.Message.Chat.ID, msg)

	players.BannedPypc = updateBannedCount(roors)

	if players.BannedPypc == len(roors) {
		msgWin := players.getWinner()
		msgWin += "\n--------------------\n" + players.getTop()
		players.resetAllPlayers()

		sendFormattedText(bot, update.Message.Chat.ID, msgWin)
	}

	saveRoors(".\\roors.txt", roors)
}
