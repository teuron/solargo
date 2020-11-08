package summary

import (
	"net/http"
	"net/http/httptest"
	"solargo/config"
	"solargo/testutils"
	"testing"
)

func TestSendSummarySuccess(t *testing.T) {
	seen := false
	expected := "/sendmessage?chat_id=&text=Solaranlage+Statistik+Today%3A%0ADaily+Production%3A+0.00+kWh%0AYearly+Production%3A+0.00+kWh%0ATotal+Production%3A+0.00+kWh"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !seen {
			actual := r.RequestURI
			if actual != expected {
				t.Errorf("Error actual = %v, and expected = %v.", actual, expected)
			}
			seen = true
		}
	}))
	defer ts.Close()

	var iv testutils.SuccessInverter
	var db testutils.SuccessDatabase
	var config config.Config
	config.Summary.SendStatistics = true
	config.Summary.TelegramURL = ts.URL

	SendSummary(&config, &iv, &db)

}

func TestSendSummaryError(t *testing.T) {
	seen := false
	expected := "/sendmessage?chat_id=&text=Solaranlage+hat+Fehler%21+Bitte+%C3%BCberpr%C3%BCfen."
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actual := r.RequestURI
		if !seen && actual != expected {
			t.Errorf("Error actual = %v, and expected = %v.", actual, expected)

		}
		seen = true
	}))
	defer ts.Close()

	var iv testutils.ErrorInverter
	var db testutils.SuccessDatabase
	var config config.Config
	config.Summary.SendStatistics = true
	config.Summary.TelegramURL = ts.URL

	SendSummary(&config, &iv, &db)
}
