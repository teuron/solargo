package summary

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/url"
	"solargo/config"
	"solargo/inverter"
	"solargo/persistence"

	log "github.com/sirupsen/logrus"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg/draw"
	"gonum.org/v1/plot/vg/vgimg"
)

//SendSummary sends the daily summary to the specified telegram bot
func SendSummary(config *config.Config, inverter inverter.GenericInverter, database persistence.GenericDatabase) {
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
		_, _ = http.Get(uri)

		ps, err := database.GetTodaysProduction()
		if err != nil {
			log.Warn("Could not receive todays production. Produced Error: ", err)
			return
		}

		p, err := plot.New()
		if err != nil {
			log.Warn("Could not create plot: ", err)
			return
		}

		p.X.Tick.Marker = plot.TimeTicks{Format: "15:04"}
		p.Title.Text = "Todays Production"
		p.X.Label.Text = "Time"
		p.Y.Label.Text = "Production (kWh)"
		p.Y.Min = 0
		p.Add(plotter.NewGrid())

		line, points, err := plotter.NewLinePoints(parseProductionToPlotter(ps))
		if err != nil {
			log.Warn("Could not create plot: ", err)
			return
		}
		line.Color = color.RGBA{R: 255, G: 214, A: 255}
		points.Shape = draw.CircleGlyph{}
		points.Color = color.RGBA{R: 255, G: 214, A: 255}

		p.Add(line, points)

		// Draw the plot to an in-memory image.
		img := image.NewRGBA(image.Rect(0, 0, 512, 512))
		c := vgimg.NewWith(vgimg.UseImage(img))
		p.Draw(draw.New(c))

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile("photo", "test.png")

		if err != nil {
			log.Warn("Could not create formfile: ", err)
			return
		}

		// Save the image in the multipart
		if err := png.Encode(part, c.Image()); err != nil {
			log.Warn("Could not write the image into the multipart: ", err)
			return
		}
		writer.Close()
		uri = fmt.Sprintf("%s%s/sendPhoto?chat_id=%s", summary.TelegramURL, summary.BotToken, summary.ChatID)

		r, err := http.NewRequest("POST", uri, body)
		
		if err != nil {
			log.Warn("Could not create new request", err)
			return
		}

		r.Header.Add("Content-Type", writer.FormDataContentType())
		client := &http.Client{}
		_, _ = client.Do(r)
	}
}

// randomPoints returns some random x, y points.
func parseProductionToPlotter(ps []persistence.ProductionStamps) plotter.XYs {
	pts := make(plotter.XYs, len(ps))
	for i := range pts {
		pts[i].X = float64(ps[i].Date.Unix())
		pts[i].Y = float64(ps[i].Value.ToKWh())
	}
	return pts
}
