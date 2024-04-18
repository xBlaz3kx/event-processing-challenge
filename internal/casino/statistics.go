package casino

type Statistics struct {
	EventsTotal                  int               `json:"events_total"`
	EventsPerMinute              float64           `json:"events_per_minute"`
	EventsPerSecondMovingAverage float64           `json:"events_per_second_moving_average"`
	TopPlayerBets                TopPlayerBets     `json:"top_player_bets"`
	TopPlayerWins                TopPlayerWins     `json:"top_player_wins"`
	TopPlayerDeposits            TopPlayerDeposits `json:"top_player_deposits"`
}

type TopPlayerBets struct {
	Id    int `json:"id"`
	Count int `json:"count"`
}

type TopPlayerWins struct {
	Id    int `json:"id"`
	Count int `json:"count"`
}

type TopPlayerDeposits struct {
	Id    int `json:"id"`
	Count int `json:"count"`
}
