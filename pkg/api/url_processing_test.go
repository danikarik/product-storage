package api

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFetchData(t *testing.T) {
	testCases := []struct {
		Name string
		URL  string
	}{
		{
			Name: "Dummy",
			URL:  "https://csv-samples.s3.amazonaws.com/dummy.csv",
		},
		{
			Name: "Phones",
			URL:  "https://csv-samples.s3.amazonaws.com/iphones.csv",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			r := require.New(t)
			s := &server{
				timeout: _defaultTimeout,
				hclient: &http.Client{},
			}

			err := s.fetchData(context.Background(), tc.URL)
			r.NoError(err)
		})
	}
}

func TestReadCSV(t *testing.T) {
	testCases := []struct {
		Name string
		Path string
	}{
		{
			Name: "Dummy",
			Path: "testdata/dummy.csv",
		},
		{
			Name: "Phones",
			Path: "testdata/iphones.csv",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			r := require.New(t)
			s := &server{}

			file, err := os.Open(tc.Path)
			r.NoError(err)
			defer file.Close()

			err = s.readCSV(file)
			r.NoError(err)
		})
	}
}
