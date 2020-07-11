package internal

type Sport struct {
	ID   int
	Name string
}

type League struct {
	Sport string
	Code  string
	Name  string
}

type Season struct {
	ID         int
	League     League
	Year       int
	Type       string
	Exhibition bool
}

type Club struct {
	ID         int
	Iteration  int
	Represents string
	Nickname   string
}

type Game struct {
	ID        int
	Season    Season
	Date      string
	Title     string
	Venue     string
	Home      Club
	HomeScore int
	Away      Club
	AwayScore int
}

type Resource struct {
	ID    int
	Title string
	URL   string
}
