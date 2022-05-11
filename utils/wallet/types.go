package wallet

import (
	"gitlab.com/scpcorp/ScPrime/crypto"
	"gitlab.com/scpcorp/ScPrime/types"
)

// Some of these funcs and consts are copied directly from the ScPrime code
// base. This was done instead of importing the packages directly because wasm
// can not support some of the ScPrime package dependencies such as boltdb.
const (
	// SeedChecksumSize is the number of bytes that are used to checksum
	// addresses to prevent accidental spending
	SeedChecksumSize = 6
)

type (
	// Seed is cryptographic entropy that is used to derive unlock conditions.
	Seed [crypto.EntropySize]byte

	// SpendableKey a set of secret keys plus the corresponding unlock conditions.
	SpendableKey struct {
		UnlockConditions types.UnlockConditions `json:"unlock_conditions"`
		SecretKeys       []crypto.SecretKey     `json:"private_key"`
	}

	// UnlockConditions are a set of conditions which must be met to execute
	// certain actions, such as spending an ScPrime coin output or terminating
	// a FileContract.
	//
	// The simplest requirement is that the block containing the UnlockConditions
	// must have a height >= 'Timelock'.
	//
	// 'PublicKeys' specifies the set of keys that can be used to satisfy the
	// UnlockConditions; of these, at least 'SignaturesRequired' unique keys must sign
	// the transaction. The keys that do not need to use the same cryptographic
	// algorithm.
	//
	// If 'SignaturesRequired' == 0, the UnlockConditions are effectively "anyone can
	// unlock." If 'SignaturesRequired' > len('PublicKeys'), then the UnlockConditions
	// cannot be fulfilled under any circumstances.
	UnlockConditions struct {
		PublicKeys         []string `json:"publickeys"`
		SignaturesRequired uint64   `json:"signaturesrequired"`
		Timelock           uint64   `json:"timelock"`
	}
)
