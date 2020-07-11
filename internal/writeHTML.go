package internal

import (
	"compress/gzip"
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	text "text/template"
)

type documentGenerator struct {
	staticPath      string
	templatePath    string
	indexTemplate   *template.Template
	gameTemplate    *template.Template
	sidebarTemplate *text.Template
}

func (dg *documentGenerator) Initialize(recapPath string) {
	dg.staticPath = filepath.Join(recapPath, "www")
	dg.templatePath = filepath.Join(recapPath, "templates")

	indexTemplate := filepath.Join(dg.templatePath, "index.tmpl")
	gameTemplate := filepath.Join(dg.templatePath, "game.tmpl")
	sidebarTemplate := filepath.Join(dg.templatePath, "sidebar.tmpl")
	headerTemplate := filepath.Join(dg.templatePath, "header.tmpl")

	funcMap := template.FuncMap{"GamePath": gamePath, "DateShort": dateShort, "DateLong": dateLong}
	dg.indexTemplate = template.Must(template.New("index.tmpl").Funcs(funcMap).ParseFiles(indexTemplate, headerTemplate))
	dg.gameTemplate = template.Must(template.New("game.tmpl").Funcs(funcMap).ParseFiles(gameTemplate, headerTemplate))
	dg.sidebarTemplate = text.Must(text.New("sidebar.tmpl").ParseFiles(sidebarTemplate))
}

func (dg documentGenerator) gamePath(game Game) string {
	return filepath.Join(dg.staticPath, gamePath(game))
}

func (dg documentGenerator) clubPath(club Club, season Season) string {
	return filepath.Join(dg.staticPath, clubPath(club, season))
}

func (dg documentGenerator) leaguePath(league League) string {
	return filepath.Join(dg.staticPath, leaguePath(league))
}

func (dg documentGenerator) indexPath() string {
	return filepath.Join(dg.staticPath, indexPath())
}

func (dg documentGenerator) clubSidebarPath(club Club, league League) string {
	return filepath.Join(dg.templatePath,
		fmt.Sprintf("sidebar/%s/%d.tmpl", league.Code, club.ID))
}

func (dg documentGenerator) leagueSidebarPath(league League) string {
	return filepath.Join(dg.templatePath,
		fmt.Sprintf("sidebar/%s/index.tmpl", league.Code))
}

func (dg documentGenerator) indexSidebarPath() string {
	return filepath.Join(dg.templatePath, "sidebar/index.tmpl")
}

func createFile(path string) (*os.File, error) {
	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	return os.Create(path)
}

func (dg documentGenerator) clubSidebar(club Club, seasons []Season, league League) error {
	path := dg.clubSidebarPath(club, league)
	file, err := createFile(path)
	if err != nil {
		return err
	}
	defer file.Close()

	doc := document{dg, file}
	return doc.clubSidebar(club, seasons)
}

func (dg documentGenerator) leagueSidebar(league League, season Season, clubs []Club) error {
	path := dg.leagueSidebarPath(league)
	file, err := createFile(path)
	if err != nil {
		return err
	}
	defer file.Close()

	doc := document{dg, file}
	return doc.leagueSidebar(season, clubs)
}

func (dg documentGenerator) indexSidebar(leagues []League) error {
	path := dg.indexSidebarPath()
	file, err := createFile(path)
	if err != nil {
		return err
	}
	defer file.Close()

	doc := document{dg, file}
	return doc.indexSidebar(leagues)
}

func (dg documentGenerator) gamePage(game Game, resources []Resource) error {
	var (
		file   *os.File
		fileGZ *os.File
		err    error
	)

	path := dg.gamePath(game)
	if file, err = createFile(path); err != nil {
		return err
	}
	defer file.Close()

	pathGZ := fmt.Sprintf("%s.gz", path)
	if fileGZ, err = createFile(pathGZ); err != nil {
		return err
	}
	defer fileGZ.Close()
	gz, _ := gzip.NewWriterLevel(fileGZ, gzip.BestCompression)
	defer gz.Close()

	doc := document{dg, io.MultiWriter(file, gz)}
	return doc.game(game, resources)
}

func (dg documentGenerator) clubIndex(club Club, season Season, games []Game) error {
	var (
		file   *os.File
		fileGZ *os.File
		err    error
	)

	path := dg.clubPath(club, season)
	if file, err = createFile(path); err != nil {
		return err
	}
	defer file.Close()

	pathGZ := fmt.Sprintf("%s.gz", path)
	if fileGZ, err = createFile(pathGZ); err != nil {
		return err
	}
	defer fileGZ.Close()
	gz, _ := gzip.NewWriterLevel(fileGZ, gzip.BestCompression)
	defer gz.Close()

	doc := document{dg, io.MultiWriter(file, gz)}
	return doc.club(club, season, games)
}

func (dg documentGenerator) leagueIndex(league League, games []Game) error {
	var (
		file   *os.File
		fileGZ *os.File
		err    error
	)

	path := dg.leaguePath(league)
	if file, err = createFile(path); err != nil {
		return err
	}
	defer file.Close()

	pathGZ := fmt.Sprintf("%s.gz", path)
	if fileGZ, err = createFile(pathGZ); err != nil {
		return err
	}
	defer fileGZ.Close()
	gz, _ := gzip.NewWriterLevel(fileGZ, gzip.BestCompression)
	defer gz.Close()

	doc := document{dg, io.MultiWriter(file, gz)}
	return doc.league(league, games)
}

func (dg documentGenerator) index(games []Game) error {
	var (
		file   *os.File
		fileGZ *os.File
		err    error
	)

	path := dg.indexPath()
	if file, err = createFile(path); err != nil {
		return err
	}
	defer file.Close()

	pathGZ := fmt.Sprintf("%s.gz", path)
	if fileGZ, err = createFile(pathGZ); err != nil {
		return err
	}
	defer fileGZ.Close()
	gz, _ := gzip.NewWriterLevel(fileGZ, gzip.BestCompression)
	defer gz.Close()

	doc := document{dg, io.MultiWriter(file, gz)}
	return doc.index(games)
}
