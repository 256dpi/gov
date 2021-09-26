package main

import (
	"net/http"

	"github.com/google/pprof/profile"
)

func getProfile(url string) (*profile.Profile, error) {
	// get profile
	res, err := http.Get(url + "?seconds=1")
	if err != nil {
		return nil, err
	}

	// ensure close
	defer res.Body.Close()

	// parse profile
	prf, err := profile.Parse(res.Body)
	if err != nil {
		return nil, err
	}

	return prf, nil
}
