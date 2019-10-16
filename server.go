package main

import (
	//	"github.com/frncscsrcc/longpoll"
	"context"
	"github.com/briscola-as-a-service/game"
	"github.com/briscola-as-a-service/game/card"
	"github.com/briscola-as-a-service/game/hand"
	"github.com/briscola-as-a-service/game/player"
	"github.com/frncscsrcc/longpoll"
	"github.com/frncscsrcc/resthelper"
	"github.com/frncscsrcc/waitinglist"
	"log"
	"net/http"
)

// Global variables
var lp *longpoll.LongPoll
var wls *waitinglist.WaitingLists

var playerIDToDecker map[string]*game.Decker

const maxGamesPerPlayer = 2

func getWaitingList(r *http.Request) string {
	// Search in URL
	waitingLists, ok := r.URL.Query()["type"]
	if ok == true && len(waitingLists) > 0 {
		return waitingLists[0]
	}
	return ""
}

func play(decker *game.Decker) {

}

func startGame(w http.ResponseWriter, r *http.Request) {
	// Check user session (it could be moved in middleware)
	sessionID := resthelper.GetSessionID(r)
	if sessionID == "" {
		resthelper.SendError(w, 400, "Missing SessionID")
		return
	}

	waitingListName := getWaitingList(r)
	if waitingListName == "" {
		resthelper.SendError(w, 400, "Missing game type")
		return
	}

	// Note feed = subscriptionID
	// PlayerID = subscriptionID
	subscriptionID := resthelper.GetNewToken(32)
	playerID := subscriptionID

	// Add the player to a waiting list
	err := wls.AddPlayer(waitingListName, playerID, playerID)
	if err != nil {
		resthelper.SendError(w, 400, err.Error())
		return
	}

	// Inject a new subscriptionID in the request contex. This will be passed to the longpool
	// library, and the library will use this value (instead of creating a random one)
	// because there is only a feed, subscriptionID = feed
	feeds := []string{subscriptionID}
	contextStruct := longpoll.ContextStruct{
		Feeds:          feeds,
		SubscriptionID: subscriptionID,
		SessionID:      sessionID,
	}
	ctx := context.WithValue(r.Context(), "contextStruct", contextStruct)

	// Add the feed in longpoll object
	lp.AddFeed(feeds)

	deckerPtr, err := wls.StartGame(waitingListName)

	if err != nil {
		if err.Error() != "Waiting for players" {
			resthelper.SendError(w, 500, err.Error())
			return
		} else {
			lp.SubscribeHandler(w, r.WithContext(ctx))
			return
		}
	}

	players := deckerPtr.GetSortedPlayers()

	for _, player := range players {
		playerID := player.ID()
		playerIDToDecker[playerID] = deckerPtr
	}

	lp.SubscribeHandler(w, r.WithContext(ctx))

	// Send the first cars
	type PlayEvent struct {
		Cards           player.Cards
		FirstPlayerCard card.Card
		Briscola        card.Card
		CurrentHands    []hand.Hand
	}

	firstPlayerFeed := players[0].ID()

	playEvent := PlayEvent{
		Cards:    *(deckerPtr.GetPlayerCards(players[0])),
		Briscola: deckerPtr.GetBriscola(),
	}

	lp.NewEvent(firstPlayerFeed, playEvent)

	return
}

func playGame(w http.ResponseWriter, r *http.Request) {
	// Check user session (it could be moved in middleware)
	sessionID := resthelper.GetSessionID(r)
	if sessionID == "" {
		resthelper.SendError(w, 400, "Missing sessionID")
		return
	}

	//This should be moved from here!
	subscriptionIDs, ok := r.URL.Query()["subscriptionID"]
	if ok == false || len(subscriptionIDs) == 0 {
		resthelper.SendError(w, 400, "Missing subscriptionID")
		return
	}

	// TODO

	lp.ListenHandler(w, r)
}

func main() {
	wls = waitinglist.New()
	var err error
	err = wls.AddList("TEST", 2)
	if err != nil {
		log.Fatal(err)
	}

	lp = longpoll.New()
	mux := http.NewServeMux()
	mux.HandleFunc("/start", startGame)
	mux.HandleFunc("/play", playGame)
	log.Println("Start server on port :8082")
	contextedMux := resthelper.AddSessionID(resthelper.LogRequest(mux))
	log.Fatal(http.ListenAndServe(":8082", contextedMux))
}

func show(i interface{}) {
	log.Printf("%+v\n", i)
}
