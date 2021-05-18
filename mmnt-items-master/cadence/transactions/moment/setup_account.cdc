import FungibleToken from "../../contracts/FungibleToken.cdc"
import Moment from "../../contracts/Moment.cdc"

// This transaction is a template for a transaction
// to add a Vault resource to their account
// so that they can use the Moment

transaction {

    prepare(signer: AuthAccount) {

        if signer.borrow<&Moment.Vault>(from: Moment.VaultStoragePath) == nil {
            // Create a new Moment Vault and put it in storage
            signer.save(<-Moment.createEmptyVault(), to: Moment.VaultStoragePath)

            // Create a public capability to the Vault that only exposes
            // the deposit function through the Receiver interface
            signer.link<&Moment.Vault{FungibleToken.Receiver}>(
                Moment.ReceiverPublicPath,
                target: Moment.VaultStoragePath
            )

            // Create a public capability to the Vault that only exposes
            // the balance field through the Balance interface
            signer.link<&Moment.Vault{FungibleToken.Balance}>(
                Moment.BalancePublicPath,
                target: Moment.VaultStoragePath
            )
        }
    }
}
