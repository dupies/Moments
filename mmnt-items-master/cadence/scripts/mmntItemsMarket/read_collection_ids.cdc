import MmntItemsMarket from "../../contracts/MmntItemsMarket.cdc"

// This script returns an array of all the NFT IDs for sale 
// in an account's SaleOffer collection.

pub fun main(address: Address): [UInt64] {
    let marketCollectionRef = getAccount(address)
        .getCapability<&MmntItemsMarket.Collection{MmntItemsMarket.CollectionPublic}>(
            MmntItemsMarket.CollectionPublicPath
        )
        .borrow()
        ?? panic("Could not borrow market collection from market address")
    
    return marketCollectionRef.getSaleOfferIDs()
}
