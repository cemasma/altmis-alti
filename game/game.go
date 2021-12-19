package game

type Game struct {
	Id            string
	PlayerMap     map[string]Player
	Cards         []Card
	Turn          int
	TotalTurn     int
	Ace           string
	Ground        []Card
	IsGameOff     bool
	offPlayerId   string
	firstPlayerId string
}

func NewGame(gameId string, player1Id, player2Id string) (game Game) {
	cards := PrepareCards()

	game.Id = gameId
	game.Ace = cards[0].Kind
	game.Turn = 0
	game.TotalTurn = 0

	player1Cards, cards := cards[len(cards)-6:], cards[:len(cards)-6]
	player2Cards, cards := cards[len(cards)-6:], cards[:len(cards)-6]

	game.Cards = cards

	player1 := Player{
		Id:         player1Id,
		Cards:      player1Cards,
		TurnScore:  0,
		TotalScore: 0,
		Turn:       true,
	}

	player2 := Player{
		Id:         player2Id,
		Cards:      player2Cards,
		TurnScore:  0,
		TotalScore: 0,
		Turn:       false,
	}
	game.PlayerMap = make(map[string]Player)

	game.PlayerMap[player1Id] = player1
	game.PlayerMap[player2Id] = player2

	game.Ground = make([]Card, 0)
	game.IsGameOff = false
	game.firstPlayerId = player1Id

	return
}

func (game *Game) NewTurn() {
	cards := PrepareCards()

	game.Ace = cards[0].Kind
	game.Turn = 0
	game.TotalTurn += 1

	player1Cards, cards := cards[len(cards)-6:], cards[:len(cards)-6]
	player2Cards, cards := cards[len(cards)-6:], cards[:len(cards)-6]

	game.Cards = cards

	for _, player := range game.PlayerMap {
		if player.Id == game.firstPlayerId {
			player.Cards = player1Cards
			player.Turn = game.TotalTurn%2 == 0
			player.TurnScore = 0
		} else {
			player.Cards = player2Cards
			player.Turn = game.TotalTurn%2 == 1
			player.TurnScore = 0
		}

		game.PlayerMap[player.Id] = player
	}

	game.Ground = make([]Card, 0)
	game.IsGameOff = false
	game.offPlayerId = ""
}

func (game *Game) CalculatePoints(player1Id, player2Id string) {
	points := make(map[string]int8)
	player1 := game.PlayerMap[player1Id]
	player2 := game.PlayerMap[player2Id]

	f := func(p1, p2 Player) (Player, Player) {
		p1.TotalScore += 1

		if p2.TurnScore == 0 {
			p1.TotalScore += 1
		}

		if p2.TurnScore < 33 {
			p1.TotalScore += 1
		}
		return p1, p2
	}

	if !game.IsGameOff {
		if player1.TurnScore < 66 && player2.TurnScore < 66 {
			points[player1Id] = 0
			points[player2Id] = 0
		} else if player1.TurnScore >= 66 {
			player1, player2 = f(player1, player2)
		} else if player2.TurnScore >= 66 {
			player2, player1 = f(player2, player1)
		}
	} else {
		if game.offPlayerId == player1Id {
			if player1.TurnScore < 66 {
				player2.TotalScore += 2
			} else {
				player1, player2 = f(player1, player2)
			}
		} else {
			if player2.TurnScore < 66 {
				player1.TotalScore += 2
			} else {
				player2, player1 = f(player2, player1)
			}
		}
	}

	game.PlayerMap[player1Id] = player1
	game.PlayerMap[player2Id] = player2
}

func (g *Game) CalculateTurnPoints(player, otherPlayer Player) {
	card1, card2 := g.Ground[0], g.Ground[1]

	if card1.CardName == "QUEEN" {
		for _, value := range otherPlayer.Cards {
			if value.CardName == "KING" && card1.Kind == value.Kind {
				var score uint8
				score = 20

				if card1.Kind == g.Ace {
					score += 20
				}

				if otherPlayer.TurnScore > 0 {
					otherPlayer.TurnScore += score
				} else {
					otherPlayer.matchPoints += score
				}
			}
		}
	}

	if card1.Kind == card2.Kind {
		if card1.Value > card2.Value {
			otherPlayer.TurnScore += card1.Value + card2.Value
			player.Turn = false
			otherPlayer.Turn = true
		} else {
			player.TurnScore += card1.Value + card2.Value
			player.Turn = true
			otherPlayer.Turn = false
		}
	} else if card2.Kind == g.Ace {
		player.TurnScore += card1.Value + card2.Value
		player.Turn = true
		otherPlayer.Turn = false
	} else {
		otherPlayer.TurnScore += card1.Value + card2.Value
		player.Turn = false
		otherPlayer.Turn = true
	}

	if player.TurnScore > 0 {
		player.TurnScore += player.matchPoints
		player.matchPoints = 0
	}

	if otherPlayer.TurnScore > 0 {
		otherPlayer.TurnScore += otherPlayer.matchPoints
		otherPlayer.matchPoints = 0
	}

	if len(g.Cards) >= 2 && !g.IsGameOff {
		newCard1, newCard2 := g.Cards[len(g.Cards)-1], g.Cards[len(g.Cards)-2]
		g.Cards = g.Cards[:len(g.Cards)-2]

		otherPlayer.Cards = append(otherPlayer.Cards, newCard1)
		player.Cards = append(player.Cards, newCard2)
	}

	g.PlayerMap[player.Id] = player
	g.PlayerMap[otherPlayer.Id] = otherPlayer
	g.Ground = make([]Card, 0)
	g.Turn += 1
}

func (g Game) IsTurnEnded(player Player) bool {
	return len(player.Cards) == 0 && (len(g.Cards) == 0 || g.IsGameOff)
}

func (g *Game) SwitchTurns(playerId, otherPlayerId string) {
	player := g.PlayerMap[playerId]
	otherPlayer := g.PlayerMap[otherPlayerId]

	player.Turn = false
	otherPlayer.Turn = true

	g.PlayerMap[playerId] = player
	g.PlayerMap[otherPlayerId] = otherPlayer
}

func (g *Game) GameOff(playerId string) {
	g.IsGameOff = true
	g.offPlayerId = playerId
}

func (g Game) GetGroundKind() (groundKind string) {
	if len(g.Ground) > 0 {
		groundKind = g.Ground[len(g.Ground)-1].Kind
	}

	return
}

func (g Game) IsGroundCardsOver() bool {
	return len(g.Cards) == 0
}

func (g Game) IsSecondCardPlayed() bool {
	return len(g.Cards) == 2
}
