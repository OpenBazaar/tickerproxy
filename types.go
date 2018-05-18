package ticker

type exchangeRateType int

const (
	exchangeRateTypeFiat exchangeRateType = iota
	exchangeRateTypeCrypto
)

func (t exchangeRateType) String() string {
	switch t {
	case exchangeRateTypeFiat:
		return "fiat"
	case exchangeRateTypeCrypto:
		return "crypto"
	}
	return ""
}
