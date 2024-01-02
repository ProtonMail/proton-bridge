// Copyright (c) 2024 Proton AG
//
// This file is part of Proton Mail Bridge.
//
// Proton Mail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Proton Mail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/ProtonMail/proton-bridge/v3/internal/kb"
	"github.com/urfave/cli/v2"
)

const flagArticles = "articles"
const flagInput = "input"

func main() {
	app := &cli.App{
		Name:            "kb-suggester",
		Usage:           "test bridge KB article suggester",
		HideHelpCommand: true,
		ArgsUsage:       "",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:      flagArticles,
				Aliases:   []string{"a"},
				Usage:     "use `articles.json` as the JSON article list",
				TakesFile: true,
			},
			&cli.StringFlag{
				Name:      flagInput,
				Aliases:   []string{"i"},
				Usage:     "read user input from the `userInput` file",
				TakesFile: true,
			},
		},
		Action: run,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func getUserInput(ctx *cli.Context) (string, error) {
	inputFile := ctx.String(flagInput)
	var bytes []byte
	var err error

	if len(inputFile) == 0 {
		var fi os.FileInfo
		if fi, err = os.Stdin.Stat(); err != nil {
			return "", err
		}

		if (fi.Mode() & os.ModeNamedPipe) == 0 {
			fmt.Println("Type your input, Ctrl+D to finish: ")
		}
		bytes, err = io.ReadAll(os.Stdin)
	} else {
		bytes, err = os.ReadFile(filepath.Clean(inputFile))
	}

	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

func getArticleList(ctx *cli.Context) (kb.ArticleList, error) {
	articleFile := ctx.String(flagArticles)
	if len(articleFile) == 0 {
		return kb.GetArticleList()
	}

	bytes, err := os.ReadFile(filepath.Clean(articleFile))
	if err != nil {
		return nil, err
	}

	var result kb.ArticleList
	err = json.Unmarshal(bytes, &result)
	return result, err
}

func run(ctx *cli.Context) error {
	if ctx.Args().Len() > 0 {
		_ = cli.ShowAppHelp(ctx)
		return errors.New("command accept no argument")
	}

	articles, err := getArticleList(ctx)
	if err != nil {
		return err
	}

	userInput, err := getUserInput(ctx)
	if err != nil {
		return err
	}

	suggestions, err := kb.GetSuggestionsFromArticleList(userInput, articles)
	if err != nil {
		return err
	}

	if len(suggestions) == 0 {
		fmt.Println("No suggestions found")
		return nil
	}

	for _, suggestion := range suggestions {
		fmt.Printf("Score %v: %v (%v)\n", suggestion.Score, suggestion.Title, suggestion.URL)
	}

	return nil
}
