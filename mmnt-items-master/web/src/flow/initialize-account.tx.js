// prettier-ignore
import {transaction, limit, proposer, payer, authorizations, authz, cdc} from "@onflow/fcl"
import {invariant} from "@onflow/util-invariant"
import {tx} from "./util/tx"

const CODE = cdc`
  import FungibleToken from 0xFungibleToken
  import NonFungibleToken from 0xNonFungibleToken
  import Moment from 0x9ff86e7c9a7b1fd9
  import MmntItems from 0x9ff86e7c9a7b1fd9
  import MmntItemsMarket from 0x9ff86e7c9a7b1fd9

  pub fun hasMoment(_ address: Address): Bool {
    let receiver = getAccount(address)
      .getCapability<&Moment.Vault{FungibleToken.Receiver}>(Moment.ReceiverPublicPath)
      .check()

    let balance = getAccount(address)
      .getCapability<&Moment.Vault{FungibleToken.Balance}>(Moment.BalancePublicPath)
      .check()

    return receiver && balance
  }

  pub fun hasItems(_ address: Address): Bool {
    return getAccount(address)
      .getCapability<&MmntItems.Collection{NonFungibleToken.CollectionPublic, MmntItems.MmntItemsCollectionPublic}>(MmntItems.CollectionPublicPath)
      .check()
  }

  pub fun hasMarket(_ address: Address): Bool {
    return getAccount(address)
      .getCapability<&MmntItemsMarket.Collection{MmntItemsMarket.CollectionPublic}>(MmntItemsMarket.CollectionPublicPath)
      .check()
  }

  transaction {
    prepare(acct: AuthAccount) {
      if !hasMoment(acct.address) {
        if acct.borrow<&Moment.Vault>(from: Moment.VaultStoragePath) == nil {
          acct.save(<-Moment.createEmptyVault(), to: Moment.VaultStoragePath)
        }
        acct.unlink(Moment.ReceiverPublicPath)
        acct.unlink(Moment.BalancePublicPath)
        acct.link<&Moment.Vault{FungibleToken.Receiver}>(Moment.ReceiverPublicPath, target: Moment.VaultStoragePath)
        acct.link<&Moment.Vault{FungibleToken.Balance}>(Moment.BalancePublicPath, target: Moment.VaultStoragePath)
      }

      if !hasItems(acct.address) {
        if acct.borrow<&MmntItems.Collection>(from: MmntItems.CollectionStoragePath) == nil {
          acct.save(<-MmntItems.createEmptyCollection(), to: MmntItems.CollectionStoragePath)
        }
        acct.unlink(MmntItems.CollectionPublicPath)
        acct.link<&MmntItems.Collection{NonFungibleToken.CollectionPublic, MmntItems.MmntItemsCollectionPublic}>(MmntItems.CollectionPublicPath, target: MmntItems.CollectionStoragePath)
      }

      if !hasMarket(acct.address) {
        if acct.borrow<&MmntItemsMarket.Collection>(from: MmntItemsMarket.CollectionStoragePath) == nil {
          acct.save(<-MmntItemsMarket.createEmptyCollection(), to: MmntItemsMarket.CollectionStoragePath)
        }
        acct.unlink(MmntItemsMarket.CollectionPublicPath)
        acct.link<&MmntItemsMarket.Collection{MmntItemsMarket.CollectionPublic}>(MmntItemsMarket.CollectionPublicPath, target:MmntItemsMarket.CollectionStoragePath)
      }
    }
  }
`

export async function initializeAccount(address, opts = {}) {
  // prettier-ignore
  invariant(address != null, "Tried to initialize an account but no address was supplied")

  return tx(
    [
      transaction(CODE),
      limit(70),
      proposer(authz),
      payer(authz),
      authorizations([authz]),
    ],
    opts
  )
}
