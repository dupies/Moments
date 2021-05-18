package test

import (
	"strings"
	"testing"

	"github.com/onflow/cadence"
	emulator "github.com/onflow/flow-emulator"
	"github.com/onflow/flow-go-sdk"
	sdk "github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	sdktemplates "github.com/onflow/flow-go-sdk/templates"
	"github.com/onflow/flow-go-sdk/test"
	"github.com/stretchr/testify/assert"
)

const (
	mmntItemsMarketRootPath             = "../../.."
	mmntItemsMarketMmntItemsMarketPath = mmntItemsMarketRootPath + "/contracts/MmntItemsMarket.cdc"
	mmntItemsMarketSetupAccountPath     = mmntItemsMarketRootPath + "/transactions/setup_account.cdc"
	mmntItemsMarketSellItemPath         = mmntItemsMarketRootPath + "/transactions/sell_market_item.cdc"
	mmntItemsMarketBuyItemPath          = mmntItemsMarketRootPath + "/transactions/buy_market_item.cdc"
	mmntItemsMarketRemoveItemPath       = mmntItemsMarketRootPath + "/transactions/remove_market_item.cdc"
)

const (
	typeID1337 = 1337
)

type TestContractsInfo struct {
	FTAddr                 flow.Address
	MomentAddr             flow.Address
	MomentSigner           crypto.Signer
	NFTAddr                flow.Address
	MmntItemsAddr         flow.Address
	MmntItemsSigner       crypto.Signer
	MmntItemsMarketAddr   flow.Address
	MmntItemsMarketSigner crypto.Signer
}

func MmntItemsMarketDeployContracts(b *emulator.Blockchain, t *testing.T) TestContractsInfo {
	accountKeys := test.AccountKeyGenerator()

	ftAddr, momentAddr, momentSigner := MomentDeployContracts(b, t)
	nftAddr, mmntItemsAddr, mmntItemsSigner := MmntItemsDeployContracts(b, t)

	// Should be able to deploy a contract as a new account with one key.
	mmntItemsMarketAccountKey, mmntItemsMarketSigner := accountKeys.NewWithSigner()
	mmntItemsMarketCode := loadMmntItemsMarket(
		ftAddr.String(),
		nftAddr.String(),
		momentAddr.String(),
		mmntItemsAddr.String(),
	)
	mmntItemsMarketAddr, err := b.CreateAccount(
		[]*flow.AccountKey{mmntItemsMarketAccountKey},
		[]sdktemplates.Contract{
			{
				Name:   "MmntItemsMarket",
				Source: string(mmntItemsMarketCode),
			},
		})
	if !assert.NoError(t, err) {
		t.Log(err.Error())
	}
	_, err = b.CommitBlock()
	assert.NoError(t, err)

	// Simplify the workflow by having contract addresses also be our initial test collections.
	MmntItemsSetupAccount(t, b, mmntItemsAddr, mmntItemsSigner, nftAddr, mmntItemsAddr)
	MmntItemsMarketSetupAccount(b, t, mmntItemsMarketAddr, mmntItemsMarketSigner, mmntItemsMarketAddr)

	return TestContractsInfo{
		ftAddr,
		momentAddr,
		momentSigner,
		nftAddr,
		mmntItemsAddr,
		mmntItemsSigner,
		mmntItemsMarketAddr,
		mmntItemsMarketSigner,
	}
}

func MmntItemsMarketSetupAccount(b *emulator.Blockchain, t *testing.T, userAddress sdk.Address, userSigner crypto.Signer, mmntItemsMarketAddr sdk.Address) {
	tx := flow.NewTransaction().
		SetScript(mmntItemsMarketGenerateSetupAccountScript(mmntItemsMarketAddr.String())).
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

// Create a new account with the Moment and MmntItems resources set up BUT no MmntItemsMarket resource.
func MmntItemsMarketCreatePurchaserAccount(b *emulator.Blockchain, t *testing.T, contracts TestContractsInfo) (sdk.Address, crypto.Signer) {
	userAddress, userSigner, _ := createAccount(t, b)
	MomentSetupAccount(t, b, userAddress, userSigner, contracts.FTAddr, contracts.MomentAddr)
	MmntItemsSetupAccount(t, b, userAddress, userSigner, contracts.NFTAddr, contracts.MmntItemsAddr)
	return userAddress, userSigner
}

// Create a new account with the Moment, MmntItems, and MmntItemsMarket resources set up.
func MmntItemsMarketCreateAccount(b *emulator.Blockchain, t *testing.T, contracts TestContractsInfo) (sdk.Address, crypto.Signer) {
	userAddress, userSigner := MmntItemsMarketCreatePurchaserAccount(b, t, contracts)
	MmntItemsMarketSetupAccount(b, t, userAddress, userSigner, contracts.MmntItemsMarketAddr)
	return userAddress, userSigner
}

func MmntItemsMarketListItem(b *emulator.Blockchain, t *testing.T, contracts TestContractsInfo, userAddress sdk.Address, userSigner crypto.Signer, tokenID uint64, price string, shouldFail bool) {
	tx := flow.NewTransaction().
		SetScript(mmntItemsMarketGenerateSellItemScript(contracts)).
		SetGasLimit(100).
		SetProposalKey(b.ServiceKey().Address, b.ServiceKey().Index, b.ServiceKey().SequenceNumber).
		SetPayer(b.ServiceKey().Address).
		AddAuthorizer(userAddress)
	tx.AddArgument(cadence.NewUInt64(tokenID))
	tx.AddArgument(CadenceUFix64(price))

	signAndSubmit(
		t, b, tx,
		[]flow.Address{b.ServiceKey().Address, userAddress},
		[]crypto.Signer{b.ServiceKey().Signer(), userSigner},
		shouldFail,
	)
}

func MmntItemsMarketPurchaseItem(
	b *emulator.Blockchain,
	t *testing.T,
	contracts TestContractsInfo,
	userAddress sdk.Address,
	userSigner crypto.Signer,
	marketCollectionAddress sdk.Address,
	tokenID uint64,
	shouldFail bool,
) {
	tx := flow.NewTransaction().
		SetScript(mmntItemsMarketGenerateBuyItemScript(contracts)).
		SetGasLimit(100).
		SetProposalKey(b.ServiceKey().Address, b.ServiceKey().Index, b.ServiceKey().SequenceNumber).
		SetPayer(b.ServiceKey().Address).
		AddAuthorizer(userAddress)
	tx.AddArgument(cadence.NewUInt64(tokenID))
	tx.AddArgument(cadence.NewAddress(marketCollectionAddress))

	signAndSubmit(
		t, b, tx,
		[]flow.Address{b.ServiceKey().Address, userAddress},
		[]crypto.Signer{b.ServiceKey().Signer(), userSigner},
		shouldFail,
	)
}

func MmntItemsMarketRemoveItem(
	b *emulator.Blockchain,
	t *testing.T,
	contracts TestContractsInfo,
	userAddress sdk.Address,
	userSigner crypto.Signer,
	marketCollectionAddress sdk.Address,
	tokenID uint64,
	shouldFail bool,
) {
	tx := flow.NewTransaction().
		SetScript(mmntItemsMarketGenerateRemoveItemScript(contracts)).
		SetGasLimit(100).
		SetProposalKey(b.ServiceKey().Address, b.ServiceKey().Index, b.ServiceKey().SequenceNumber).
		SetPayer(b.ServiceKey().Address).
		AddAuthorizer(userAddress)
	tx.AddArgument(cadence.NewUInt64(tokenID))

	signAndSubmit(
		t, b, tx,
		[]flow.Address{b.ServiceKey().Address, userAddress},
		[]crypto.Signer{b.ServiceKey().Signer(), userSigner},
		shouldFail,
	)
}

func TestMmntItemsMarketDeployContracts(t *testing.T) {
	b := newEmulator()
	MmntItemsMarketDeployContracts(b, t)
}

func TestMmntItemsMarketSetupAccount(t *testing.T) {
	b := newEmulator()

	contracts := MmntItemsMarketDeployContracts(b, t)

	t.Run("Should be able to create an empty Collection", func(t *testing.T) {
		userAddress, userSigner, _ := createAccount(t, b)
		MmntItemsMarketSetupAccount(b, t, userAddress, userSigner, contracts.MmntItemsMarketAddr)
	})
}

func TestMmntItemsMarketCreateSaleOffer(t *testing.T) {
	b := newEmulator()

	contracts := MmntItemsMarketDeployContracts(b, t)

	t.Run("Should be able to create a sale offer and list it", func(t *testing.T) {
		tokenToList := uint64(0)
		tokenPrice := "1.11"
		userAddress, userSigner := MmntItemsMarketCreateAccount(b, t, contracts)
		// Contract mints item
		MmntItemsMintItem(
			b,
			t,
			contracts.NFTAddr,
			contracts.MmntItemsAddr,
			contracts.MmntItemsSigner,
			typeID1337,
		)
		// Contract transfers item to another seller account (we don't need to do this)
		MmntItemsTransferItem(
			b,
			t,
			contracts.NFTAddr,
			contracts.MmntItemsAddr,
			contracts.MmntItemsSigner,
			tokenToList,
			userAddress,
			false,
		)
		// Other seller account lists the item
		MmntItemsMarketListItem(
			b,
			t,
			contracts,
			userAddress,
			userSigner,
			tokenToList,
			tokenPrice,
			false,
		)
	})

	t.Run("Should be able to accept a sale offer", func(t *testing.T) {
		tokenToList := uint64(1)
		tokenPrice := "1.11"
		userAddress, userSigner := MmntItemsMarketCreateAccount(b, t, contracts)
		// Contract mints item
		MmntItemsMintItem(
			b,
			t,
			contracts.NFTAddr,
			contracts.MmntItemsAddr,
			contracts.MmntItemsSigner,
			typeID1337,
		)
		// Contract transfers item to another seller account (we don't need to do this)
		MmntItemsTransferItem(
			b,
			t,
			contracts.NFTAddr,
			contracts.MmntItemsAddr,
			contracts.MmntItemsSigner,
			tokenToList,
			userAddress,
			false,
		)
		// Other seller account lists the item
		MmntItemsMarketListItem(
			b,
			t,
			contracts,
			userAddress,
			userSigner,
			tokenToList,
			tokenPrice,
			false,
		)
		buyerAddress, buyerSigner := MmntItemsMarketCreatePurchaserAccount(b, t, contracts)
		// Fund the purchase
		MomentMint(
			t,
			b,
			contracts.FTAddr,
			contracts.MomentAddr,
			contracts.MomentSigner,
			buyerAddress,
			"100.0",
			false,
		)
		// Make the purchase
		MmntItemsMarketPurchaseItem(
			b,
			t,
			contracts,
			buyerAddress,
			buyerSigner,
			userAddress,
			tokenToList,
			false,
		)
	})

	t.Run("Should be able to remove a sale offer", func(t *testing.T) {
		tokenToList := uint64(2)
		tokenPrice := "1.11"
		userAddress, userSigner := MmntItemsMarketCreateAccount(b, t, contracts)
		// Contract mints item
		MmntItemsMintItem(
			b,
			t,
			contracts.NFTAddr,
			contracts.MmntItemsAddr,
			contracts.MmntItemsSigner,
			typeID1337,
		)
		// Contract transfers item to another seller account (we don't need to do this)
		MmntItemsTransferItem(
			b,
			t,
			contracts.NFTAddr,
			contracts.MmntItemsAddr,
			contracts.MmntItemsSigner,
			tokenToList,
			userAddress,
			false,
		)
		// Other seller account lists the item
		MmntItemsMarketListItem(
			b,
			t,
			contracts,
			userAddress,
			userSigner,
			tokenToList,
			tokenPrice,
			false,
		)
		// Make the purchase
		MmntItemsMarketRemoveItem(
			b,
			t,
			contracts,
			userAddress,
			userSigner,
			userAddress,
			tokenToList,
			false,
		)
	})
}

func replaceMmntItemsMarketAddressPlaceholders(codeBytes []byte, contracts TestContractsInfo) []byte {
	code := string(codeBytes)

	code = strings.ReplaceAll(code, ftAddressPlaceholder, "0x"+contracts.FTAddr.String())
	code = strings.ReplaceAll(code, momentAddressPlaceHolder, "0x"+contracts.MomentAddr.String())
	code = strings.ReplaceAll(code, nftAddressPlaceholder, "0x"+contracts.NFTAddr.String())
	code = strings.ReplaceAll(code, mmntItemsAddressPlaceHolder, "0x"+contracts.MmntItemsAddr.String())
	code = strings.ReplaceAll(code, mmntItemsMarketPlaceholder, "0x"+contracts.MmntItemsMarketAddr.String())

	return []byte(code)
}

func loadMmntItemsMarket(ftAddr, nftAddr, momentAddr, mmntItemsAddr string) []byte {
	code := string(readFile(mmntItemsMarketMmntItemsMarketPath))

	code = strings.ReplaceAll(code, ftAddressPlaceholder, "0x"+ftAddr)
	code = strings.ReplaceAll(code, momentAddressPlaceHolder, "0x"+momentAddr)
	code = strings.ReplaceAll(code, nftAddressPlaceholder, "0x"+nftAddr)
	code = strings.ReplaceAll(code, mmntItemsAddressPlaceHolder, "0x"+mmntItemsAddr)

	return []byte(code)
}

func mmntItemsMarketGenerateSetupAccountScript(mmntItemsMarketAddr string) []byte {
	code := string(readFile(mmntItemsMarketSetupAccountPath))

	code = strings.ReplaceAll(code, mmntItemsMarketPlaceholder, "0x"+mmntItemsMarketAddr)

	return []byte(code)
}

func mmntItemsMarketGenerateSellItemScript(contracts TestContractsInfo) []byte {
	return replaceMmntItemsMarketAddressPlaceholders(
		readFile(mmntItemsMarketSellItemPath),
		contracts,
	)
}

func mmntItemsMarketGenerateBuyItemScript(contracts TestContractsInfo) []byte {
	return replaceMmntItemsMarketAddressPlaceholders(
		readFile(mmntItemsMarketBuyItemPath),
		contracts,
	)
}

func mmntItemsMarketGenerateRemoveItemScript(contracts TestContractsInfo) []byte {
	return replaceMmntItemsMarketAddressPlaceholders(
		readFile(mmntItemsMarketRemoveItemPath),
		contracts,
	)
}
