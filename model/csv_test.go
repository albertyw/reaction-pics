package model

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadPostsFromCSV(t *testing.T) {
	posts := ReadPostsFromCSV(getCSV(true))
	assert.Equal(t, len(posts), 1)
	assert.Equal(t, posts[0].ID, int64(1234))
	assert.Equal(t, posts[0].Title, "title")
	assert.Equal(t, posts[0].URL, "url")
	assert.Equal(t, posts[0].Image, "https://img.reaction.pics/file/reaction-pics/abcd.gif")
	assert.Equal(t, posts[0].Likes, int64(123))
}

func TestReadPercentFromCSV(t *testing.T) {
	data := []byte(`1234,a% b,url,image,123`)
	posts := ReadPostsFromCSV(data)
	assert.Equal(t, posts[0].Title, "a% b")
}

func TestReadPostsFromCSVQuotedTitle(t *testing.T) {
	// Titles containing double quotes must be CSV-escaped with "" and work correctly.
	data := []byte(`1234,"title with a ""quoted"" word",url,image,5`)
	posts := ReadPostsFromCSV(data)
	assert.Equal(t, 1, len(posts))
	assert.Equal(t, `title with a "quoted" word`, posts[0].Title)
}

func TestProdCSVCompleteness(t *testing.T) {
	// Every newline in the CSV corresponds to one record (no multi-line fields).
	// This catches any truncation caused by malformed CSV quoting.
	expected := bytes.Count(prodCSV, []byte("\n"))
	posts := ReadPostsFromCSV(prodCSV)
	assert.Equal(t, expected, len(posts))
}
