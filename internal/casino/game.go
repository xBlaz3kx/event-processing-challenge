package casino

var Games = map[int]Game{
	100: {Title: "Rocket Dice"},
	101: {Title: "It's bananas!"},
	102: {Title: "Wild Spin"},
	103: {Title: "Book of Dead"},
	104: {Title: "Pirate Jackpots"},
	105: {Title: "Western Gold 2"},
	106: {Title: "Super Rainbow Megaways"},
	107: {Title: "#BarsAndBells"},
	108: {Title: "Fortune Three"},
	109: {Title: "ChilliPop"},
}

type Game struct {
	Title string
}
