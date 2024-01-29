package handlers

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strconv"
	"sync"

	bd "github.com/echenim/barter/internal/bidder"
	md "github.com/echenim/barter/internal/models"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

const (
	MarketETH md.Market = "ETH"

	MarketOrder md.OrderType = "MARKET"
	LimitOrder  md.OrderType = "LIMIT"
)

type Exchange struct {
	Client *ethclient.Client
	mu     sync.RWMutex
	Users  map[int64]*md.User
	// bid maps a user to his bids.
	bids     map[int64][]*bd.Bid
	PrivateKey *ecdsa.PrivateKey
	bidBk map[md.Market]*bd.BookBid
}

// NewExchange creates a new Exchange instance with a specified private key and Ethereum client.
// It initializes the bid book for the ETH market.
func NewExchange(privateKey string, client *ethclient.Client) (*Exchange, error) {
	bidbooks := make(map[md.Market]*bd.BookBid)
	bidbooks[MarketETH] = bd.NewBookBid()

	pk, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		return nil, err
	}

	return &Exchange{
		Client:     client,
		Users:      make(map[int64]*md.User),
		bids:     make(map[int64][]*bd.Bid),
		PrivateKey: pk,
		bidBk: bidbooks,
	}, nil
}

type GetOrdersResponse struct {
	Asks []md.Order
	Bids []md.Order
}

// RegisterUser adds a new user to the exchange with a given private key and user ID.
func (ex *Exchange) RegisterUser(pk string, userId int64) {
	user := md.NewUser(pk, userId)
	ex.Users[userId] = user

	logrus.WithFields(logrus.Fields{
		"id": userId,
	}).Info("new exchange user")
}

// GetTrades returns the trades for a specified market from the exchange.
func (ex *Exchange) GetTrades(c echo.Context) error {
	market := md.Market(c.Param("market"))
	ob, ok := ex.bidBk[market]
	if !ok {
		return c.JSON(http.StatusBadRequest, md.APIError{Error: "bidBK not found"})
	}

	return c.JSON(http.StatusOK, ob.Trades)
}

// GetOrders retrieves the current orders (bids and asks) for a specific user.
func (ex *Exchange) GetOrders(c echo.Context) error {
	userIDStr := c.Param("userID")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		return err
	}

	ex.mu.RLock()
	orderbookOrders := ex.bids[int64(userID)]
	ordersResp := &GetOrdersResponse{
		Asks: []md.Order{},
		Bids: []md.Order{},
	}

	for i := 0; i < len(orderbookOrders); i++ {
		// It could be that the bid is getting filled even though its included in this
		// response. We must double check if the limit is not nil
		if orderbookOrders[i].Limit == nil {
			continue
		}

		order := md.Order{
			ID:        orderbookOrders[i].ID,
			UserID:    orderbookOrders[i].UserID,
			Price:     orderbookOrders[i].Limit.Price,
			Size:      orderbookOrders[i].Size,
			Timestamp: orderbookOrders[i].Timestamp,
			Bid:       orderbookOrders[i].Bid,
		}

		if order.Bid {
			ordersResp.Bids = append(ordersResp.Bids, order)
		} else {
			ordersResp.Asks = append(ordersResp.Asks, order)
		}
	}
	ex.mu.RUnlock()

	return c.JSON(http.StatusOK, ordersResp)
}

// GetBook provides the order book for a specified market, including total bid and ask volumes.
func (ex *Exchange) GetBook(c echo.Context) error {
	market := md.Market(c.Param("market"))
	ob, ok := ex.bidBk[market]
	if !ok {
		return c.JSON(http.StatusBadRequest, map[string]any{"msg": "market not found"})
	}

	orderbookData := md.BookedBidData{
		TotalBidVolume: ob.BidTotalVolume(),
		TotalAskVolume: ob.AskTotalVolume(),
		Asks:           []*md.Order{},
		Bids:           []*md.Order{},
	}

	for _, limit := range ob.Asks() {
		for _, order := range limit.Orders {
			o := md.Order{
				UserID:    order.UserID,
				ID:        order.ID,
				Price:     limit.Price,
				Size:      order.Size,
				Bid:       order.Bid,
				Timestamp: order.Timestamp,
			}
			orderbookData.Asks = append(orderbookData.Asks, &o)
		}
	}

	for _, limit := range ob.Bids() {
		for _, order := range limit.Orders {
			o := md.Order{
				UserID:    order.UserID,
				ID:        order.ID,
				Price:     limit.Price,
				Size:      order.Size,
				Bid:       order.Bid,
				Timestamp: order.Timestamp,
			}
			orderbookData.Bids = append(orderbookData.Bids, &o)
		}
	}

	return c.JSON(http.StatusOK, orderbookData)
}

type PriceResponse struct {
	Price float64
}

// GetBestBid finds the best (highest) bid for a specified market.
func (ex *Exchange) GetBestBid(c echo.Context) error {
	var (
		market = md.Market(c.Param("market"))
		ob     = ex.bidBk[market]
		order  = md.Order{}
	)

	if len(ob.Bids()) == 0 {
		return c.JSON(http.StatusOK, order)
	}

	bestLimit := ob.Bids()[0]
	bestOrder := bestLimit.Orders[0]

	order.Price = bestLimit.Price
	order.UserID = bestOrder.UserID

	return c.JSON(http.StatusOK, order)
}

// GetBestAsk finds the best (lowest) ask for a specified market.
func (ex *Exchange) GetBestAsk(c echo.Context) error {
	var (
		market = md.Market(c.Param("market"))
		ob     = ex.bidBk[market]
		order  = md.Order{}
	)

	if len(ob.Asks()) == 0 {
		return c.JSON(http.StatusOK, order)
	}

	bestLimit := ob.Asks()[0]
	bestOrder := bestLimit.Orders[0]

	order.Price = bestLimit.Price
	order.UserID = bestOrder.UserID

	return c.JSON(http.StatusOK, order)
}

// CancelOrder cancels a specific bid by its ID in the ETH market.
func (ex *Exchange) CancelBid(c echo.Context) error {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	ob := ex.bidBk[MarketETH]
	bid := ob.Orders[int64(id)]
	ob.CancelOrder(bid)

	log.Println("order canceled id => ", id)

	return c.JSON(200, map[string]any{"msg": "bid deleted"})
}

// placeMarketOrder processes a market bid, matching it with existing bid in the book.
func (ex *Exchange) placeMarketOrder(market md.Market, bid *bd.Bid) ([]bd.Match, []*md.MatchedBid) {
	ob := ex.bidBk[market]
	matches := ob.PlaceMarketOrder(bid)
	matchedBid := make([]*md.MatchedBid, len(matches))

	isBid := false
	if bid.Bid {
		isBid = true
	}

	totalSizeFilled := 0.0
	sumPrice := 0.0
	for i := 0; i < len(matchedBid); i++ {
		id := matches[i].Bid.ID
		limitUserID := matches[i].Bid.UserID
		if isBid {
			limitUserID = matches[i].Ask.UserID
			id = matches[i].Ask.ID
		}

		matchedBid[i] = &md.MatchedBid{
			UserID: limitUserID,
			ID:     id,
			Size:   matches[i].SizeFilled,
			Price:  matches[i].Price,
		}

		totalSizeFilled += matches[i].SizeFilled
		sumPrice += matches[i].Price
	}

	avgPrice := sumPrice / float64(len(matches))

	logrus.WithFields(logrus.Fields{
		"type":     bid.Type(),
		"size":     totalSizeFilled,
		"avgPrice": avgPrice,
	}).Info("filled market bid")

	newOrderMap := make(map[int64][]*bd.Bid)

	ex.mu.Lock()
	for userID, bidBK := range ex.bids {
		for i := 0; i < len(bidBK); i++ {
			// If the bid is not filled we place it in the map copy.
			// this means that size of the bids = 0
			if !bidBK[i].IsFilled() {
				newOrderMap[userID] = append(newOrderMap[userID], bidBK[i])
			}
		}
	}
	ex.bids = newOrderMap
	ex.mu.Unlock()

	return matches, matchedBid
}

// placeLimitOrder adds a limit bid to the bid book for a specific market.
func (ex *Exchange) placeLimitOrder(market md.Market, price float64, bid *bd.Bid) error {
	ob := ex.bidBk[market]
	ob.PlaceLimitOrder(price, bid)

	// keep track of the user bid
	ex.mu.Lock()
	ex.bids[bid.UserID] = append(ex.bids[bid.UserID], bid)
	ex.mu.Unlock()

	return nil
}

type PlaceOrderResponse struct {
	OrderID int64
}

// PlaceOrder processes an bid request from a user, placing either a market or limit bid.
func (ex *Exchange) PlaceOrder(c echo.Context) error {
	var placeOrderData md.PlaceOrderRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&placeOrderData); err != nil {
		return err
	}

	market := md.Market(placeOrderData.Market)
	order := bd.NewBid(placeOrderData.Bid, placeOrderData.Size, placeOrderData.UserID)

	// Limit bids
	if placeOrderData.Type == LimitOrder {
		if err := ex.placeLimitOrder(market, placeOrderData.Price, order); err != nil {
			return err
		}
	}

	// market bid
	if placeOrderData.Type == MarketOrder {
		matches, _ := ex.placeMarketOrder(market, order)
		if err := ex.matches(matches); err != nil {
			return err
		}
	}

	resp := &PlaceOrderResponse{
		OrderID: order.ID,
	}

	return c.JSON(200, resp)
}

// matches updates user balances based on the results of executed trades.
func (ex *Exchange) matches(matches []bd.Match) error {
	for _, match := range matches {
		fromUser, ok := ex.Users[match.Ask.UserID]
		if !ok {
			return fmt.Errorf("user not found: %d", match.Ask.UserID)
		}

		toUser, ok := ex.Users[match.Bid.UserID]
		if !ok {
			return fmt.Errorf("user not found: %d", match.Bid.UserID)
		}
		toAddresss := crypto.PubkeyToAddress(toUser.PrivateKey.PublicKey)

		// this is only used for the fees
		// exchangePubKey := ex.PrivateKey.Public()
		// publicKeyECDSA, ok := exchangePubKey.(*ecdsa.PublicKey)
		// if !ok {
		// 	return fmt.Errorf("error casting public key to ECDSA")
		// }

		amount := big.NewInt(int64(match.SizeFilled))
		transferETH(ex.Client, fromUser.PrivateKey, toAddresss, amount)
	}

	return nil
}

// transferETH handles the Ethereum transaction for transferring ETH from one user to another.
func transferETH(client *ethclient.Client, fromPrivKey *ecdsa.PrivateKey, to common.Address, amount *big.Int) error {
	ctx := context.Background()
	publicKey := fromPrivKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return fmt.Errorf("error casting public key to ECDSA")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		return err
	}

	gasLimit := uint64(21000)
	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		log.Fatal(err)
	}

	tx := types.NewTransaction(nonce, to, amount, gasLimit, gasPrice, nil)

	chainID := big.NewInt(1337)
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), fromPrivKey)
	if err != nil {
		return err
	}

	return client.SendTransaction(ctx, signedTx)
}
