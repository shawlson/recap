package internal

import (
	"fmt"
	"html/template"
	"io"
	"sort"
	"strings"
	"time"
)

type document struct {
	documentGenerator
	io.Writer
}

type link struct {
	HREF    string
	Display string
}

type breadcrumb struct {
	PathToRoot string
	League     link
	Home       link
	Away       link
}

type indexPage struct {
	Breadcrumb breadcrumb
	Title      string
	Subtitle   string
	Games      []Game
}

type navSection struct {
	Header string
	Links  []link
}

type sidebar struct {
	Breadcrumb breadcrumb
	Sections   []navSection
}

func dateShort(date string) string {
	t, _ := time.Parse("2006-01-02", date)
	return t.Format("Jan 02 2006")
}

func dateLong(date string) string {
	t, _ := time.Parse("2006-01-02", date)
	return t.Format("Mon January 02, 2006")
}

func gamePath(game Game) string {
	return fmt.Sprintf("/%s/%d/%s/games/%d.html",
		game.Season.League.Code,
		game.Season.Year,
		game.Season.Type,
		game.ID)
}

func clubPath(club Club, season Season) string {
	return fmt.Sprintf("/%s/%d/%s/teams/%d.html",
		season.League.Code,
		season.Year,
		season.Type,
		club.ID)
}

func leaguePath(league League) string {
	return fmt.Sprintf("/%s/index.html", league.Code)
}

func indexPath() string {
	return "/index.html"
}

func (doc document) game(game Game, resources []Resource) error {

	crumbs := breadcrumb{}
	crumbs.PathToRoot = "../../../.."
	crumbs.League = link{leaguePath(game.Season.League), game.Season.League.Name}
	crumbs.Home = link{clubPath(game.Home, game.Season), fmt.Sprintf("%s %s", game.Home.Represents, game.Home.Nickname)}
	crumbs.Away = link{clubPath(game.Away, game.Season), fmt.Sprintf("%s %s", game.Away.Represents, game.Away.Nickname)}

	data := struct {
		Breadcrumb breadcrumb
		Game       Game
		Resources  []Resource
	}{crumbs, game, resources}

	return doc.gameTemplate.Execute(doc, data)
}

func (doc document) club(club Club, season Season, games []Game) error {

	crumbs := breadcrumb{}
	crumbs.PathToRoot = "../../../.."
	crumbs.League = link{leaguePath(season.League), season.League.Name}

	var page indexPage
	page.Breadcrumb = crumbs
	page.Title = fmt.Sprintf("%s %s", club.Represents, club.Nickname)
	page.Subtitle = fmt.Sprintf("%d %s %s", season.Year, season.League.Name, season.Type)
	page.Games = games

	index := template.Must(
		template.Must(doc.indexTemplate.Clone()).
			ParseFiles(doc.clubSidebarPath(club, season.League)))

	return index.Execute(doc, page)
}

func (doc document) league(league League, games []Game) error {

	crumbs := breadcrumb{}
	crumbs.PathToRoot = ".."

	var page indexPage
	page.Breadcrumb = crumbs
	page.Title = fmt.Sprintf("%s %s", league.Name, league.Sport)
	page.Games = games

	index := template.Must(
		template.Must(doc.indexTemplate.Clone()).
			ParseFiles(doc.leagueSidebarPath(league)))

	return index.Execute(doc, page)
}

func (doc document) index(games []Game) error {

	crumbs := breadcrumb{}
	crumbs.PathToRoot = "."

	var page indexPage
	page.Breadcrumb = crumbs
	page.Title = "Recent Games"
	page.Games = games

	index := template.Must(
		template.Must(doc.indexTemplate.Clone()).
			ParseFiles(doc.indexSidebarPath()))

	return index.Execute(doc, page)
}

func (doc document) clubSidebar(club Club, seasons []Season) error {

	crumbs := breadcrumb{}
	crumbs.PathToRoot = "../../../.."

	// Separate out normal seasons from exhibition seasons
	regular := make([]Season, 0, len(seasons))
	exhibition := make([]Season, 0, len(seasons))
	for _, season := range seasons {
		if season.Exhibition {
			exhibition = append(exhibition, season)
		} else {
			regular = append(regular, season)
		}
	}

	// Create a navigation sections for both the normal seasons
	// and any exhibition seasons
	data := make([]navSection, 0, 2)
	if len(regular) != 0 {
		sort.Slice(regular, func(i, j int) bool {
			return regular[j].Year < regular[i].Year
		})

		section := navSection{Header: "Other Seasons"}
		section.Links = make([]link, len(regular))
		for i, season := range regular {
			section.Links[i] = link{
				clubPath(club, season),
				fmt.Sprintf("%d %s Season", season.Year, season.League.Name),
			}
		}

		data = append(data, section)
	}

	if len(exhibition) != 0 {
		sort.Slice(exhibition, func(i, j int) bool {
			return exhibition[j].Year < exhibition[i].Year
		})

		section := navSection{Header: "See also"}
		section.Links = make([]link, len(exhibition))
		for i, season := range exhibition {
			section.Links[i] = link{
				clubPath(club, season),
				fmt.Sprintf("%d %s %s", season.Year, season.League.Name, season.Type),
			}
		}

		data = append(data, section)
	}

	return doc.sidebarTemplate.Execute(doc, sidebar{Breadcrumb: crumbs, Sections: data})
}

func (doc document) leagueSidebar(season Season, clubs []Club) error {

	crumbs := breadcrumb{}
	crumbs.PathToRoot = ".."

	var section navSection
	section.Header = "Teams"
	section.Links = make([]link, len(clubs))

	sort.Slice(clubs, func(i, j int) bool {
		return strings.ToLower(clubs[i].Represents) < strings.ToLower(clubs[j].Represents)
	})

	for i, club := range clubs {
		section.Links[i] = link{
			clubPath(club, season),
			fmt.Sprintf("%s %s", club.Represents, club.Nickname),
		}
	}

	return doc.sidebarTemplate.Execute(doc, sidebar{Breadcrumb: crumbs, Sections: []navSection{section}})
}

func (doc document) indexSidebar(leagues []League) error {

	crumbs := breadcrumb{}
	crumbs.PathToRoot = "."

	var section navSection
	section.Header = "Leagues"
	section.Links = make([]link, len(leagues))

	sort.Slice(leagues, func(i, j int) bool {
		return strings.ToLower(leagues[i].Name) < strings.ToLower(leagues[j].Name)
	})

	for i, league := range leagues {
		section.Links[i] = link{
			leaguePath(league),
			fmt.Sprintf("%s %s", league.Name, league.Sport),
		}
	}

	return doc.sidebarTemplate.Execute(doc, sidebar{Breadcrumb: crumbs, Sections: []navSection{section}})
}
