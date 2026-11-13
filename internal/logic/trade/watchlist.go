package trade

import (
	"context"
	"strings"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
)

// The Alpaca watchlists this app owns. They also show in the Alpaca dashboard.
const (
	WatchlistName = "alpaca-app"           // the watchlist tab
	FavoritesName = "alpaca-app-favorites" // the star on each symbol row
)

// watchlistID finds a watchlist by name; "" when it does not exist yet.
// ponytail: one lookup per request, no cache; cache the id if the tab feels slow.
func watchlistID(name string) (string, error) {
	lists, err := client.GetWatchlists()
	if err != nil {
		return "", err
	}
	for _, l := range lists {
		if l.Name == name {
			return l.ID, nil
		}
	}
	return "", nil
}

// Watchlist returns the symbols of one of the app's watchlists, in the order they were added.
func Watchlist(ctx context.Context, name string) ([]string, error) {
	id, err := watchlistID(name)
	if err != nil || id == "" {
		return []string{}, err
	}
	w, err := client.GetWatchlist(id) // the list endpoint does not return the assets
	if err != nil {
		return nil, err
	}
	return watchSymbols(w), nil
}

// WatchlistAdd adds a symbol and returns the new list. The first add creates the watchlist.
// Alpaca rejects an unknown symbol and a symbol that is already in the list.
func WatchlistAdd(ctx context.Context, name, symbol string) ([]string, error) {
	symbol = strings.ToUpper(symbol)
	id, err := watchlistID(name)
	if err != nil {
		return nil, err
	}
	var w *alpaca.Watchlist
	if id == "" {
		w, err = client.CreateWatchlist(alpaca.CreateWatchlistRequest{Name: name, Symbols: []string{symbol}})
	} else {
		w, err = client.AddSymbolToWatchlist(id, alpaca.AddSymbolToWatchlistRequest{Symbol: symbol})
	}
	if err != nil {
		return nil, rejected(err, "לא ניתן להוסיף את "+symbol)
	}
	return watchSymbols(w), nil
}

func WatchlistRemove(ctx context.Context, name, symbol string) error {
	symbol = strings.ToUpper(symbol)
	id, err := watchlistID(name)
	if err != nil || id == "" {
		return err
	}
	err = client.RemoveSymbolFromWatchlist(id, alpaca.RemoveSymbolFromWatchlistRequest{Symbol: symbol})
	return rejected(err, "לא ניתן להסיר את "+symbol)
}

func watchSymbols(w *alpaca.Watchlist) []string {
	out := make([]string, len(w.Assets))
	for i, a := range w.Assets {
		out[i] = a.Symbol
	}
	return out
}
