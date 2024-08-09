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

package kb

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_ArticleList(t *testing.T) {
	articles, err := GetArticleList()
	require.NoError(t, err)
	require.NotEmpty(t, articles)
	var bits uint64
	for _, article := range articles {
		require.Truef(t, article.Index < 64, "Invalid KB article index %d, (must be < 64)", article.Index)
		require.Zerof(t, bits&(1<<article.Index), "Duplicate index %d in knowledge base", article.Index)
		bits |= bits | (1 << article.Index)
		require.NotEmpty(t, article.URL, "KB article with index %d has no URL", article.Index)
		require.NotEmpty(t, article.Title, "KB article with index %d has no title", article.Index)
		require.NotEmpty(t, article.Keywords, "KB article with index %d has no keyword", article.Index)
	}
}

func Test_GetSuggestions(t *testing.T) {
	suggestions, err := GetSuggestions("Thunderbird is not working, error during password")
	require.NoError(t, err)
	count := len(suggestions)
	require.True(t, (count > 0) && (count <= 5))
	suggestions, err = GetSuggestions("Supercalifragilisticexpialidocious Sesquipedalian Worcestershire")
	require.NoError(t, err)
	require.Empty(t, suggestions)
}

func Test_GetSuggestionsFromArticleList(t *testing.T) {
	articleList := ArticleList{}
	suggestions, err := GetSuggestionsFromArticleList("Thunderbird", articleList)
	require.NoError(t, err)
	require.Empty(t, suggestions)

	articleList = ArticleList{
		&Article{
			Index:    0,
			URL:      "https://proton.me",
			Title:    "Proton home page",
			Keywords: []string{"proton"},
		},
		&Article{
			Index:    1,
			URL:      "https://mozilla.org",
			Title:    "Mozilla home page",
			Keywords: []string{"mozilla"},
		},
	}
	suggestions, err = GetSuggestionsFromArticleList("PRoToN", articleList)
	require.NoError(t, err)
	require.Len(t, suggestions, 1)
	require.Equal(t, suggestions[0].URL, "https://proton.me")
}

func Test_GetArticleIndex(t *testing.T) {
	index1, err := GetArticleIndex("https://proton.me/support/apple-mail-certificate")
	require.NoError(t, err)

	index2, err := GetArticleIndex("HTTPS://PROTON.ME/support/apple-mail-certificate")
	require.NoError(t, err)
	require.Equal(t, index1, index2)

	_, err = GetArticleIndex("https://proton.me")
	require.ErrorIs(t, err, ErrArticleNotFound)
}

func Test_simplifyUserInput(t *testing.T) {
	require.Equal(t, "word1 ñóÄ don't déjà 33 pizza", simplifyUserInput("  \nword1    \n\tñóÄ   don't\n\n\ndéjà, 33    pizza=🍕\n,\n"))
}
