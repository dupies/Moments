import NonFungibleToken from "../../contracts/NonFungibleToken.cdc"
import MmntItems from "../../contracts/MmntItems.cdc"

// This script returns the metadata for an NFT in an account's collection.

pub fun main(address: Address, itemID: UInt64): UInt64 {

    // get the public account object for the token owner
    let owner = getAccount(address)

    let collectionBorrow = owner.getCapability(MmntItems.CollectionPublicPath)!
        .borrow<&{MmntItems.MmntItemsCollectionPublic}>()
        ?? panic("Could not borrow MmntItemsCollectionPublic")

    // borrow a reference to a specific NFT in the collection
    let mmntItem = collectionBorrow.borrowMmntItem(id: itemID)
        ?? panic("No such itemID in that collection")

    return mmntItem.typeID
}
