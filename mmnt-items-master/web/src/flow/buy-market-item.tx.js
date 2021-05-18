import * as fcl from "@onflow/fcl"
import * as t from "@onflow/types"
import {tx} from "./util/tx"
import {invariant} from "@onflow/util-invariant"

const CODE = fcl.cdc`
  import FungibleToken from 0xFungibleToken
  import NonFungibleToken from 0xNonFungibleToken
  import Moment from 0x9ff86e7c9a7b1fd9
  import MmntItems from 0x9ff86e7c9a7b1fd9
  import MmntItemsMarket from 0x9ff86e7c9a7b1fd9

  transaction(itemID: UInt64, marketCollectionAddress: Address) {
      let paymentVault: @FungibleToken.Vault
      let mmntItemsCollection: &MmntItems.Collection{NonFungibleToken.Receiver}
      let marketCollection: &MmntItemsMarket.Collection{MmntItemsMarket.CollectionPublic}

      prepare(acct: AuthAccount) {
          self.marketCollection = getAccount(marketCollectionAddress)
              .getCapability<&MmntItemsMarket.Collection{MmntItemsMarket.CollectionPublic}>(MmntItemsMarket.CollectionPublicPath)
              .borrow() ?? panic("Could not borrow market collection from market address")

          let price = self.marketCollection.borrowSaleItem(itemID: itemID)!.price

          let mainMomentVault = acct.borrow<&Moment.Vault>(from: Moment.VaultStoragePath)
              ?? panic("Cannot borrow Moment vault from acct storage")
          self.paymentVault <- mainMomentVault.withdraw(amount: price)

          self.mmntItemsCollection = acct.borrow<&MmntItems.Collection{NonFungibleToken.Receiver}>(
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
`

// prettier-ignore
export function buyMarketItem({itemID, ownerAddress}, opts = {}) {
  invariant(itemID != null, "buyMarketItem({itemID, ownerAddress}) -- itemID required")
  invariant(ownerAddress != null, "buyMarketItem({itemID, ownerAddress}) -- ownerAddress required")

  return tx([
    fcl.transaction(CODE),
    fcl.args([
      fcl.arg(Number(itemID), t.UInt64),
      fcl.arg(String(ownerAddress), t.Address),
    ]),
    fcl.proposer(fcl.authz),
    fcl.payer(fcl.authz),
    fcl.authorizations([fcl.authz]),
    fcl.limit(1000),
  ], opts)
}
