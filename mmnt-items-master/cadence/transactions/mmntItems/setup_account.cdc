import NonFungibleToken from "../../contracts/NonFungibleToken.cdc"
import MmntItems from "../../contracts/MmntItems.cdc"

// This transaction configures an account to hold Mmnt Items.

transaction {
    prepare(signer: AuthAccount) {
        // if the account doesn't already have a collection
        if signer.borrow<&MmntItems.Collection>(from: MmntItems.CollectionStoragePath) == nil {

            // create a new empty collection
            let collection <- MmntItems.createEmptyCollection()
            
            // save it to the account
            signer.save(<-collection, to: MmntItems.CollectionStoragePath)

            // create a public capability for the collection
            signer.link<&MmntItems.Collection{NonFungibleToken.CollectionPublic, MmntItems.MmntItemsCollectionPublic}>(MmntItems.CollectionPublicPath, target: MmntItems.CollectionStoragePath)
        }
    }
}
