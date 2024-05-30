package schedule

type Gauntlet struct {
	player_count int
	game_number  int
}

func (g *Gauntlet) Initialize(n int) {
	g.player_count = n
	g.game_number = 0
}

func (g *Gauntlet) NextEncounter() (int, int) {
	g.game_number++
	return 0, g.game_number
}

func (g *Gauntlet) TotalEncounters() int {
	return g.player_count - 1
}
