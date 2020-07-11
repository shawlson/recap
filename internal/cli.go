package internal

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"regexp"
	"strconv"
)

type CLI struct {
	store  store
	docGen documentGenerator
}

func (cli *CLI) Initialize(database *sql.DB, recapDirectory string) {
	cli.store = store{database}
	docGen := new(documentGenerator)
	docGen.Initialize(recapDirectory)
	cli.docGen = *docGen
}

func (cli CLI) Start() {
	actions := []string{"Add game", "Edit Game", "Generate Sidebars", "Generate Indices", "Generate Site"}
	funcs := []func(){cli.addGame, cli.editGame, cli.generateSidebars, cli.generateIndices, cli.generateSite}
	funcs[promptList(actions)]()
}

func promptInt(name string) (val int) {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Printf("Enter %s: ", name)
		scanner.Scan()
		var err error
		val, err = strconv.Atoi(scanner.Text())
		if err == nil {
			return
		}
	}
}

func promptString(name string, min, max int) (val string) {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Printf("Enter %s: ", name)
		scanner.Scan()
		val = scanner.Text()
		if len(val) >= min && len(val) <= max {
			return
		}
	}
}

func promptBool(prompt string) (val bool) {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Printf("%s? (y/n): ", prompt)
		scanner.Scan()
		response := scanner.Text()
		switch response {
		case "y", "Y", "yes":
			val = true
			return
		case "n", "N", "no":
			val = false
			return
		}
	}
}

func promptList(options []string) int {
	for i, option := range options {
		fmt.Printf("[%d] %s\n", i+1, option)
	}
	scanner := bufio.NewScanner(os.Stdin)
	choice := 0
	for choice < 1 || choice > len(options) {
		fmt.Printf("Enter choice: ")
		scanner.Scan()
		choice, _ = strconv.Atoi(scanner.Text())
	}
	return choice - 1
}

func promptDate() string {
	scanner := bufio.NewScanner(os.Stdin)
	date := ""
	format := "^[0-9]{4}-[0-9]{1,2}-[0-9]{1,2}$"
	for valid := false; !valid; valid, _ = regexp.MatchString(format, date) {
		fmt.Printf("Enter date: ")
		scanner.Scan()
		date = scanner.Text()
	}
	return date
}

func (cli CLI) promptLeagues() League {
	leagues, err := cli.store.leagues()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	if len(leagues) == 0 {
		fmt.Println("No leagues")
		os.Exit(0)
	}

	leaguesStr := make([]string, len(leagues))
	for i, league := range leagues {
		leaguesStr[i] = fmt.Sprintf("%s %s", league.Name, league.Sport)
	}

	fmt.Println("Select league:")
	return leagues[promptList(leaguesStr)]
}

func (cli CLI) promptSeasons(league League) Season {
	filter := seasonFilter{}
	filter.SetLeague(league)
	seasons, err := cli.store.seasons(filter)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	if len(seasons) == 0 {
		fmt.Println("No seasons")
		os.Exit(0)
	}

	seasonsStr := make([]string, len(seasons))
	for i, season := range seasons {
		seasonsStr[i] = fmt.Sprintf("%d %s", season.Year, season.Type)
	}

	fmt.Println("Select season:")
	return seasons[promptList(seasonsStr)]
}

func (cli CLI) promptClubs(season Season) Club {
	clubs, err := cli.store.clubsBySeason(season)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	if len(clubs) == 0 {
		fmt.Println("No clubs")
		os.Exit(0)
	}

	clubsStr := make([]string, len(clubs))
	for i, club := range clubs {
		clubsStr[i] = fmt.Sprintf("%s %s", club.Represents, club.Nickname)
	}

	return clubs[promptList(clubsStr)]
}

func (cli CLI) generateGamePage(game Game) {
	resources, err := cli.store.resources(game)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	if err = cli.docGen.gamePage(game, resources); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func (cli CLI) generateClubIndex(club Club, season Season) {
	filter := gameFilter{Clubs: []Club{club}, Seasons: []Season{season}}
	var games []Game
	var err error
	if games, err = cli.store.games(filter); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	if err = cli.docGen.clubIndex(club, season, games); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}
}

func (cli CLI) generateLeagueIndex(league League) {
	filter := gameFilter{Leagues: []League{league}}
	filter.SetLimit(20)
	var games []Game
	var err error
	if games, err = cli.store.games(filter); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	if err = cli.docGen.leagueIndex(league, games); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func (cli CLI) generateIndex() {
	filter := gameFilter{}
	filter.SetLimit(20)
	var games []Game
	var err error
	if games, err = cli.store.games(filter); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	if err = cli.docGen.index(games); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func (cli CLI) addGame() {

	type dirtyClub struct {
		club   Club
		season Season
	}
	dirtyLeagues := make(map[League]bool)
	dirtyClubs := make(map[dirtyClub]bool)

	done := false
	for !done {
		league := cli.promptLeagues()
		season := cli.promptSeasons(league)
		fmt.Println("Select home team:")
		home := cli.promptClubs(season)
		fmt.Println("Select away team:")
		away := cli.promptClubs(season)
		date := promptDate()
		homeScore := promptInt("home score")
		awayScore := promptInt("away score")
		title := promptString("title", 0, 128)
		venue := promptString("venue", 0, 128)

		game, err := cli.store.createGame(season,
			date,
			home,
			homeScore,
			away,
			awayScore,
			title,
			venue)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}

		resources := make([]Resource, 0)
		doneResources := !promptBool("Add a resource")
		for !doneResources {
			title := promptString("title", 1, 128)
			url := promptString("url", 1, 256)
			resource, err := cli.store.createResource(game, title, url)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(1)
			} else {
				resources = append(resources, resource)
			}
			doneResources = !promptBool("Add a resource")
		}

		if err := cli.docGen.gamePage(game, resources); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}

		dirtyLeagues[league] = true
		dirtyClubs[dirtyClub{home, season}] = true
		dirtyClubs[dirtyClub{away, season}] = true

		done = !promptBool("Add another game")
	}

	for club := range dirtyClubs {
		cli.generateClubIndex(club.club, club.season)
	}

	for league := range dirtyLeagues {
		cli.generateLeagueIndex(league)
	}

	cli.generateIndex()
}

func (cli CLI) editGame() {

	type dirtyClub struct {
		club   Club
		season Season
	}
	dirtyLeagues := make(map[League]bool)
	dirtyClubs := make(map[dirtyClub]bool)

	done := false
	for !done {
		gameID := promptInt("game ID")
		var game Game
		var err error
		if game, err = cli.store.game(gameID); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		var resources []Resource
		if resources, err = cli.store.resources(game); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}

		fmt.Println("Game:", gameID)
		fmt.Println("Home:", game.Home.Represents, game.Home.Nickname)
		fmt.Println("Away:", game.Away.Represents, game.Away.Nickname)
		fmt.Println("Date:", game.Date)
		fmt.Println("Title:", game.Title)
		fmt.Println("Venue:", game.Venue)
		fmt.Println("Home Score:", game.HomeScore)
		fmt.Println("Away Score:", game.AwayScore)

		edit := gameEdit{}
		fields := []string{"Date", "Title", "Venue", "Home Score", "Away Score", "Resources"}

		doneEditing := false
		for !doneEditing {
			fmt.Println("Select field:")
			field := fields[promptList(fields)]
			switch field {
			case "Date":
				edit.SetDate(promptDate())
			case "Title":
				edit.SetTitle(promptString("title", 0, 128))
			case "Venue":
				edit.SetVenue(promptString("venue", 0, 128))
			case "Home Score":
				edit.SetHomeScore(promptInt("home score"))
			case "Away Score":
				edit.SetAwayScore(promptInt("away score"))
			case "Resources":
				actions := []string{"Add Resource", "Delete Resource"}
				switch actions[promptList(actions)] {
				case "Add Resource":
					title := promptString("title", 1, 128)
					url := promptString("url", 1, 256)
					_, err := cli.store.createResource(game, title, url)
					if err != nil {
						fmt.Fprintf(os.Stderr, "%v\n", err)
						os.Exit(1)
					}
				case "Delete Resource":
					resources, err := cli.store.resources(game)
					if err != nil {
						fmt.Fprintf(os.Stderr, "%v\n", err)
						os.Exit(1)
					}
					if len(resources) == 0 {
						fmt.Println("No resources to delete")
						break
					}
					resourcesStr := make([]string, len(resources))
					for i, resource := range resources {
						resourcesStr[i] = fmt.Sprintf("Title: %s\tURL: %s", resource.Title, resource.URL)
					}

					fmt.Println("Select resource to delete:")
					err = cli.store.deleteResource(resources[promptList(resourcesStr)])
					if err != nil {
						fmt.Fprintf(os.Stderr, "%v\n", err)
						os.Exit(1)
					}
				}
			}
			doneEditing = !promptBool("Continue editing")
		}

		game, err = cli.store.editGame(game, edit)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}

		resources, err = cli.store.resources(game)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}

		if err = cli.docGen.gamePage(game, resources); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}

		dirtyLeagues[game.Season.League] = true
		dirtyClubs[dirtyClub{game.Home, game.Season}] = true
		dirtyClubs[dirtyClub{game.Away, game.Season}] = true

		done = !promptBool("Edit another game")
	}

	for club := range dirtyClubs {
		cli.generateClubIndex(club.club, club.season)
	}

	for league := range dirtyLeagues {
		cli.generateLeagueIndex(league)
	}

	cli.generateIndex()
}

func (cli CLI) generateSidebars() {
	leagues, err := cli.store.leagues()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	if err = cli.docGen.indexSidebar(leagues); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	for _, league := range leagues {
		season, err := cli.store.activeSeason(league)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}

		clubs, err := cli.store.clubsByLeague(league, true)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}

		if err = cli.docGen.leagueSidebar(league, season, clubs); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}

		clubs, err = cli.store.clubsByLeague(league, false)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}

		for _, club := range clubs {
			filter := seasonFilter{}
			filter.SetClub(club)
			filter.SetLeague(league)
			seasons, err := cli.store.seasons(filter)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(1)
			}
			if err = cli.docGen.clubSidebar(club, seasons, league); err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(1)
			}
		}
	}
}

func (cli CLI) generateIndices() {
	leagues, err := cli.store.leagues()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	for _, league := range leagues {
		filter := seasonFilter{}
		filter.SetLeague(league)
		seasons, err := cli.store.seasons(filter)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}

		for _, season := range seasons {
			clubs, err := cli.store.clubsBySeason(season)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(1)
			}

			for _, club := range clubs {
				cli.generateClubIndex(club, season)
			}
		}

		cli.generateLeagueIndex(league)
	}

	cli.generateIndex()
}

func (cli CLI) generateSite() {
	games, err := cli.store.games(gameFilter{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	for _, game := range games {
		cli.generateGamePage(game)
	}

	cli.generateSidebars()
	cli.generateIndices()
}
