package scryfall

import (
	"accroccator/internal/model"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func CallScryfallBulkApi(url string) (*model.ScryfallBulkResponse, error) {

	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("User-Agent", "Accroccator/0.1")
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)

	//if error is different than nil, we return early because in that case the resp would
	//be a pointer to nil, which would make the program panic
	if err != nil {
		return nil, err
	}
	//this is a schedule to close the resource when the method returns.
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status when reaching scryfall %d", resp.StatusCode)
	}

	var bulkResponse model.ScryfallBulkResponse

	err = json.NewDecoder(resp.Body).Decode(&bulkResponse)

	if err != nil {
		return nil, err
	}

	return &bulkResponse, nil

}
