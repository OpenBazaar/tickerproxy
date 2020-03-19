package ticker

import "regexp"

const testCMCQueryLimit = 5

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

	buildCMCEndpoint("sandbox"): `{
		"metadata": {"num_cryptocurrencies": 102},
		"data": [
			{
				"id": 1,
				"symbol": "SOIL",
				"quote": {"BTC": {"price": 0.0012345}}
			},
			{
				"id": 1831,
				"symbol": "BCH",
				"quote": {"BTC": {"price": 0.5}}
			},
			{
				"id": 2224,
				"symbol": "ACC",
				"quote": {"BTC": {"price": 0.002224}}
			},
			{
				"id": 2225,
				"symbol": "ACC",
				"quote": {"BTC": {"price": 0.002225}}
			},
			{
				"id": 2226,
				"symbol": "ACC",
				"quote": {"BTC": {"price": 0.002226}}
			}
		]
	}`,

	buildCMCEndpoint("sandbox"): `{
		"metadata": {"num_cryptocurrencies": 102},
		"data": [
			{
				"id": 101,
				"symbol": "$$$",
				"quote": {"BTC": {"price": 0.101}}
			},
			{
				"id": 102,
				"symbol": "IOTA",
				"quote": {"BTC": {"price": 0.00102}}
			}
		]
	}`,
}

var testExpectedFetchData = regexp.MustCompile("\\s").ReplaceAllString(`{
	"$$$": {
			"ask": 9.9009905,
			"bid": 9.9009905,
			"last": 9.9009905,
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
