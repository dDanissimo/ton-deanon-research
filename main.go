//     A program designed to parse TON Blockchain transaction list related to the "Anonymous Telegram Numbers" NFT Collection and return the relevant smart contract address for each token in the collection.
//     Copyright (C) 2023  Daniil Aleksandrovich Ivankin

//     This program is free software: you can redistribute it and/or modify
//     it under the terms of the GNU Affero General Public License as published
//     by the Free Software Foundation, either version 3 of the License, or
//     (at your option) any later version.

//     This program is distributed in the hope that it will be useful,
//     but WITHOUT ANY WARRANTY; without even the implied warranty of
//     MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//     GNU Affero General Public License for more details.

//     You should have received a copy of the GNU Affero General Public License
//     along with this program.  If not, see <https://www.gnu.org/licenses/>.

package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/ton/nft"
)

func main() {
	client := liteclient.NewConnectionPool()
	err := client.AddConnectionsFromConfigUrl(context.Background(), "https://ton-blockchain.github.io/global.config.json")
	if err != nil {
		panic(err)
	}
	api := ton.NewAPIClient(client)

	b, err := api.CurrentMasterchainInfo(context.Background())
	if err != nil {
		log.Fatalln("get block err:", err.Error())
		return
	}

	numbersCollectionAddress := address.MustParseAddr("EQAOQdwdw8kGftJCSFgOErM1mBjYPe4DBPq8-AhF6vr9si5N")
	res, err := api.GetAccount(context.Background(), b, numbersCollectionAddress)
	if err != nil {
		log.Fatalln("get account err:", err.Error())
		return
	}

	// @see https://github.com/xssnick/tonutils-go/blob/v1.4.1/example/account-state/main.go
	lastHash := res.LastTxHash
	lastLt := res.LastTxLT
	for {
		if lastLt == 0 {
			break
		}

		list, err := api.ListTransactions(context.Background(), numbersCollectionAddress, 1000, lastLt, lastHash)
		if err != nil {
			log.Printf("send err: %s", err.Error())
			return
		}

		for _, t := range list {
			// fmt.Println(t.String())

			for _, m := range t.IO.Out {
				phoneAddr := m.Msg.DestAddr()

				// @see https://github.com/xssnick/tonutils-go/blob/master/example/nft-info/main.go
				item := nft.NewItemClient(api, phoneAddr)
				nftData, err := item.GetNFTData(context.Background())
				if err == nil {
					if nftData.Initialized {
						contentURI := nftData.Content.(*nft.ContentOffchain).URI
						rawPhone := strings.TrimSuffix(strings.TrimPrefix(contentURI, "https://nft.fragment.com/number/"), ".json")
						fmt.Printf("%v\thttps://tonscan.org/address/%v\n", rawPhone, phoneAddr.String())
					}
				}
			}
		}
		lastHash = list[0].PrevTxHash
		lastLt = list[0].PrevTxLT
	}
}
