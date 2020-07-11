package internal

import (
	"database/sql"
	"fmt"
	"strings"
)

type store struct {
	*sql.DB
}

type filter struct {
	limit    int
	limitSet bool
}

func (f filter) Limit() (int, bool) {
	return f.limit, f.limitSet
}

func (f *filter) SetLimit(n int) {
	f.limit = n
	f.limitSet = true
}

type gameFilter struct {
	filter
	IDs     []int
	Clubs   []Club
	Seasons []Season
	Leagues []League
}

type seasonFilter struct {
	filter
	league    League
	leagueSet bool
	club      Club
	clubSet   bool
}

func (sf seasonFilter) League() (League, bool) {
	return sf.league, sf.leagueSet
}

func (sf *seasonFilter) SetLeague(l League) {
	sf.league = l
	sf.leagueSet = true
}

func (sf seasonFilter) Club() (Club, bool) {
	return sf.club, sf.clubSet
}

func (sf *seasonFilter) SetClub(c Club) {
	sf.club = c
	sf.clubSet = true
}

type gameEdit struct {
	date         string
	title        string
	venue        string
	homeScore    int
	awayScore    int
	dateSet      bool
	titleSet     bool
	venueSet     bool
	homeScoreSet bool
	awayScoreSet bool
}

func (ge *gameEdit) SetDate(date string) {
	ge.date = date
	ge.dateSet = true
}

func (ge *gameEdit) SetTitle(title string) {
	ge.title = title
	ge.titleSet = true
}

func (ge *gameEdit) SetVenue(venue string) {
	ge.venue = venue
	ge.venueSet = true
}

func (ge *gameEdit) SetHomeScore(score int) {
	ge.homeScore = score
	ge.homeScoreSet = true
}

func (ge *gameEdit) SetAwayScore(score int) {
	ge.awayScore = score
	ge.awayScoreSet = true
}

func (ge gameEdit) Date() (string, bool) {
	return ge.date, ge.dateSet
}

func (ge gameEdit) Title() (string, bool) {
	return ge.title, ge.titleSet
}

func (ge gameEdit) Venue() (string, bool) {
	return ge.venue, ge.venueSet
}

func (ge gameEdit) HomeScore() (int, bool) {
	return ge.homeScore, ge.homeScoreSet
}

func (ge gameEdit) AwayScore() (int, bool) {
	return ge.awayScore, ge.awayScoreSet
}

func (store store) league(code string) (League, error) {
	q :=
		`
		SELECT league_code, sport_name, league_name
		FROM league
		NATURAL JOIN sport
		WHERE league_code = upper($1)
		`
	var league League
	err := store.
		QueryRow(q, code).
		Scan(&league.Code, &league.Sport, &league.Name)

	return league, err
}

func (store store) leagues() ([]League, error) {
	q :=
		`
		SELECT league_code, sport_name, league_name
		FROM league
		NATURAL JOIN sport
		ORDER BY league_code asc
		`
	var leagues []League
	rows, err := store.Query(q)
	if err != nil {
		return leagues, err
	}
	defer rows.Close()

	for rows.Next() {
		var league League
		err = rows.Scan(&league.Code, &league.Sport, &league.Name)
		if err != nil {
			return leagues, err
		}
		leagues = append(leagues, league)
	}

	if rows.Err() != nil {
		return leagues, rows.Err()
	}

	return leagues, nil
}

func (store store) seasons(filter seasonFilter) ([]Season, error) {
	var seasons []Season
	q, args := seasonQuery(filter)
	rows, err := store.Query(q, args...)
	if err != nil {
		return seasons, err
	}
	defer rows.Close()

	for rows.Next() {
		var season Season
		err = rows.Scan(&season.ID, &season.League.Sport, &season.League.Code,
			&season.League.Name, &season.Year, &season.Type, &season.Exhibition)
		if err != nil {
			return seasons, err
		}
		seasons = append(seasons, season)
	}

	if rows.Err() != nil {
		return seasons, rows.Err()
	}

	return seasons, nil
}

func (store store) activeSeason(league League) (Season, error) {
	var season Season
	q :=
		`
		SELECT
			season_id, sport_name, league_code, league_name,
			start_year, season_type, exhibition
		FROM sport
		NATURAL JOIN league
		NATURAL JOIN active_league_season
		NATURAL JOIN season
		WHERE league_code = upper($1)
		`
	err := store.
		QueryRow(q, league.Code).
		Scan(&season.ID, &season.League.Sport, &season.League.Code,
			&season.League.Name, &season.Year, &season.Type, &season.Exhibition)

	return season, err
}

func (store store) clubsByLeague(league League, active bool) ([]Club, error) {
	var q string
	if active {
		q =
			`
			SELECT club_id, club_iteration, represents, nickname
			FROM active_league_club_view
			WHERE league_code = upper($1)
			ORDER BY represents asc, nickname asc
			`
	} else {
		q =
			`
			SELECT
				DISTINCT ON(club_id) "club_id",
				club_iteration,
				represents,
				nickname
			FROM season_club
			NATURAL JOIN club
			WHERE league_code = upper($1)
			ORDER BY club_id, club_iteration desc
			`
	}

	var clubs []Club
	rows, err := store.Query(q, league.Code)
	if err != nil {
		return clubs, err
	}
	defer rows.Close()

	for rows.Next() {
		var club Club
		err = rows.Scan(&club.ID, &club.Iteration,
			&club.Represents, &club.Nickname)
		if err != nil {
			return clubs, err
		}
		clubs = append(clubs, club)
	}

	if rows.Err() != nil {
		return clubs, rows.Err()
	}
	return clubs, nil
}

func (store store) clubsBySeason(season Season) ([]Club, error) {
	query :=
		`
		SELECT club_id, club_iteration, represents, nickname
		FROM season_club_view
		WHERE season_id = $1
		`

	var clubs []Club
	rows, err := store.Query(query, season.ID)
	if err != nil {
		return clubs, err
	}
	defer rows.Close()

	for rows.Next() {
		var club Club
		err = rows.Scan(&club.ID, &club.Iteration,
			&club.Represents, &club.Nickname)
		if err != nil {
			return clubs, err
		}
		clubs = append(clubs, club)
	}

	if rows.Err() != nil {
		return clubs, rows.Err()
	}
	return clubs, nil
}

func (store store) game(id int) (Game, error) {
	q :=
		`
		SELECT
			game_id, season_id, sport_name,
			league_code, league_name, start_year,
			season_type, exhibition, game_date::text,
			coalesce(title, ''), coalesce(venue, ''),
			home_id, home_iteration,
			home_represents, home_nickname, home_score,
			away_id, away_iteration, away_represents,
			away_nickname, away_score
		FROM game_view
		WHERE game_id = $1
		`
	var game Game
	err := store.
		QueryRow(q, id).
		Scan(&game.ID, &game.Season.ID, &game.Season.League.Sport,
			&game.Season.League.Code, &game.Season.League.Name,
			&game.Season.Year, &game.Season.Type, &game.Season.Exhibition,
			&game.Date, &game.Title, &game.Venue, &game.Home.ID,
			&game.Home.Iteration, &game.Home.Represents,
			&game.Home.Nickname, &game.HomeScore, &game.Away.ID,
			&game.Away.Iteration, &game.Away.Represents,
			&game.Away.Nickname, &game.AwayScore)

	return game, err
}

func (store store) games(filter gameFilter) ([]Game, error) {
	var games []Game
	q, args := gameQuery(filter)
	rows, err := store.Query(q, args...)
	if err != nil {
		return games, err
	}
	defer rows.Close()

	for rows.Next() {
		var game Game
		err = rows.Scan(&game.ID, &game.Season.ID, &game.Season.League.Sport,
			&game.Season.League.Code, &game.Season.League.Name,
			&game.Season.Year, &game.Season.Type, &game.Season.Exhibition,
			&game.Date, &game.Title, &game.Venue, &game.Home.ID,
			&game.Home.Iteration, &game.Home.Represents,
			&game.Home.Nickname, &game.HomeScore, &game.Away.ID,
			&game.Away.Iteration, &game.Away.Represents,
			&game.Away.Nickname, &game.AwayScore)
		if err != nil {
			return games, err
		}
		games = append(games, game)
	}

	if rows.Err() != nil {
		return games, rows.Err()
	}

	return games, nil
}

func (store store) createGame(season Season, date string,
	home Club, homeScore int, away Club, awayScore int,
	title string, venue string) (Game, error) {

	var id int
	q := "SELECT new_game(upper($1), $2, $3, $4, $5, $6, $7, $8, $9)"
	err := store.
		QueryRow(q,
			season.League.Code,
			season.ID,
			home.ID,
			homeScore,
			away.ID,
			awayScore,
			date,
			title,
			venue).
		Scan(&id)

	if err != nil {
		return Game{}, err
	}

	return store.game(id)
}

func (store store) editGame(game Game, edit gameEdit) (Game, error) {
	// Need to make up to three updates - game, home game_club,
	// and away game_club.
	// Prepare update to game if necessary
	updateGame := false
	var updateGameQuery strings.Builder

	arg := 1
	gameUpdates := make([]string, 0, 3)
	gameUpdateArgs := make([]interface{}, 0, 4)

	if date, set := edit.Date(); set {
		update := fmt.Sprintf("%s = $%d", "game_date", arg)
		gameUpdates = append(gameUpdates, update)
		gameUpdateArgs = append(gameUpdateArgs, date)
		arg++
	}

	if title, set := edit.Title(); set {
		update := fmt.Sprintf("%s = $%d", "title", arg)
		gameUpdates = append(gameUpdates, update)
		gameUpdateArgs = append(gameUpdateArgs, title)
		arg++
	}

	if venue, set := edit.Venue(); set {
		update := fmt.Sprintf("%s = $%d", "venue", arg)
		gameUpdates = append(gameUpdates, update)
		gameUpdateArgs = append(gameUpdateArgs, venue)
		arg++
	}

	if arg > 1 {
		updateGame = true
		updateGameQuery.WriteString("UPDATE game SET ")
		updateGameQuery.WriteString(strings.Join(gameUpdates, ", "))
		fmt.Fprintf(&updateGameQuery, " WHERE game_id = $%d", arg)
		gameUpdateArgs = append(gameUpdateArgs, game.ID)
	}

	// Prepare update to home game_club if necessary
	updateClubHome := false
	var updateClubHomeQuery strings.Builder
	updateClubHomeArgs := make([]interface{}, 3)
	if score, set := edit.HomeScore(); set {
		updateClubHome = true
		updateClubHomeQuery.WriteString("UPDATE game_club SET ")
		updateClubHomeQuery.WriteString("score = $1")
		updateClubHomeQuery.WriteString(" WHERE game_id = $2 and club_id = $3")
		updateClubHomeArgs[0] = score
		updateClubHomeArgs[1] = game.ID
		updateClubHomeArgs[2] = game.Home.ID
	}

	// Prepare update to away game_club if necessary
	updateClubAway := false
	var updateClubAwayQuery strings.Builder
	updateClubAwayArgs := make([]interface{}, 3)
	if score, set := edit.AwayScore(); set {
		updateClubAway = true
		updateClubAwayQuery.WriteString("UPDATE game_club SET ")
		updateClubAwayQuery.WriteString("score = $1")
		updateClubAwayQuery.WriteString(" WHERE game_id = $2 and club_id = $3")
		updateClubAwayArgs[0] = score
		updateClubAwayArgs[1] = game.ID
		updateClubAwayArgs[2] = game.Away.ID
	}

	// Store the updates
	tx, err := store.Begin()
	if err != nil {
		return game, err
	}

	if updateGame {
		_, err := tx.Exec(updateGameQuery.String(), gameUpdateArgs...)
		if err != nil {
			tx.Rollback()
			return game, err
		}
	}

	if updateClubHome {
		_, err := tx.Exec(updateClubHomeQuery.String(), updateClubHomeArgs...)
		if err != nil {
			tx.Rollback()
			return game, err
		}
	}

	if updateClubAway {
		_, err := tx.Exec(updateClubAwayQuery.String(), updateClubAwayArgs...)
		if err != nil {
			tx.Rollback()
			return game, err
		}
	}

	if err := tx.Commit(); err != nil {
		return game, err
	}

	return store.game(game.ID)
}

func (store store) resources(game Game) ([]Resource, error) {
	var resources []Resource
	q :=
		`
		SELECT resource_id, title, url
		FROM resource
		WHERE game_id = $1
		`
	rows, err := store.Query(q, game.ID)
	if err != nil {
		return resources, err
	}
	defer rows.Close()

	for rows.Next() {
		var resource Resource
		err = rows.Scan(&resource.ID, &resource.Title, &resource.URL)
		if err != nil {
			return resources, err
		}
		resources = append(resources, resource)
	}

	if rows.Err() != nil {
		return resources, rows.Err()
	}

	return resources, nil
}

func (store store) createResource(game Game, title, url string) (Resource, error) {
	var r Resource
	q :=
		`
		INSERT INTO resource(game_id, title, url)
		VALUES($1, $2, $3)
		RETURNING resource_id, title, url
		`
	err := store.QueryRow(q, game.ID, title, url).Scan(&r.ID, &r.Title, &r.URL)

	if err != nil {
		return r, err
	}
	return r, nil
}

func (store store) deleteResource(resource Resource) error {
	q := fmt.Sprintf("DELETE FROM resource WHERE resource_id = $1")
	_, err := store.Exec(q, resource.ID)
	return err
}

// seasonQuery creates an SQL query string and a list of []interface{}
// arguments suitable for Query() from the given seasonFilter
func seasonQuery(sf seasonFilter) (string, []interface{}) {
	// Base season query
	var q strings.Builder
	q.WriteString(
		`
		SELECT 
			season_id, sport_name, league_code, league_name,
			start_year, season_type, exhibition
		FROM season
		NATURAL JOIN league
		NATURAL JOIN sport
		`)

	// Apply any filters from sf to query
	where := make([]string, 0, 2)
	args := make([]interface{}, 0, 2)
	arg := 1

	// Filter on league
	if league, set := sf.League(); set {
		where = append(where, fmt.Sprintf("league_code = $%d", arg))
		args = append(args, league.Code)
		arg++
	}

	// Filter on club
	if club, set := sf.Club(); set {
		where = append(where, fmt.Sprintf("season_id in (select season_id from season_club where club_id = $%d)", arg))
		args = append(args, club.ID)
		arg++
	}

	// Add any filter to query string now
	if arg > 1 {
		fmt.Fprintf(&q, " WHERE %s", strings.Join(where, " AND "))
	}

	// Seasons always ordered in chronological order.
	// Limit is set by filter
	q.WriteString(" ORDER BY start_year asc, season_id desc")
	if lim, set := sf.Limit(); set {
		fmt.Fprintf(&q, " LIMIT %d", lim)
	}

	return q.String(), args
}

// gameQuery creates an SQL query string a list of []interface{}
// suitable to pass to Query() from the given gameFilter
func gameQuery(gf gameFilter) (string, []interface{}) {
	// Start with base game query
	var q strings.Builder
	q.WriteString(
		`
		SELECT
			game_id, season_id, sport_name,
			league_code, league_name, start_year,
			season_type, exhibition, game_date::text,
			coalesce(title, ''), coalesce(venue, ''),
			home_id, home_iteration,
			home_represents, home_nickname, home_score,
			away_id, away_iteration, away_represents,
			away_nickname, away_score
		FROM game_view
		`)

	// Apply any filters from gf to query
	var where []string
	arg := 1

	// Filter by game ID
	if len(gf.IDs) > 0 {
		where = append(
			where,
			fmt.Sprintf("game_id in (%s)", ordinate(arg, len(gf.IDs))))
		arg += len(gf.IDs)
	}

	// Filter by clubs playing
	if len(gf.Clubs) > 0 {
		where = append(
			where,
			fmt.Sprintf("game_id in (select game_id from game_club where club_id in (%s))",
				ordinate(arg, len(gf.Clubs))))
		arg += len(gf.Clubs)
	}

	// Filter by season of play
	if len(gf.Seasons) > 0 {
		where = append(
			where,
			fmt.Sprintf("season_id in (%s)", ordinate(arg, len(gf.Seasons))))
		arg += len(gf.Seasons)
	}

	// Filter by league of play
	if len(gf.Leagues) > 0 {
		where = append(
			where,
			fmt.Sprintf("league_code in (%s)", ordinate(arg, len(gf.Leagues))))
		arg += len(gf.Leagues)
	}

	// If gf contained any filters, add them to the query now
	if arg > 1 {
		fmt.Fprintf(&q, " WHERE %s", strings.Join(where, " AND "))
	}

	// Games always ordered in reverse chronological order.
	// Limit set by filter
	q.WriteString(" ORDER BY game_date desc, game_id desc")
	if lim, set := gf.Limit(); set {
		fmt.Fprintf(&q, " LIMIT %d", lim)
	}

	// Create slice of all arguments passed in gf
	// in the same ordering they appear in the query
	args := make([]interface{}, 0, arg-1)
	for _, id := range gf.IDs {
		args = append(args, id)
	}
	for _, club := range gf.Clubs {
		args = append(args, club.ID)
	}
	for _, season := range gf.Seasons {
		args = append(args, season.ID)
	}
	for _, league := range gf.Leagues {
		args = append(args, strings.ToUpper(league.Code))
	}

	return q.String(), args
}

// ordinate returns a string of the form
// $start $(start + 1) ... $(start + len - 1)
func ordinate(start, len int) string {
	var b strings.Builder
	for i := start; i < start+len; i++ {
		fmt.Fprintf(&b, "$%d ", i)
	}
	return b.String()
}
