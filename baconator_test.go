package baconator

import (
	"encoding/gob"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBaconator_Links(t *testing.T) {
	b := newTestBaconator(t)
	got, err := b.links("James Dean", "Ruth Buzzi")
	require.NoError(t, err)
	require.Greater(t, len(got), 0)
}

func TestCenter(t *testing.T) {
	baconator := newTestBaconator(t)
	kevin := baconator.CastNodes["Kevin Bacon"]
	res := baconator.center(kevin)
	require.NotNil(t, res)
	require.Greater(t, res.AvgDistance, 2.0)
}

func fileExists(t *testing.T, filename string) bool {
	t.Helper()
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	require.NoError(t, err)
	return true
}

func downloadTestData(t *testing.T) {
	t.Helper()
	u := "https://oracleofbacon.org/data.txt.bz2"
	res, err := http.Get(u)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode)
	file, err := os.Create("tmp/data.txt.bz2")
	require.NoError(t, err)
	_, err = io.Copy(file, res.Body)
	require.NoError(t, err)
	require.NoError(t, res.Body.Close())
	require.NoError(t, file.Close())
}

func newTestBaconator(t *testing.T) *Baconator {
	t.Helper()
	err := os.MkdirAll("tmp", 0o700)
	require.NoError(t, err)
	gobFilename := filepath.FromSlash("tmp/baconator.gob")
	if fileExists(t, gobFilename) {
		var file *os.File
		file, err = os.Open(gobFilename)
		require.NoError(t, err)
		var baconator Baconator
		err = gob.NewDecoder(file).Decode(&baconator)
		require.NoErrorf(t, err, "error loading %q. try deleting it an allowing it to be rebuilt", gobFilename)
		return &baconator
	}
	dataFilename := filepath.FromSlash("tmp/data.txt.bz2")
	if !fileExists(t, dataFilename) {
		downloadTestData(t)
	}
	movies, err := loadMovies(dataFilename)
	require.NoError(t, err)
	baconator := buildBaconator(movies)
	file, err := os.Create(gobFilename)
	require.NoError(t, err)
	err = gob.NewEncoder(file).Encode(baconator)
	require.NoError(t, err)
	require.NoError(t, file.Close())
	return newTestBaconator(t)
}
