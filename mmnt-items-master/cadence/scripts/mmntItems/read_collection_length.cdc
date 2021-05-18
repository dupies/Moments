import NonFungibleToken from "../../contracts/NonFungibleToken.cdc"
import MmntItems from "../../contracts/MmntItems.cdc"

// This script returns the size of an account's MmntItems collection.

pub fun main(address: Address): Int {
    let account = getAccount(address)

    let collectionRef = account.getCapability(MmntItems.CollectionPublicPath)!
        .borrow<&{NonFungibleToken.CollectionPublic}>()
        ?? panic("Could not borrow capability from public collection")
    
    return collectionRef.getIDs().length
}
