package refresh

import (
	"accroccator/internal/repository"
	"accroccator/internal/scryfall"
	"log"
	"time"
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

	// import the bulk file — not built yet

	_, err = bulkRepo.UpdateLastBulk(*response)
	if err != nil {
		return err
	}

	return nil

}
