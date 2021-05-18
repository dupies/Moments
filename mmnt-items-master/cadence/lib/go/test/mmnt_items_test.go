package test

import (
	"strings"
	"testing"

	"github.com/onflow/cadence"
	jsoncdc "github.com/onflow/cadence/encoding/json"
	emulator "github.com/onflow/flow-emulator"
	sdk "github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	sdktemplates "github.com/onflow/flow-go-sdk/templates"
	"github.com/onflow/flow-go-sdk/test"

	"github.com/stretchr/testify/assert"

	"github.com/onflow/flow-go-sdk"

	nft_contracts "github.com/onflow/flow-nft/lib/go/contracts"
)

const (
	mmntItemsRootPath                   = "../../.."
	mmntItemsMmntItemsPath             = mmntItemsRootPath + "/contracts/MmntItems.cdc"
	mmntItemsSetupAccountPath           = mmntItemsRootPath + "/transactions/setup_account.cdc"
	mmntItemsMintMmntItemPath          = mmntItemsRootPath + "/transactions/mint_mmnt_item.cdc"
	mmntItemsTransferMmntItemPath      = mmntItemsRootPath + "/transactions/transfer_mmnt_item.cdc"
	mmntItemsInspectMmntItemSupplyPath = mmntItemsRootPath + "/scripts/read_mmnt_items_supply.cdc"
	mmntItemsInspectCollectionLenPath   = mmntItemsRootPath + "/scripts/read_collection_length.cdc"
	mmntItemsInspectCollectionIdsPath   = mmntItemsRootPath + "/scripts/read_collection_ids.cdc"

	typeID1 = 1000
	typeID2 = 2000
)

func MmntItemsDeployContracts(b *emulator.Blockchain, t *testing.T) (flow.Address, flow.Address, crypto.Signer) {
	accountKeys := test.AccountKeyGenerator()

	// Should be able to deploy a contract as a new account with no keys.
	nftCode := loadNonFungibleToken()
	nftAddr, err := b.CreateAccount(
		nil,
		[]sdktemplates.Contract{
			{
				Name:   "NonFungibleToken",
				Source: string(nftCode),
			},
		})
	if !assert.NoError(t, err) {
		t.Log(err.Error())
	}
	_, err = b.CommitBlock()
	assert.NoError(t, err)

	// Should be able to deploy a contract as a new account with one key.
	mmntItemsAccountKey, mmntItemsSigner := accountKeys.NewWithSigner()
	mmntItemsCode := loadMmntItems(nftAddr.String())
	mmntItemsAddr, err := b.CreateAccount(
		[]*flow.AccountKey{mmntItemsAccountKey},
		[]sdktemplates.Contract{
			{
				Name:   "MmntItems",
				Source: string(mmntItemsCode),
			},
		})
	if !assert.NoError(t, err) {
		t.Log(err.Error())
	}
	_, err = b.CommitBlock()
	assert.NoError(t, err)

	// Simplify the workflow by having the contract address also be our initial test collection.
	MmntItemsSetupAccount(t, b, mmntItemsAddr, mmntItemsSigner, nftAddr, mmntItemsAddr)

	return nftAddr, mmntItemsAddr, mmntItemsSigner
}

func MmntItemsSetupAccount(t *testing.T, b *emulator.Blockchain, userAddress sdk.Address, userSigner crypto.Signer, nftAddr sdk.Address, mmntItemsAddr sdk.Address) {
	tx := flow.NewTransaction().
		SetScript(mmntItemsGenerateSetupAccountScript(nftAddr.String(), mmntItemsAddr.String())).
		SetGasLimit(100).
		SetProposalKey(b.ServiceKey().Address, b.ServiceKey().Index, b.ServiceKey().SequenceNumber).
		SetPayer(b.ServiceKey().Address).
		AddAuthorizer(userAddress)

	signAndSubmit(
		t, b, tx,
		[]flow.Address{b.ServiceKey().Address, userAddress},
		[]crypto.Signer{b.ServiceKey().Signer(), userSigner},
		false,
	)
}

func MmntItemsCreateAccount(t *testing.T, b *emulator.Blockchain, nftAddr sdk.Address, mmntItemsAddr sdk.Address) (sdk.Address, crypto.Signer) {
	userAddress, userSigner, _ := createAccount(t, b)
	MmntItemsSetupAccount(t, b, userAddress, userSigner, nftAddr, mmntItemsAddr)
	return userAddress, userSigner
}

func MmntItemsMintItem(b *emulator.Blockchain, t *testing.T, nftAddr, mmntItemsAddr flow.Address, mmntItemsSigner crypto.Signer, typeID uint64) {
	tx := flow.NewTransaction().
		SetScript(mmntItemsGenerateMintMmntItemScript(nftAddr.String(), mmntItemsAddr.String())).
		SetGasLimit(100).
		SetProposalKey(b.ServiceKey().Address, b.ServiceKey().Index, b.ServiceKey().SequenceNumber).
		SetPayer(b.ServiceKey().Address).
		AddAuthorizer(mmntItemsAddr)
	tx.AddArgument(cadence.NewAddress(mmntItemsAddr))
	tx.AddArgument(cadence.NewUInt64(typeID))

	signAndSubmit(
		t, b, tx,
		[]flow.Address{b.ServiceKey().Address, mmntItemsAddr},
		[]crypto.Signer{b.ServiceKey().Signer(), mmntItemsSigner},
		false,
	)
}

func MmntItemsTransferItem(b *emulator.Blockchain, t *testing.T, nftAddr, mmntItemsAddr flow.Address, mmntItemsSigner crypto.Signer, typeID uint64, recipientAddr flow.Address, shouldFail bool) {
	tx := flow.NewTransaction().
		SetScript(mmntItemsGenerateTransferMmntItemScript(nftAddr.String(), mmntItemsAddr.String())).
		SetGasLimit(100).
		SetProposalKey(b.ServiceKey().Address, b.ServiceKey().Index, b.ServiceKey().SequenceNumber).
		SetPayer(b.ServiceKey().Address).
		AddAuthorizer(mmntItemsAddr)
	tx.AddArgument(cadence.NewAddress(recipientAddr))
	tx.AddArgument(cadence.NewUInt64(typeID))

	signAndSubmit(
		t, b, tx,
		[]flow.Address{b.ServiceKey().Address, mmntItemsAddr},
		[]crypto.Signer{b.ServiceKey().Signer(), mmntItemsSigner},
		shouldFail,
	)
}

func TestMmntItemsDeployContracts(t *testing.T) {
	b := newEmulator()
	MmntItemsDeployContracts(b, t)
}

func TestCreateMmntItem(t *testing.T) {
	b := newEmulator()

	nftAddr, mmntItemsAddr, mmntItemsSigner := MmntItemsDeployContracts(b, t)

	supply := executeScriptAndCheck(t, b, mmntItemsGenerateInspectMmntItemSupplyScript(nftAddr.String(), mmntItemsAddr.String()), nil)
	assert.Equal(t, cadence.NewUInt64(0), supply.(cadence.UInt64))

	len := executeScriptAndCheck(
		t,
		b,
		mmntItemsGenerateInspectCollectionLenScript(nftAddr.String(), mmntItemsAddr.String()),
		[][]byte{jsoncdc.MustEncode(cadence.NewAddress(mmntItemsAddr))},
	)
	assert.Equal(t, cadence.NewInt(0), len.(cadence.Int))

	t.Run("Should be able to mint a mmntItems", func(t *testing.T) {
		MmntItemsMintItem(b, t, nftAddr, mmntItemsAddr, mmntItemsSigner, typeID1)

		// Assert that the account's collection is correct
		len := executeScriptAndCheck(
			t,
			b,
			mmntItemsGenerateInspectCollectionLenScript(nftAddr.String(), mmntItemsAddr.String()),
			[][]byte{jsoncdc.MustEncode(cadence.NewAddress(mmntItemsAddr))},
		)
		assert.Equal(t, cadence.NewInt(1), len.(cadence.Int))

		// Assert that the token type is correct
		/*typeID := executeScriptAndCheck(
			t,
			b,
			mmntItemsGenerateInspectMmntItemTypeIDScript(nftAddr.String(), mmntItemsAddr.String()),
			// Cheat: We know it's token ID 0
			[][]byte{jsoncdc.MustEncode(cadence.NewUInt64(0))},
		)
		assert.Equal(t, cadence.NewUInt64(typeID1), typeID.(cadence.UInt64))*/
	})

	/*t.Run("Shouldn't be able to borrow a reference to an NFT that doesn't exist", func(t *testing.T) {
		// Assert that the account's collection is correct
		result, err := b.ExecuteScript(mmntItemsGenerateInspectCollectionScript(nftAddr, mmntItemsAddr, mmntItemsAddr, "MmntItems", "MmntItemsCollection", 5), nil)
		require.NoError(t, err)
		assert.True(t, result.Reverted())
	})*/
}

func TestTransferNFT(t *testing.T) {
	b := newEmulator()

	nftAddr, mmntItemsAddr, mmntItemsSigner := MmntItemsDeployContracts(b, t)

	userAddress, userSigner, _ := createAccount(t, b)

	// create a new Collection
	t.Run("Should be able to create a new empty NFT Collection", func(t *testing.T) {
		MmntItemsSetupAccount(t, b, userAddress, userSigner, nftAddr, mmntItemsAddr)

		len := executeScriptAndCheck(
			t,
			b, mmntItemsGenerateInspectCollectionLenScript(nftAddr.String(), mmntItemsAddr.String()),
			[][]byte{jsoncdc.MustEncode(cadence.NewAddress(userAddress))},
		)
		assert.Equal(t, cadence.NewInt(0), len.(cadence.Int))

	})

	t.Run("Shouldn't be able to withdraw an NFT that doesn't exist in a collection", func(t *testing.T) {
		MmntItemsTransferItem(b, t, nftAddr, mmntItemsAddr, mmntItemsSigner, 3333333, userAddress, true)

		//executeScriptAndCheck(t, b, mmntItemsGenerateInspectCollectionLenScript(nftAddr, mmntItemsAddr, userAddress, "MmntItems", "MmntItemsCollection", 0))

		// Assert that the account's collection is correct
		//executeScriptAndCheck(t, b, mmntItemsGenerateInspectCollectionLenScript(nftAddr, mmntItemsAddr, mmntItemsAddr, "MmntItems", "MmntItemsCollection", 1))
	})

	// transfer an NFT
	t.Run("Should be able to withdraw an NFT and deposit to another accounts collection", func(t *testing.T) {
		MmntItemsMintItem(b, t, nftAddr, mmntItemsAddr, mmntItemsSigner, typeID1)
		// Cheat: we have minted one item, its ID will be zero
		MmntItemsTransferItem(b, t, nftAddr, mmntItemsAddr, mmntItemsSigner, 0, userAddress, false)

		// Assert that the account's collection is correct
		//executeScriptAndCheck(t, b, mmntItemsGenerateInspectCollectionScript(nftAddr, mmntItemsAddr, userAddress, "MmntItems", "MmntItemsCollection", 0))

		//executeScriptAndCheck(t, b, mmntItemsGenerateInspectCollectionLenScript(nftAddr, mmntItemsAddr, userAddress, "MmntItems", "MmntItemsCollection", 1))

		// Assert that the account's collection is correct
		//executeScriptAndCheck(t, b, mmntItemsGenerateInspectCollectionLenScript(nftAddr, mmntItemsAddr, mmntItemsAddr, "MmntItems", "MmntItemsCollection", 0))
	})

	// transfer an NFT
	/*t.Run("Should be able to withdraw an NFT and destroy it, not reducing the supply", func(t *testing.T) {
		tx := flow.NewTransaction().
			SetScript(mmntItemsGenerateDestroyScript(nftAddr, mmntItemsAddr, "MmntItems", "MmntItemsCollection", 0)).
			SetGasLimit(100).
			SetProposalKey(b.ServiceKey().Address, b.ServiceKey().Index, b.ServiceKey().SequenceNumber).
			SetPayer(b.ServiceKey().Address).
			AddAuthorizer(userAddress)

		signAndSubmit(
			t, b, tx,
			[]flow.Address{b.ServiceKey().Address, userAddress},
			[]crypto.Signer{b.ServiceKey().Signer(), userSigner},
			false,
		)

		executeScriptAndCheck(t, b, mmntItemsGenerateInspectCollectionLenScript(nftAddr, mmntItemsAddr, userAddress, "MmntItems", "MmntItemsCollection", 0))

		// Assert that the account's collection is correct
		executeScriptAndCheck(t, b, mmntItemsGenerateInspectCollectionLenScript(nftAddr, mmntItemsAddr, mmntItemsAddr, "MmntItems", "MmntItemsCollection", 0))

		executeScriptAndCheck(t, b, mmntItemsGenerateInspectNFTSupplyScript(nftAddr, mmntItemsAddr, "MmntItems", 1))

	})*/
}

func replaceMmntItemsAddressPlaceholders(code, nftAddress, mmntItemsAddress string) []byte {
	return []byte(replaceStrings(
		code,
		map[string]string{
			nftAddressPlaceholder:        "0x" + nftAddress,
			mmntItemsAddressPlaceHolder: "0x" + mmntItemsAddress,
		},
	))
}

func loadNonFungibleToken() []byte {
	return nft_contracts.NonFungibleToken()
}

func loadMmntItems(nftAddr string) []byte {
	return []byte(strings.ReplaceAll(
		string(readFile(mmntItemsMmntItemsPath)),
		nftAddressPlaceholder,
		"0x"+nftAddr,
	))
}

func mmntItemsGenerateSetupAccountScript(nftAddr, mmntItemsAddr string) []byte {
	return replaceMmntItemsAddressPlaceholders(
		string(readFile(mmntItemsSetupAccountPath)),
		nftAddr,
		mmntItemsAddr,
	)
}

func mmntItemsGenerateMintMmntItemScript(nftAddr, mmntItemsAddr string) []byte {
	return replaceMmntItemsAddressPlaceholders(
		string(readFile(mmntItemsMintMmntItemPath)),
		nftAddr,
		mmntItemsAddr,
	)
}

func mmntItemsGenerateTransferMmntItemScript(nftAddr, mmntItemsAddr string) []byte {
	return replaceMmntItemsAddressPlaceholders(
		string(readFile(mmntItemsTransferMmntItemPath)),
		nftAddr,
		mmntItemsAddr,
	)
}

func mmntItemsGenerateInspectMmntItemSupplyScript(nftAddr, mmntItemsAddr string) []byte {
	return replaceMmntItemsAddressPlaceholders(
		string(readFile(mmntItemsInspectMmntItemSupplyPath)),
		nftAddr,
		mmntItemsAddr,
	)
}

func mmntItemsGenerateInspectCollectionLenScript(nftAddr, mmntItemsAddr string) []byte {
	return replaceMmntItemsAddressPlaceholders(
		string(readFile(mmntItemsInspectCollectionLenPath)),
		nftAddr,
		mmntItemsAddr,
	)
}

func mmntItemsGenerateInspectCollectionIdsScript(nftAddr, mmntItemsAddr string) []byte {
	return replaceMmntItemsAddressPlaceholders(
		string(readFile(mmntItemsInspectCollectionIdsPath)),
		nftAddr,
		mmntItemsAddr,
	)
}
