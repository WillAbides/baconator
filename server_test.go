package baconator

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestServer_Center(t *testing.T) {
	baconator := newTestBaconator(t)
	server := httptest.NewServer(NewServer(baconator))
	u := server.URL + "/center?p=" + url.QueryEscape("Kevin Bacon")
	got, err := http.Get(u)
	require.NoError(t, err)
	_, err = ioutil.ReadAll(got.Body)
	require.NoError(t, err)
}

func TestServer_link(t *testing.T) {
	baconator := newTestBaconator(t)
	server := httptest.NewServer(NewServer(baconator))
	u := fmt.Sprintf("%s/link?a=%s&b=%s", server.URL,
		url.QueryEscape("Wonsanman"),
		url.QueryEscape("Kevin Bacon"),
	)
	_, err := http.Get(u)
	require.NoError(t, err)
}
