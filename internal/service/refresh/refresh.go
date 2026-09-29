package refresh

import (
	"accroccator/internal/repository"
	"accroccator/internal/scryfall"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

func needsToUpdate(dtoUpdate *time.Time, dbUpdate *time.Time) bool {
	return dtoUpdate.After(*dbUpdate)
}

func PerformRefreshCheck(scryfallUrl string, bulkRepo *repository.BulkRepository) error {
	response, err := scryfall.CallScryfallBulkApi(scryfallUrl)

	if err != nil {
		return err
	}

	savedBulk, err := bulkRepo.FindLastSavedBulk()

	if err != nil {
		return err
	}

	if savedBulk != nil && !needsToUpdate(&response.UpdatedAt, &savedBulk.UpdatedAt) {
		log.Println("bulk data already up to date, skipping")
		return nil // already up to date: stop
	}

	err = importBulk(response.DownloadURI, bulkRepo)
	if err != nil {
		return fmt.Errorf("importing bulk: %w", err)
	}

	err = bulkRepo.CreateStagingIndexes()
	if err != nil {
		return fmt.Errorf("creating indexes: %w", err)
	}

	_, err = bulkRepo.UpdateLastBulk(*response)
	if err != nil {
		return err
	}

	return nil

}

func importBulk(url string, bulkrepo *repository.BulkRepository) error {

	//drop the previous staging
	err := bulkrepo.DropStaging()
	if err != nil {
		return err
	}

	body, err := scryfall.StartDownload(url)
	if err != nil {
		return err
	}
	defer body.Close()

	gz, err := gzip.NewReader(body)
	if err != nil {
		return err
	}
	defer gz.Close()

	dec := json.NewDecoder(gz)
	count := 0
	batchSlice := []any{}
	for {
		var card bson.M
		err := dec.Decode(&card)

		if err == io.EOF {
			//end of file
			break
		}

		if err != nil {
			return err
		}

		batchSlice = append(batchSlice, card)
		if len(batchSlice) == 10000 {
			err = bulkrepo.InsertBatches(batchSlice)
			if err != nil {
				return err
			}
			log.Printf("inserted %d cards", count)

			batchSlice = nil
		}

		count++

	}

	if len(batchSlice) > 0 {
		err = bulkrepo.InsertBatches(batchSlice)
		if err != nil {
			return err
		}
	}

	log.Printf("the count of cards is %d", count)

	return nil
}
