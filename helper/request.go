package helper

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"

	"github.com/rs/zerolog"
)

func GenerateRefCode() string {
	currentUnixTime := time.Now().Unix()

	refCode := fmt.Sprintf("REF%d", currentUnixTime)

	return refCode
}

func PrintAllRequest(w http.ResponseWriter, r *http.Request, logger zerolog.Logger) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	bodyStr := string(body)
	re := regexp.MustCompile(`\s+`)
	bodyStr = re.ReplaceAllString(bodyStr, "")

	logger.Info().Any("Request Body", bodyStr).Msg("Print All Request")

}

func ParseRequestBody(r *http.Request, v interface{}, logger zerolog.Logger) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}

	bodyStr := string(body)
	re := regexp.MustCompile(`\s+`)
	bodyStr = re.ReplaceAllString(bodyStr, "")

	logger.Info().Any("Request Body", bodyStr).Msg("Print All Request")

	return json.Unmarshal([]byte(bodyStr), v)
}
