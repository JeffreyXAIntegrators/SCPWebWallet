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
func ComputeSummarizedTransactions(pts []modules.ProcessedTransaction, blockHeight types.BlockHeight, w modules.Wallet) ([]SummarizedTransaction, error) {
	sts := []SummarizedTransaction{}
	vts, err := wallet.ComputeValuedTransactions(pts, blockHeight)
	if err != nil {
		return nil, err
	}
	utxos, _ := w.UnspentOutputs()
	for _, txn := range vts {
		isSpfTransfer := false
		// Determine the number of outgoing coins and funds.
		var outgoingFunds types.Currency
		for _, input := range txn.Inputs {
			if input.FundType == types.SpecifierSiafundInput && input.WalletAddress {
				outgoingFunds = outgoingFunds.Add(input.Value)
				isSpfTransfer = true
			}
		}
		// Determine the number of incoming funds.
		var incomingFunds types.Currency
		for _, output := range txn.Outputs {
			if output.FundType == types.SpecifierSiafundOutput && output.WalletAddress {
				incomingFunds = incomingFunds.Add(output.Value)
				isSpfTransfer = true
			}
		}
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
		// For funds, need to avoid having a negative types.Currency.
		if !isSpfTransfer {
			st.Scp = fmt.Sprintf("%15.2f SCP", incomingCoinsFloat-outgoingCoinsFloat)
		} else if incomingFunds.Cmp(outgoingFunds) > 0 {
			st.Spf = fmt.Sprintf("%14v SPF", incomingFunds.Sub(outgoingFunds))
			if incomingCoinsFloat-outgoingCoinsFloat > .009 || incomingCoinsFloat-outgoingCoinsFloat < -.009 {
				st.Scp = fmt.Sprintf("%15.2f SCP", incomingCoinsFloat-outgoingCoinsFloat)
			}
		} else if incomingFunds.Cmp(outgoingFunds) < 0 {
			st.Spf = "-" + strings.TrimSpace(fmt.Sprintf("%14v SPF", outgoingFunds.Sub(incomingFunds)))
			if incomingCoinsFloat-outgoingCoinsFloat > .009 || incomingCoinsFloat-outgoingCoinsFloat < -.009 {
				st.Scp = fmt.Sprintf("%15.2f SCP", incomingCoinsFloat-outgoingCoinsFloat)
			}
		} else if isSpfTransfer {
			var incomingClaims types.Currency
			for _, input := range txn.Transaction.SiafundInputs {
				incomingClaims = incomingClaims.Add(siaClaimOutputSum(input, txn.ConfirmationHeight, pts, utxos))
			}
			incomingClaimsFloat, _ := new(big.Rat).SetFrac(incomingClaims.Big(), types.ScPrimecoinPrecision.Big()).Float64()
			txnTotalFloat := incomingClaimsFloat + incomingCoinsFloat - outgoingCoinsFloat
			if incomingClaims.IsZero() && st.Type == "SETUP" {
				st.Type = "COLLECT SPF REWARDS"
				st.Scp = "IN PROGRESS"
			} else if incomingClaims.IsZero() {
				st.Type = "SETUP"
				st.Scp = fmt.Sprintf("%15.2f SCP", float64(0))
			} else {
				st.Type = "COLLECT SPF REWARDS"
				st.Scp = fmt.Sprintf("%15.8f SCP", txnTotalFloat)
			}
		}
		sts = append(sts, st)
	}
	return sts, nil
}

// siaClaimOutputSum returns the sum of all the siacoin claim outputs created
// when the siafund input is spent. The supplied floor defines the earliest
// block that can contain a spent claim input. The supplied txns defines the
// pool of spent transactions that might have spent the claim input. The
// supplied utxos defines the pool of unspent outputs that might contain the
// unspent claim input.
func siaClaimOutputSum(input types.SiafundInput, floor types.BlockHeight, txns []modules.ProcessedTransaction, utxos []modules.UnspentOutput) types.Currency {
	for _, unspent := range utxos {
		if unspent.ID.String() == input.ParentID.SiaClaimOutputID().String() {
			return unspent.Value
		}
	}
	for _, txn := range txns {
		if txn.ConfirmationHeight >= floor {
			for _, spent := range txn.Inputs {
				if spent.ParentID.String() == input.ParentID.SiaClaimOutputID().String() {
					return spent.Value
				}
			}
		}
	}
	return ZeroCurrency
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
