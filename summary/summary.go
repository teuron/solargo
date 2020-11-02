package summary

import (
	"fmt"
	"net/http"
	"net/url"
	"solargo/config"
	"solargo/inverter"

	log "github.com/sirupsen/logrus"
)

//SendSummary sends the daily summary to the specified telegram bot
func SendSummary(config *config.Config, inverter inverter.GenericInverter) {
	var message string
	summary := config.Summary

	if summary.SendStatistics {
		statistics, err := inverter.GetInverterStatistics()
		if err != nil {
			log.Warn("Could not receive daily statistics. Produced Error: ", err)
			message = "Solaranlage hat Fehler! Bitte überprüfen."
		} else {
			message = fmt.Sprintf("Solaranlage Statistik Today:\n%s", statistics.String())
		}

		message = url.QueryEscape(message)

		uri := fmt.Sprintf("%s%s/sendmessage?chat_id=%s&text=%s", summary.TelegramURL, summary.BotToken, summary.ChatID, message)
		http.Get(uri)

	}
}
