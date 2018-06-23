package ticker

import (
	"encoding/json"
	"log"
)

// AltSymbolsToCanonicalSymbols maps symbols that may be used in some sources to
// represent coins that we use a different symbol for
var AltSymbolsToCanonicalSymbols = map[string]string{
	"MIOTA": "IOTA",
}

// PinnedSymbolsToIDs maps symbols that may be used by multiple coins to a
// single coin by its CMC IDs.
var PinnedSymbolsToIDs = map[string]int64{
	"BTC":  1,    // Bitcoin
	"LTC":  2,    // Litecoin
	"NXT":  66,   // Nxt
	"DOGE": 74,   // Dogecoin
	"DASH": 131,  // Dash
	"XMR":  328,  // Monero
	"ETH":  1027, // Ethereum
	"ZEC":  1437, // Zcash
	"BCH":  1831, // Bitcoin Cash

	"BTG":  2083, // Bitcoin Gold
	"CMT":  2246, // CyberMiles
	"KNC":  1982, // Kyber Network
	"BTM":  1866, // Bytom
	"ICN":  1408, // Iconomi
	"GTC":  2336, // Game.com
	"BLZ":  2505, // Bluzelle
	"HOT":  2682, // Holo
	"RCN":  2096, // Ripio Credit Network
	"FAIR": 224,  // FairCoin
	"EDR":  2835, // Endor Protocol
	"CPC":  2482, // CPChain
	"QBT":  2242, // Qbao
	"KEY":  2398, // Selfkey
	"RED":  2771, // RED
	"HMC":  2484, // Hi Mutual Society
	"NET":  1811, // Nimiq Exchange Token
	"LNC":  2677, // Linker Coin
	"CAN":  2343, // CanYaCoin
	"BET":  1771, // DAO.Casino
	"SPD":  2616, // Stipend
	"CAT":  2334, // BitClave
	"GCC":  1531, // Global Cryptocurrency
	"PUT":  2419, // Profile Utility Token
	"MAG":  2218, // Magnet
	"CRC":  2664, // CryCash
	"ACC":  2225, // Accelerator Network
	"PXC":  35,   // Phoenixcoin
	"ETT":  1714, // EncryptoTel [WAVES]
	"XIN":  2349, // Mixin
	"HERO": 1805, // Sovereign Hero
	"HNC":  1004, // Helleniccoin
	"ENT":  1474, // Eternity
	"LBTC": 1825, // LiteBitcoin
	"CMS":  2262, // COMSA [ETH]
}

var pinnedSymbolsToIDsJSON []byte

// PinnedSymbolsToIDsJSON returns the PinnedSymbolsToIDs marshaled to JSON
func PinnedSymbolsToIDsJSON() []byte {
	return pinnedSymbolsToIDsJSON
}

func init() {
	var err error
	pinnedSymbolsToIDsJSON, err = json.Marshal(&PinnedSymbolsToIDs)
	if err != nil {
		log.Fatalln(err)
	}
}

// CanonicalizeSymbol returns the canonical symbol from the given one, which may
// or may not be a nickname
func CanonicalizeSymbol(symbol string) string {
	if canonicalSymbol, ok := AltSymbolsToCanonicalSymbols[symbol]; ok {
		symbol = canonicalSymbol
	}
	return symbol
}

// IsCorrectIDForSymbol checks if the given id is the correct one for the given
// symbol based on the map `PinnedSymbolsToIDs`
func IsCorrectIDForSymbol(symbol string, id int64) bool {
	pinnedSymbolID, symbolHasDupes := PinnedSymbolsToIDs[symbol]
	if !symbolHasDupes || pinnedSymbolID == id {
		return true
	}
	return false
}
