import MmntItemsMarket from "../../contracts/MmntItemsMarket.cdc"

// This transaction configures an account to hold SaleOffer items.

transaction {
    prepare(signer: AuthAccount) {

        // if the account doesn't already have a collection
        if signer.borrow<&MmntItemsMarket.Collection>(from: MmntItemsMarket.CollectionStoragePath) == nil {

            // create a new empty collection
            let collection <- MmntItemsMarket.createEmptyCollection() as! @MmntItemsMarket.Collection
            
            // save it to the account
            signer.save(<-collection, to: MmntItemsMarket.CollectionStoragePath)

            // create a public capability for the collection
            signer.link<&MmntItemsMarket.Collection{MmntItemsMarket.CollectionPublic}>(MmntItemsMarket.CollectionPublicPath, target: MmntItemsMarket.CollectionStoragePath)
        }
    }
}
