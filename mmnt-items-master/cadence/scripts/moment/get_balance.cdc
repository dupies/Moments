import Moment from "../../contracts/Moment.cdc"
import FungibleToken from "../../contracts/FungibleToken.cdc"

// This script returns an account's Moment balance.

pub fun main(address: Address): UFix64 {
    let account = getAccount(address)
    
    let vaultRef = account.getCapability(Moment.BalancePublicPath)!.borrow<&Moment.Vault{FungibleToken.Balance}>()
        ?? panic("Could not borrow Balance reference to the Vault")

    return vaultRef.balance
}
