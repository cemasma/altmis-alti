package game

type Player struct {
	Id          string
	Cards       []Card
	TurnScore   uint8
	TotalScore  uint8
	Turn        bool
	matchPoints uint8
}

func (player *Player) Move(kind, cardName, groundKind string, isGameModeOff, isFirst bool, isItLastTurn bool) bool {
	_, _, exists := player.IsPlayerHaveCard(kind, cardName)

	if !player.Turn {
		return false
	}

	if !exists {
		return false
	}

	kindExists := player.isPlayerHaveKind(groundKind)

	if isGameModeOff && !isFirst && kind != groundKind && kindExists {
		return false
	}

	if isItLastTurn && kind != groundKind && kindExists {
		return false
	}

	return true
}

func (player Player) isPlayerHaveKind(kind string) (exists bool) {
	for _, value := range player.Cards {

		if value.Kind == kind {
			exists = true
			break
		}
	}

	return
}

func (player Player) IsPlayerHaveCard(kind, cardName string) (card Card, index int, exists bool) {
	for i, value := range player.Cards {

		if value.Kind == kind && value.CardName == cardName {
			exists = true
			card = value
			index = i
			break
		}
	}

	return
}

func (player *Player) MoveCard(index int, g *Game) {
	g.Ground = append(g.Ground, player.Cards[index])
	player.Cards = append(player.Cards[:index], player.Cards[index+1:]...)
	g.PlayerMap[player.Id] = *player
}
