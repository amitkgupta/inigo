package auctionrunner

import "github.com/cloudfoundry-incubator/auction/auctiontypes"

/*

Get the bids from the subset of reps
	Select the best 5
		Pick a winner randomly from that set

*/

func pickAmongBestAuction(client auctiontypes.RepPoolClient, auctionRequest auctiontypes.StartAuctionRequest) (string, int, int) {
	rounds, numCommunications := 1, 0
	auctionInfo := auctiontypes.NewStartAuctionInfoFromLRPStartAuction(auctionRequest.LRPStartAuction)

	for ; rounds <= auctionRequest.Rules.MaxRounds; rounds++ {
		//pick a subset
		firstRoundReps := auctionRequest.RepGuids.RandomSubsetByFraction(auctionRequest.Rules.MaxBiddingPoolFraction, auctionRequest.Rules.MinBiddingPool)

		//get everyone's bid, if they're all full: bail
		numCommunications += len(firstRoundReps)
		firstRoundScores := client.BidForStartAuction(firstRoundReps, auctionInfo)
		if firstRoundScores.AllFailed() {
			continue
		}

		winners := firstRoundScores.FilterErrors().Shuffle().Sort()
		max := 5
		if len(winners) < max {
			max = len(winners)
		}
		top5Winners := winners[:max]

		winner := top5Winners.Shuffle()[0]

		result := client.RebidThenTentativelyReserve([]string{winner.Rep}, auctionInfo)[0]
		numCommunications += 1
		if result.Error != "" {
			continue
		}

		client.Run(winner.Rep, auctionRequest.LRPStartAuction)
		numCommunications += 1

		return winner.Rep, rounds, numCommunications
	}

	return "", rounds, numCommunications
}
