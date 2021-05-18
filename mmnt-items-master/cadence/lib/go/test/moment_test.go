package test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/onflow/cadence"
	jsoncdc "github.com/onflow/cadence/encoding/json"
	emulator "github.com/onflow/flow-emulator"
	"github.com/onflow/flow-go-sdk"
	sdk "github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/onflow/flow-go-sdk/templates"
	"github.com/onflow/flow-go-sdk/test"

	ft_contracts "github.com/onflow/flow-ft/lib/go/contracts"
)

const (
	momentRootPath           = "../../.."
	momentMomentPath         = momentRootPath + "/contracts/Moment.cdc"
	momentSetupAccountPath   = momentRootPath + "/transactions/setup_account.cdc"
	momentTransferTokensPath = momentRootPath + "/transactions/transfer_tokens.cdc"
	momentMintTokensPath     = momentRootPath + "/transactions/mint_tokens.cdc"
	momentBurnTokensPath     = momentRootPath + "/transactions/burn_tokens.cdc"
	momentGetBalancePath     = momentRootPath + "/scripts/get_balance.cdc"
	momentGetSupplyPath      = momentRootPath + "/scripts/get_supply.cdc"
)

func MomentDeployContracts(b *emulator.Blockchain, t *testing.T) (flow.Address, flow.Address, crypto.Signer) {
	accountKeys := test.AccountKeyGenerator()

	// Should be able to deploy a contract as a new account with no keys.
	fungibleTokenCode := loadFungibleToken()
	fungibleAddr, err := b.CreateAccount(
		[]*flow.AccountKey{},
		[]templates.Contract{templates.Contract{
			Name:   "FungibleToken",
			Source: string(fungibleTokenCode),
		}},
	)
	assert.NoError(t, err)

	_, err = b.CommitBlock()
	assert.NoError(t, err)

	momentAccountKey, momentSigner := accountKeys.NewWithSigner()
	momentCode := loadMoment(fungibleAddr)

	momentAddr, err := b.CreateAccount(
		[]*flow.AccountKey{momentAccountKey},
		[]templates.Contract{templates.Contract{
			Name:   "Moment",
			Source: string(momentCode),
		}},
	)
	assert.NoError(t, err)

	_, err = b.CommitBlock()
	assert.NoError(t, err)

	// Simplify testing by having the contract address also be our initial Vault.
	MomentSetupAccount(t, b, momentAddr, momentSigner, fungibleAddr, momentAddr)

	return fungibleAddr, momentAddr, momentSigner
}

func MomentSetupAccount(t *testing.T, b *emulator.Blockchain, userAddress sdk.Address, userSigner crypto.Signer, fungibleAddr sdk.Address, momentAddr sdk.Address) {
	tx := flow.NewTransaction().
		SetScript(momentGenerateSetupMomentAccountTransaction(fungibleAddr, momentAddr)).
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

func MomentCreateAccount(t *testing.T, b *emulator.Blockchain, fungibleAddr sdk.Address, momentAddr sdk.Address) (sdk.Address, crypto.Signer) {
	userAddress, userSigner, _ := createAccount(t, b)
	MomentSetupAccount(t, b, userAddress, userSigner, fungibleAddr, momentAddr)
	return userAddress, userSigner
}

func MomentMint(t *testing.T, b *emulator.Blockchain, fungibleAddr sdk.Address, momentAddr sdk.Address, momentSigner crypto.Signer, recipientAddress flow.Address, amount string, shouldRevert bool) {
	tx := flow.NewTransaction().
		SetScript(momentGenerateMintMomentTransaction(fungibleAddr, momentAddr)).
		SetGasLimit(100).
		SetProposalKey(b.ServiceKey().Address, b.ServiceKey().Index, b.ServiceKey().SequenceNumber).
		SetPayer(b.ServiceKey().Address).
		AddAuthorizer(momentAddr)

	_ = tx.AddArgument(cadence.NewAddress(recipientAddress))
	_ = tx.AddArgument(CadenceUFix64(amount))

	signAndSubmit(
		t, b, tx,
		[]flow.Address{b.ServiceKey().Address, momentAddr},
		[]crypto.Signer{b.ServiceKey().Signer(), momentSigner},
		shouldRevert,
	)

}

func TestMomentDeployment(t *testing.T) {
	b := newEmulator()

	fungibleAddr, momentAddr, _ := MomentDeployContracts(b, t)

	t.Run("Should have initialized Supply field correctly", func(t *testing.T) {
		supply := executeScriptAndCheck(t, b, momentGenerateGetSupplyScript(fungibleAddr, momentAddr), nil)
		expectedSupply, expectedSupplyErr := cadence.NewUFix64("0.0")
		assert.NoError(t, expectedSupplyErr)
		assert.Equal(t, expectedSupply, supply.(cadence.UFix64))
	})
}

func TestMomentSetupAccount(t *testing.T) {
	b := newEmulator()

	t.Run("Should be able to create empty Vault that doesn't affect supply", func(t *testing.T) {

		fungibleAddr, momentAddr, _ := MomentDeployContracts(b, t)

		userAddress, _ := MomentCreateAccount(t, b, fungibleAddr, momentAddr)

		balance := executeScriptAndCheck(t, b, momentGenerateGetBalanceScript(fungibleAddr, momentAddr), [][]byte{jsoncdc.MustEncode(cadence.Address(userAddress))})
		assert.Equal(t, CadenceUFix64("0.0"), balance)

		supply := executeScriptAndCheck(t, b, momentGenerateGetSupplyScript(fungibleAddr, momentAddr), nil)
		assert.Equal(t, CadenceUFix64("0.0"), supply.(cadence.UFix64))
	})
}

func TestMomentMinting(t *testing.T) {
	b := newEmulator()

	fungibleAddr, momentAddr, momentSigner := MomentDeployContracts(b, t)

	userAddress, _ := MomentCreateAccount(t, b, fungibleAddr, momentAddr)

	t.Run("Shouldn't be able to mint zero tokens", func(t *testing.T) {
		MomentMint(t, b, fungibleAddr, momentAddr, momentSigner, userAddress, "0.0", true)
	})

	t.Run("Should mint tokens, deposit, and update balance and total supply", func(t *testing.T) {
		MomentMint(t, b, fungibleAddr, momentAddr, momentSigner, userAddress, "50.0", false)

		// Assert that the vault's balance is correct
		result, err := b.ExecuteScript(momentGenerateGetBalanceScript(fungibleAddr, momentAddr), [][]byte{jsoncdc.MustEncode(cadence.Address(userAddress))})
		require.NoError(t, err)
		if !assert.True(t, result.Succeeded()) {
			t.Log(result.Error.Error())
		}
		balance := result.Value
		assert.Equal(t, CadenceUFix64("50.0"), balance.(cadence.UFix64))

		// Make sure that the total supply is correct
		supply := executeScriptAndCheck(t, b, momentGenerateGetSupplyScript(fungibleAddr, momentAddr), nil)
		assert.Equal(t, CadenceUFix64("50.0"), supply.(cadence.UFix64))
	})
}

func TestMomentTransfers(t *testing.T) {
	b := newEmulator()

	fungibleAddr, momentAddr, momentSigner := MomentDeployContracts(b, t)

	userAddress, _ := MomentCreateAccount(t, b, fungibleAddr, momentAddr)

	MomentMint(t, b, fungibleAddr, momentAddr, momentSigner, momentAddr, "1000.0", false)

	t.Run("Shouldn't be able to withdraw more than the balance of the Vault", func(t *testing.T) {
		tx := flow.NewTransaction().
			SetScript(momentGenerateTransferVaultScript(fungibleAddr, momentAddr)).
			SetGasLimit(100).
			SetProposalKey(b.ServiceKey().Address, b.ServiceKey().Index, b.ServiceKey().SequenceNumber).
			SetPayer(b.ServiceKey().Address).
			AddAuthorizer(momentAddr)

		_ = tx.AddArgument(CadenceUFix64("30000.0"))
		_ = tx.AddArgument(cadence.NewAddress(userAddress))

		signAndSubmit(
			t, b, tx,
			[]flow.Address{b.ServiceKey().Address, momentAddr},
			[]crypto.Signer{b.ServiceKey().Signer(), momentSigner},
			true,
		)

		// Assert that the vaults' balances are correct
		result, err := b.ExecuteScript(momentGenerateGetBalanceScript(fungibleAddr, momentAddr), [][]byte{jsoncdc.MustEncode(cadence.Address(momentAddr))})
		require.NoError(t, err)
		if !assert.True(t, result.Succeeded()) {
			t.Log(result.Error.Error())
		}
		balance := result.Value
		assert.Equal(t, balance.(cadence.UFix64), CadenceUFix64("1000.0"))

		result, err = b.ExecuteScript(momentGenerateGetBalanceScript(fungibleAddr, momentAddr), [][]byte{jsoncdc.MustEncode(cadence.Address(userAddress))})
		require.NoError(t, err)
		if !assert.True(t, result.Succeeded()) {
			t.Log(result.Error.Error())
		}
		balance = result.Value
		assert.Equal(t, balance.(cadence.UFix64), CadenceUFix64("0.0"))
	})

	t.Run("Should be able to withdraw and deposit tokens from a vault", func(t *testing.T) {
		tx := flow.NewTransaction().
			SetScript(momentGenerateTransferVaultScript(fungibleAddr, momentAddr)).
			SetGasLimit(100).
			SetProposalKey(b.ServiceKey().Address, b.ServiceKey().Index, b.ServiceKey().SequenceNumber).
			SetPayer(b.ServiceKey().Address).
			AddAuthorizer(momentAddr)

		_ = tx.AddArgument(CadenceUFix64("300.0"))
		_ = tx.AddArgument(cadence.NewAddress(userAddress))

		signAndSubmit(
			t, b, tx,
			[]flow.Address{b.ServiceKey().Address, momentAddr},
			[]crypto.Signer{b.ServiceKey().Signer(), momentSigner},
			false,
		)

		// Assert that the vaults' balances are correct
		result, err := b.ExecuteScript(momentGenerateGetBalanceScript(fungibleAddr, momentAddr), [][]byte{jsoncdc.MustEncode(cadence.Address(momentAddr))})
		require.NoError(t, err)
		if !assert.True(t, result.Succeeded()) {
			t.Log(result.Error.Error())
		}
		balance := result.Value
		assert.Equal(t, balance.(cadence.UFix64), CadenceUFix64("700.0"))

		result, err = b.ExecuteScript(momentGenerateGetBalanceScript(fungibleAddr, momentAddr), [][]byte{jsoncdc.MustEncode(cadence.Address(userAddress))})
		require.NoError(t, err)
		if !assert.True(t, result.Succeeded()) {
			t.Log(result.Error.Error())
		}
		balance = result.Value
		assert.Equal(t, balance.(cadence.UFix64), CadenceUFix64("300.0"))

		supply := executeScriptAndCheck(t, b, momentGenerateGetSupplyScript(fungibleAddr, momentAddr), nil)
		assert.Equal(t, supply.(cadence.UFix64), CadenceUFix64("1000.0"))
	})
}

func momentReplaceAddressPlaceholders(code string, fungibleAddress, momentAddress string) []byte {
	return []byte(replaceStrings(
		code,
		map[string]string{
			ftAddressPlaceholder:     "0x" + fungibleAddress,
			momentAddressPlaceHolder: "0x" + momentAddress,
		},
	))
}

func loadFungibleToken() []byte {
	return ft_contracts.FungibleToken()
}

func loadMoment(fungibleAddr flow.Address) []byte {
	return []byte(strings.ReplaceAll(
		string(readFile(momentMomentPath)),
		ftAddressPlaceholder,
		"0x"+fungibleAddr.String(),
	))
}

func momentGenerateGetSupplyScript(fungibleAddr, momentAddr flow.Address) []byte {
	return momentReplaceAddressPlaceholders(
		string(readFile(momentGetSupplyPath)),
		fungibleAddr.String(),
		momentAddr.String(),
	)
}

func momentGenerateGetBalanceScript(fungibleAddr, momentAddr flow.Address) []byte {
	return momentReplaceAddressPlaceholders(
		string(readFile(momentGetBalancePath)),
		fungibleAddr.String(),
		momentAddr.String(),
	)
}
func momentGenerateTransferVaultScript(fungibleAddr, momentAddr flow.Address) []byte {
	return momentReplaceAddressPlaceholders(
		string(readFile(momentTransferTokensPath)),
		fungibleAddr.String(),
		momentAddr.String(),
	)
}

func momentGenerateSetupMomentAccountTransaction(fungibleAddr, momentAddr flow.Address) []byte {
	return momentReplaceAddressPlaceholders(
		string(readFile(momentSetupAccountPath)),
		fungibleAddr.String(),
		momentAddr.String(),
	)
}

func momentGenerateMintMomentTransaction(fungibleAddr, momentAddr flow.Address) []byte {
	return momentReplaceAddressPlaceholders(
		string(readFile(momentMintTokensPath)),
		fungibleAddr.String(),
		momentAddr.String(),
	)
}
