package auctionrunner

import (
	"errors"
	"time"

	"github.com/cloudfoundry-incubator/auction/auctiontypes"
)

var AllBiddersFull = errors.New("all the bidders were full")

var DefaultStartAuctionRules = auctiontypes.StartAuctionRules{
	Algorithm:              "reserve_n_best",
	MaxRounds:              40,
	MaxBiddingPoolFraction: 0.2,
	MinBiddingPool:         10,
}

type auctionRunner struct {
	client auctiontypes.RepPoolClient
}

func New(client auctiontypes.RepPoolClient) *auctionRunner {
	return &auctionRunner{
		client: client,
	}
}

func (a *auctionRunner) RunLRPStartAuction(auctionRequest auctiontypes.StartAuctionRequest) auctiontypes.StartAuctionResult {
	result := auctiontypes.StartAuctionResult{
		LRPStartAuction: auctionRequest.LRPStartAuction,
	}

	t := time.Now()
	switch auctionRequest.Rules.Algorithm {
	case "all_rebid":
		result.Winner, result.NumRounds, result.NumCommunications = allRebidAuction(a.client, auctionRequest)
	case "all_reserve":
		result.Winner, result.NumRounds, result.NumCommunications = allReserveAuction(a.client, auctionRequest)
	case "pick_among_best":
		result.Winner, result.NumRounds, result.NumCommunications = pickAmongBestAuction(a.client, auctionRequest)
	case "pick_best":
		result.Winner, result.NumRounds, result.NumCommunications = pickBestAuction(a.client, auctionRequest)
	case "reserve_n_best":
		result.Winner, result.NumRounds, result.NumCommunications = reserveNBestAuction(a.client, auctionRequest)
	case "random":
		result.Winner, result.NumRounds, result.NumCommunications = randomAuction(a.client, auctionRequest)
	default:
		panic("unkown algorithm " + auctionRequest.Rules.Algorithm)
	}
	result.BiddingDuration = time.Since(t)

	if result.Winner == "" {
		result.Error = auctiontypes.InsufficientResources
	}

	return result
}

func (a *auctionRunner) RunLRPStopAuction(auctionRequest auctiontypes.StopAuctionRequest) auctiontypes.StopAuctionResult {
	result := auctiontypes.StopAuctionResult{
		LRPStopAuction: auctionRequest.LRPStopAuction,
	}

	t := time.Now()
	result.Winner, result.NumCommunications, result.Error = stopAuction(a.client, auctionRequest)
	result.BiddingDuration = time.Since(t)

	return result
}
