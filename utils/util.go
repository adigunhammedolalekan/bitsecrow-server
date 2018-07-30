package utils

import (
	"net/http"
	"time"
	"math"
	"encoding/json"
	"fmt"
	"github.com/blockcypher/gobcy"
	"os"
)

func Message(status bool, message string) (map[string]interface{}) {
	return map[string]interface{} {"status" : status, "message" : message}
}

func Respond(w http.ResponseWriter, data map[string] interface{})  {
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func BC() (*gobcy.API) {
	return &gobcy.API{
		Coin: "btc", Chain: os.Getenv("bc_chain"), Token: os.Getenv("bc_token"),
	}
}

func GetReadableTime(t time.Time) (string) {

	duration := time.Since(t)
	seconds := math.Round(duration.Seconds())
	if seconds <= 59 {

		if seconds <= 1 {
			return "just now"
		}
		return fmt.Sprintf("%d seconds ago", int(seconds))
	}

	minutes := math.Round(duration.Minutes())
	if minutes <= 59 {
		return fmt.Sprintf("%d minutes ago", int(minutes))
	}

	hours := math.Round(duration.Hours())
	if hours <= 24 {
		return fmt.Sprintf("%d hours ago", int(hours))
	}

	days := hours * 30
	if days <= 730 {
		d := days / 24
		return fmt.Sprintf("%d days ago", int(math.Round(d)))
	}
	months := days / 730
	return fmt.Sprintf("%d months ago", int(math.Round(months)))
}

func ShatoshiToBtc(shatoshi float64) float64 {
	Shatoshi := 100000000.00
	return shatoshi / Shatoshi
}

func BtcToShatoshi(btc float64) float64 {
	Shatoshi := 100000000.00
	return btc * Shatoshi
}

func TempTx(from, to string, amount int) gobcy.TX {
	return gobcy.TempNewTX(from, to, amount)
}

