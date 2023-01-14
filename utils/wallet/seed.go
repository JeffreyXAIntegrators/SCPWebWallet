package wallet

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode"

	mnemonics "gitlab.com/NebulousLabs/entropy-mnemonics"
	"gitlab.com/NebulousLabs/fastrand"
	"gitlab.com/scpcorp/ScPrime/crypto"
	"gitlab.com/scpcorp/ScPrime/types"
)

// NewWalletSeed creates a new unique 28 or 29 word wallet seed
func NewWalletSeed() (string, error) {
	var entropy [crypto.EntropySize]byte

	fastrand.Read(entropy[:])

	fullChecksum := crypto.HashObject(entropy)
	checksumSeed := append(entropy[:], fullChecksum[:SeedChecksumSize]...)
	phrase, err := mnemonics.ToPhrase(checksumSeed, mnemonics.DictionaryID("english"))

	if err != nil {
		return "", err
	}

	return phrase.String(), nil
}

// StringToSeed converts a string to a wallet seed.
// Copied from gitlab.com/scpcorp/ScPrime/modules/wallet.go
func StringToSeed(str string, did mnemonics.DictionaryID) (Seed, error) {
	// Ensure the string is all lowercase letters and spaces
	for _, char := range str {
		if unicode.IsUpper(char) {
			return Seed{}, errors.New("seed is not valid: all words must be lowercase")
		}
		if !unicode.IsLetter(char) && !unicode.IsSpace(char) {
			return Seed{}, fmt.Errorf("seed is not valid: illegal character '%v'", char)
		}
	}

	// Decode the string into the checksummed byte slice.
	checksumSeedBytes, err := mnemonics.FromString(str, did)
	if err != nil {
		return Seed{}, err
	}

	// ToDo: Add other languages
	switch {
	case did == "english":
		// Check seed has 28 or 29 words
		if len(strings.Fields(str)) != 28 && len(strings.Fields(str)) != 29 {
			return Seed{}, errors.New("seed is not valid: must be 28 or 29 words")
		}

		// Check for other formatting errors (English only)
		IsFormat := regexp.MustCompile(`^([a-z]{4,12}){1}( {1}[a-z]{4,12}){27,28}$`).MatchString
		if !IsFormat(str) {
			return Seed{}, errors.New("seed is not valid: invalid formatting")
		}
	case did == "german":
	case did == "japanese":
	default:
		return Seed{}, fmt.Errorf("seed is not valid: unsupported dictionary '%v'", did)
	}

	// Ensure the seed is 38 bytes (this check is not too helpful since it doesn't
	// give any hints about what is wrong to the end user, which is why it's the
	// last thing checked)
	if len(checksumSeedBytes) != 38 {
		return Seed{}, fmt.Errorf("seed is not valid: illegal number of bytes '%v'", len(checksumSeedBytes))
	}

	// Copy the seed from the checksummed slice.
	var seed Seed
	copy(seed[:], checksumSeedBytes)
	fullChecksum := crypto.HashObject(seed)
	if len(checksumSeedBytes) != crypto.EntropySize+SeedChecksumSize || !bytes.Equal(fullChecksum[:SeedChecksumSize], checksumSeedBytes[crypto.EntropySize:]) {
		return Seed{}, errors.New("seed failed checksum verification")
	}
	return seed, nil
}

// GetAddress returns the spendable address at the specified index
func GetAddress(seed Seed, index uint64) SpendableKey {
	sk, pk := crypto.GenerateKeyPairDeterministic(crypto.HashAll(seed, index))
	return SpendableKey{
		UnlockConditions: types.UnlockConditions{
			PublicKeys:         []types.SiaPublicKey{types.Ed25519PublicKey(pk)},
			SignaturesRequired: 1,
		},
		SecretKeys: []crypto.SecretKey{sk},
	}
}
