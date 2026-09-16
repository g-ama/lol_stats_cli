package api

import (
	"encoding/json"
	"fmt"
	"io"
	"lol_stats/internal/model"
	"net/http"
	"strings"
)

const apiBasePuuidURL = "https://americas.api.riotgames.com/riot/account/v1/accounts/by-riot-id/"
const apiBaseMatchesURL = "https://americas.api.riotgames.com/lol/match/v5/matches/by-puuid/"
const apiBaseMatchURL = "https://americas.api.riotgames.com/lol/match/v5/matches/"
const apiPrefix = "?api_key="

const url = "https://americas.api.riotgames.com/riot/account/v1/accounts/by-riot-id/FREE%20PALESTINE/tox?api_key=RGAPI-a66242b8-2b2a-4911-aadf-0ecd983bea2d"

func QueryAccount(username string, tagline string, apiKey string) (model.Account, error) {

	username = strings.ReplaceAll(username, " ", "%20") + "/"

	query := apiBasePuuidURL + username + tagline + apiPrefix + apiKey

	response, err := http.Get(query)

	if err != nil {
		return model.Account{}, fmt.Errorf("unable to fetch Account: %w", err)
	}

	if response.StatusCode != http.StatusOK {
		return model.Account{}, fmt.Errorf("bad request: %d", response.StatusCode)
	}

	responseBody, err := io.ReadAll(response.Body)

	if err != nil {
		return model.Account{}, fmt.Errorf("unable to read response body %v", err)
	}

	account := model.Account{}

	if err := json.Unmarshal(responseBody, &account); err != nil {
		return model.Account{}, fmt.Errorf("unable to to unmarshal response body")
	}

	return account, nil
}

func QueryGoMatch(matchID string, apiKey string, channel chan model.Match) {

	response, err := http.Get(apiBaseMatchURL + matchID + apiPrefix + apiKey)

	if response.StatusCode != http.StatusOK || err != nil {
		fmt.Printf("bad request for match query: %d", response.StatusCode)
		panic("status code issues")
	}

	responseBody, err := io.ReadAll(response.Body)

	if err != nil {
		panic("io issues")
	}

	match := model.Match{}

	if err := json.Unmarshal(responseBody, &match); err != nil {
		panic("marshaling issue")
	}

	channel <- match

}

func QueryMatch(matchID string, apiKey string) (model.Match, error) {

	response, err := http.Get(apiBaseMatchURL + matchID + apiPrefix + apiKey)

	if response.StatusCode != http.StatusOK {
		return model.Match{}, fmt.Errorf("bad request for match query: %d", response.StatusCode)
	}

	if err != nil {
		return model.Match{}, fmt.Errorf("error")
	}

	responseBody, err := io.ReadAll(response.Body)

	if err != nil {
		return model.Match{}, fmt.Errorf("error")
	}

	match := model.Match{}

	if err := json.Unmarshal(responseBody, &match); err != nil {
		return model.Match{}, fmt.Errorf("error")
	}

	return match, nil

}

// TODO: Implement a go routine to make the querying faster

func QueryMatches(puuid string, apiKey string) ([]model.Match, error) {

	query := apiBaseMatchesURL + puuid + "/" + "ids?queue=420&type=ranked&start=0&count=50&api_key=" + apiKey

	response, err := http.Get(query)

	if response.StatusCode != http.StatusOK {
		return []model.Match{}, fmt.Errorf("bad request: %d", response.StatusCode)
	}

	if err != nil {
		return []model.Match{}, fmt.Errorf("unable to query matches: %w", err)
	}

	responseBody, err := io.ReadAll(response.Body)

	if err != nil {
		return []model.Match{}, fmt.Errorf("unable to read response body: %w", err)
	}

	var matchIDs []string

	if err := json.Unmarshal(responseBody, &matchIDs); err != nil {
		return []model.Match{}, fmt.Errorf("unable to unmarshal matchIDs: %w", err)
	}

	matches := []model.Match{}

	channel := make(chan model.Match)

	for _, i := range matchIDs {
		go QueryGoMatch(i, apiKey, channel)

	}

	for range matchIDs {
		match := <-channel
		matches = append(matches, match)
	}

	return matches, nil

}
