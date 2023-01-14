package server

import (
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"gitlab.com/scpcorp/ScPrime/modules"
	"gitlab.com/scpcorp/ScPrime/modules/wallet"
	"gitlab.com/scpcorp/ScPrime/types"
)

const (
	// For an unconfirmed Transaction, the TransactionTimestamp field is set to the
	// maximum value of a uint64.
	unconfirmedTransactionTimestamp = ^uint64(0)
)

var (
	// ErrParseCurrencyAmount is returned when the input is unable to be parsed
	// into a currency unit due to a malformed amount.
	ErrParseCurrencyAmount = errors.New("malformed amount")
	// ErrParseCurrencyInteger is returned when the input is unable to be parsed
	// into a currency unit due to a non-integer value.
	ErrParseCurrencyInteger = errors.New("non-integer number of hastings")
	// ErrParseCurrencyUnits is returned when the input is unable to be parsed
	// into a currency unit due to missing units.
	ErrParseCurrencyUnits = errors.New("amount is missing currency units. Currency units are case sensitive")
	// ErrNegativeCurrency is the error that is returned if performing an
	// operation results in a negative currency.
	ErrNegativeCurrency = errors.New("negative currency not allowed")
	// ErrUint64Overflow is the error that is returned if converting to a
	// unit64 would cause an overflow.
	ErrUint64Overflow = errors.New("cannot return the uint64 of this currency - result is an overflow")
	// ZeroCurrency defines a currency of value zero.
	ZeroCurrency = types.NewCurrency64(0)
)

// SummarizedTransaction is a transaction that has been formatted for·
// humans to read.
type SummarizedTransaction struct {
	TxnID     string `json:"txn_id"`
	Type      string `json:"type"`
	Time      string `json:"time"`
	Confirmed string `json:"confirmed"`
	Scp       string `json:"scp"`
	Spf       string `json:"spf"`
}

// ComputeSummarizedTransactions creates a set of SummarizedTransactions
// from a set of ProcessedTransactions.
func ComputeSummarizedTransactions(pts []modules.ProcessedTransaction, blockHeight types.BlockHeight) ([]SummarizedTransaction, error) {
	sts := []SummarizedTransaction{}
	vts, err := wallet.ComputeValuedTransactions(pts, blockHeight)
	if err != nil {
		return nil, err
	}
	for _, txn := range vts {
		// Determine the number of outgoing coins and funds.
		var outgoingFunds types.Currency
		for _, input := range txn.Inputs {
			if input.FundType == types.SpecifierSiafundInput && input.WalletAddress {
				outgoingFunds = outgoingFunds.Add(input.Value)
			}
		}
		// Determine the number of incoming funds.
		var incomingFunds types.Currency
		for _, output := range txn.Outputs {
			if output.FundType == types.SpecifierSiafundOutput && output.WalletAddress {
				incomingFunds = incomingFunds.Add(output.Value)
			}
		}
		// Convert the scp to a float.
		incomingCoinsFloat, _ := new(big.Rat).SetFrac(txn.ConfirmedIncomingValue.Big(), types.ScPrimecoinPrecision.Big()).Float64()
		outgoingCoinsFloat, _ := new(big.Rat).SetFrac(txn.ConfirmedOutgoingValue.Big(), types.ScPrimecoinPrecision.Big()).Float64()
		// Summarize transaction
		st := SummarizedTransaction{}
		st.TxnID = strings.ToUpper(fmt.Sprintf("%v", txn.TransactionID))
		st.Type = strings.ToUpper(strings.Replace(fmt.Sprintf("%v", txn.TxType), "_", " ", -1))
		if uint64(txn.ConfirmationTimestamp) != unconfirmedTransactionTimestamp {
			st.Time = time.Unix(int64(txn.ConfirmationTimestamp), 0).Format("2006-01-02 15:04")
			st.Confirmed = "Yes"
		} else {
			st.Confirmed = "No"
		}
		st.Scp = fmt.Sprintf("%15.2f SCP", incomingCoinsFloat-outgoingCoinsFloat)
		// For funds, need to avoid having a negative types.Currency.
		if incomingFunds.Cmp(outgoingFunds) > 0 {
			st.Spf = fmt.Sprintf("%14v SPF %v\n", incomingFunds.Sub(outgoingFunds), txn.TxType)
		} else if incomingFunds.Cmp(outgoingFunds) < 0 {
			st.Spf = fmt.Sprintf("-%14v SPF %v\n", outgoingFunds.Sub(incomingFunds), txn.TxType)
		}
		sts = append(sts, st)
	}
	return sts, nil
}

// NewCurrencyStr creates a Currency value from a supplied string with unit suffix.
// Valid unit suffixes are: H, pS, nS, uS, mS, SCP, KS, MS, GS, TS, SPF
// Unit Suffixes are case sensitive.
func NewCurrencyStr(amount string) (types.Currency, error) {
	base := ""
	units := []string{"pS", "nS", "uS", "mS", "SCP", "KS", "MS", "GS", "TS"}
	amount = strings.TrimSpace(amount)
	for i, unit := range units {
		if strings.HasSuffix(amount, unit) {
			// Trim spaces after removing the suffix to allow spaces between the
			// value and the unit.
			value := strings.TrimSpace(strings.TrimSuffix(amount, unit))
			// scan into big.Rat
			r, ok := new(big.Rat).SetString(value)
			if !ok {
				return types.Currency{}, ErrParseCurrencyAmount
			}
			// convert units
			exp := 27 + 3*(int64(i)-4)
			mag := new(big.Int).Exp(big.NewInt(10), big.NewInt(exp), nil)
			r.Mul(r, new(big.Rat).SetInt(mag))
			// r must be an integer at this point
			if !r.IsInt() {
				return types.Currency{}, ErrParseCurrencyInteger
			}
			base = r.RatString()
		}
	}
	// check for hastings separately
	if strings.HasSuffix(amount, "H") {
		base = strings.TrimSpace(strings.TrimSuffix(amount, "H"))
	}
	// check for SPF separately
	if strings.HasSuffix(amount, "SPF") {
		value := strings.TrimSpace(strings.TrimSuffix(amount, "SPF"))
		// scan into big.Rat
		r, ok := new(big.Rat).SetString(value)
		if !ok {
			return types.Currency{}, ErrParseCurrencyAmount
		}
		// r must be an integer at this point
		if !r.IsInt() {
			return types.Currency{}, ErrParseCurrencyInteger
		}
		base = r.RatString()
	}
	if base == "" {
		return types.Currency{}, ErrParseCurrencyUnits
	}
	var currency types.Currency
	_, err := fmt.Sscan(base, &currency)
	if err != nil {
		return types.Currency{}, ErrParseCurrencyAmount
	}
	return currency, nil
}
