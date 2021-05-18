import MmntItemsMarket from "../../contracts/MmntItemsMarket.cdc"

transaction(itemID: UInt64) {
    let marketCollection: &MmntItemsMarket.Collection

    prepare(signer: AuthAccount) {
        self.marketCollection = signer.borrow<&MmntItemsMarket.Collection>(from: MmntItemsMarket.CollectionStoragePath)
            ?? panic("Missing or mis-typed MmntItemsMarket Collection")
    }

    execute {
        let offer <-self.marketCollection.remove(itemID: itemID)
        destroy offer
    }
}
