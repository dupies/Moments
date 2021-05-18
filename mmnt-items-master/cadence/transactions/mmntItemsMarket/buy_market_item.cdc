import FungibleToken from "../../contracts/FungibleToken.cdc"
import NonFungibleToken from "../../contracts/NonFungibleToken.cdc"
import Moment from "../../contracts/Moment.cdc"
import MmntItems from "../../contracts/MmntItems.cdc"
import MmntItemsMarket from "../../contracts/MmntItemsMarket.cdc"

transaction(itemID: UInt64, marketCollectionAddress: Address) {
    let paymentVault: @FungibleToken.Vault
    let mmntItemsCollection: &MmntItems.Collection{NonFungibleToken.Receiver}
    let marketCollection: &MmntItemsMarket.Collection{MmntItemsMarket.CollectionPublic}

    prepare(signer: AuthAccount) {
        self.marketCollection = getAccount(marketCollectionAddress)
            .getCapability<&MmntItemsMarket.Collection{MmntItemsMarket.CollectionPublic}>(
                MmntItemsMarket.CollectionPublicPath
            )!
            .borrow()
            ?? panic("Could not borrow market collection from market address")

        let saleItem = self.marketCollection.borrowSaleItem(itemID: itemID)
                    ?? panic("No item with that ID")
        let price = saleItem.price

        let mainMomentVault = signer.borrow<&Moment.Vault>(from: Moment.VaultStoragePath)
            ?? panic("Cannot borrow Moment vault from acct storage")
        self.paymentVault <- mainMomentVault.withdraw(amount: price)

        self.mmntItemsCollection = signer.borrow<&MmntItems.Collection{NonFungibleToken.Receiver}>(
            from: MmntItems.CollectionStoragePath
        ) ?? panic("Cannot borrow MmntItems collection receiver from acct")
    }

    execute {
        self.marketCollection.purchase(
            itemID: itemID,
            buyerCollection: self.mmntItemsCollection,
            buyerPayment: <- self.paymentVault
        )
    }
}
