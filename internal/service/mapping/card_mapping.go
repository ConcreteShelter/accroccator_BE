package card_mapping

import "accroccator/internal/model"

func MapOwned(scryfallCards []model.ScryfallCard, ownedCards []model.OwnedCardReference) []model.Card {
	//let's build the blank map for later
	ownedMap := make(map[string]model.OwnedCardReference)

	for _, ownedCard := range ownedCards {
		ownedMap[ownedCard.ScryfallID] = ownedCard
	}

	mappedScryfallCards := make([]model.Card, 0)

	for _, scryfallCard := range scryfallCards {
		owned, ok := ownedMap[scryfallCard.ID]
		if ok {
			mappedScryfallCards = append(mappedScryfallCards, model.Card{
				ScryfallCard: scryfallCard,
				ContainedIn:  owned.ContainedIn,
			})
		} else {
			mappedScryfallCards = append(mappedScryfallCards, model.Card{
				ScryfallCard: scryfallCard,
				ContainedIn:  []model.InventoryOwnership{},
			})
		}
	}

	return mappedScryfallCards

}
