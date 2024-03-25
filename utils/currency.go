package utils

// supported currencies Constant
const (
	USD = "USD" 
	EUR = "EUR"
	RMB = "RMB"
)

// IsSupportedCurrency returns true if the currency is supported
func IsSupportedCurrency(currency string) bool {
	switch currency {
	case USD, EUR, RMB:
		return true
	}
	return false
}