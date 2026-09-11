package card_mapping

import "accroccator/internal/model"

func MapOwned(scryfallCards []model.ScryfallCard, ownedCards []model.Card) []model.Card {
	ownedMap := make(map[string][]model.Card)
	for _, oc := range ownedCards {
		ownedMap[oc.ID] = append(ownedMap[oc.ID], oc)
	}

	result := make([]model.Card, 0)
	for _, sc := range scryfallCards {
		entries, isOwned := ownedMap[sc.ID]
		if isOwned {
			for _, e := range entries {
				result = append(result, model.Card{
					ScryfallCard: sc,
					OwnerShip:    e.OwnerShip,
				})
			}
		} else {
			result = append(result, model.Card{ScryfallCard: sc})
		}
	}

	return result
}
