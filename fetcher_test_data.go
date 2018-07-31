package ticker

import "regexp"

var httpMocks = map[string]string{
	btcavgFiatEndpoint: `{
		"BTCUSD": {"ask": "1","bid": "2","last": "3"},
		"NOTABTCRATE": {}
	}`,

	btcavgCryptoEndpoint: `{
		"BCHBTC": {"ask": "0.5","0.5": "0.5","last": "0.5"},
		"NOTBTC": {"ask": "121","bid": "122","last": "123"},
		"SOILBTC": {"ask": "0.0012345","bid": "0.0012345","last": "0.0012345"},
		"IOTABTC": {"ask": "0.00102","bid": "0.00102","last": "0.00102"},
		"ACCBTC": {"ask": "0.002225","bid": "0.002225","last": "0.002225"},
		"ZEROBTC": {"ask": "0","0": "0","last": "0"},
		"NOTANALTCOINRATE": {}
	}`,

	buildCMCQueryEndpoint(cmcQueryFirstID, defaultCMCQueryLimit): `{
		"metadata": {"num_cryptocurrencies": 102},
		"data": {
			"1": {
				"id": 1,
				"symbol": "SOIL",
				"quotes": {"BTC": {"price": 0.0012345}}
			},
			"1831": {
				"id": 1831,
				"symbol": "BCH",
				"quotes": {"BTC": {"price": 0.5}}
			},
			"2224": {
				"id": 2224,
				"symbol": "ACC",
				"quotes": {"BTC": {"price": 0.002224}}
			},
			"2225": {
				"id": 2225,
				"symbol": "ACC",
				"quotes": {"BTC": {"price": 0.002225}}
			},
			"2226": {
				"id": 2226,
				"symbol": "ACC",
				"quotes": {"BTC": {"price": 0.002226}}
			}
		}
	}`,

	buildCMCQueryEndpoint(cmcQueryFirstID+defaultCMCQueryLimit, defaultCMCQueryLimit): `{
		"metadata": {"num_cryptocurrencies": 102},
		"data": {
			"101": {
				"id": 101,
				"symbol": "$$$",
				"quotes": {"BTC": {"price": 0.101}}
			},
			"102": {
				"id": 102,
				"symbol": "IOTA",
				"quotes": {"BTC": {"price": 0.00102}}
			},
			"103": {
				"id": 103,
				"symbol": "EMPTYPRICE",
				"quotes": {"BTC": {}}
			}
		}
	}`,
}

var testExpectedFetchData = regexp.MustCompile("\\s").ReplaceAllString(`{
	"$$$": {
			"ask": 9.9009905,
			"bid": 9.9009905,
			"last": 9.9009905,
			"type": "crypto"
	},
	"ACC": {
			"ask": 449.4382,
			"bid": 449.4382,
			"last": 449.4382,
			"type": "crypto"
	},
	"BCH": {
			"ask": 2,
			"bid": 2,
			"last": 2,
			"type": "crypto"
	},
	"BTC": {
			"ask": 1,
			"bid": 1,
			"last": 1,
			"type": "crypto"
	},
	"MIOTA": {
			"ask": 980.39215,
			"bid": 980.39215,
			"last": 980.39215,
			"type": "crypto"
	},
	"NOT": {
			"ask": 0.008264462,
			"bid": 0.008196721,
			"last": 0.008130081,
			"type": "crypto"
	},
	"SOIL": {
			"ask": 810.04456,
			"bid": 810.04456,
			"last": 810.04456,
			"type": "crypto"
	},
	"USD": {
			"ask": 1,
			"bid": 2,
			"last": 3,
			"type": "fiat"
	}
}`, "")
