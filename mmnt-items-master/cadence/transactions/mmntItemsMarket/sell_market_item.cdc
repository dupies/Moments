import FungibleToken from "../../contracts/FungibleToken.cdc"
import NonFungibleToken from "../../contracts/NonFungibleToken.cdc"
import Moment from "../../contracts/Moment.cdc"
import MmntItems from "../../contracts/MmntItems.cdc"
import MmntItemsMarket from "../../contracts/MmntItemsMarket.cdc"

transaction(itemID: UInt64, price: UFix64) {
    let momentVault: Capability<&Moment.Vault{FungibleToken.Receiver}>
    let mmntItemsCollection: Capability<&MmntItems.Collection{NonFungibleToken.Provider, MmntItems.MmntItemsCollectionPublic}>
    let marketCollection: &MmntItemsMarket.Collection

    prepare(signer: AuthAccount) {
        // we need a provider capability, but one is not provided by default so we create one.
        let MmntItemsCollectionProviderPrivatePath = /private/mmntItemsCollectionProvider

        self.momentVault = signer.getCapability<&Moment.Vault{FungibleToken.Receiver}>(Moment.ReceiverPublicPath)!
        assert(self.momentVault.borrow() != nil, message: "Missing or mis-typed Moment receiver")

        if !signer.getCapability<&MmntItems.Collection{NonFungibleToken.Provider, MmntItems.MmntItemsCollectionPublic}>(MmntItemsCollectionProviderPrivatePath)!.check() {
            signer.link<&MmntItems.Collection{NonFungibleToken.Provider, MmntItems.MmntItemsCollectionPublic}>(MmntItemsCollectionProviderPrivatePath, target: MmntItems.CollectionStoragePath)
        }

        self.mmntItemsCollection = signer.getCapability<&MmntItems.Collection{NonFungibleToken.Provider, MmntItems.MmntItemsCollectionPublic}>(MmntItemsCollectionProviderPrivatePath)!
        assert(self.mmntItemsCollection.borrow() != nil, message: "Missing or mis-typed MmntItemsCollection provider")

        self.marketCollection = signer.borrow<&MmntItemsMarket.Collection>(from: MmntItemsMarket.CollectionStoragePath)
            ?? panic("Missing or mis-typed MmntItemsMarket Collection")
    }

    execute {
        let offer <- MmntItemsMarket.createSaleOffer (
            sellerItemProvider: self.mmntItemsCollection,
            itemID: itemID,
            typeID: self.mmntItemsCollection.borrow()!.borrowMmntItem(id: itemID)!.typeID,
            sellerPaymentReceiver: self.momentVault,
            price: price
        )
        self.marketCollection.insert(offer: <-offer)
    }
}
