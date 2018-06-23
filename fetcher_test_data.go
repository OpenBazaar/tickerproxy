package ticker

import "regexp"

var httpMocks = map[string]string{
	btcavgFiatEndpoint: `{
		"BTCUSD": {"ask": "1","bid": "2","last": "3"},
		"NOTABTCRATE": {}
	}`,

	btcavgCryptoEndpoint: `{
		"BCHBTC": {"ask": "111","bid": "112","last": "113"},
		"NOTBTC": {"ask": "121","bid": "122","last": "123"},
		"SOILBTC": {"ask": "131","bid": "132","last": "133"},
		"MIOTABTC": {"ask": "141","bid": "142","last": "143"},
		"ACCBTC": {"ask": "151","bid": "152","last": "153"},
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
    "last": 9.9009905
	},
  "ACC": {
    "ask": 449.4382,
    "bid": 449.4382,
    "last": 449.4382
  },
  "BCH": {
    "ask": 2,
    "bid": 2,
    "last": 2
  },
  "BTC": {
    "ask": 1,
    "bid": 1,
    "last": 1
  },
  "IOTA": {
    "ask": 980.39215,
    "bid": 980.39215,
    "last": 980.39215
	},
  "NOT": {
    "ask": 0.008264462,
    "bid": 0.008196721,
    "last": 0.008130081
  },
  "SOIL": {
    "ask": 810.04456,
    "bid": 810.04456,
    "last": 810.04456
  },
  "USD": {
    "ask": 1,
    "bid": 2,
    "last": 3
  }
}`, "")
