import * as fcl from "@onflow/fcl"
import * as t from "@onflow/types"
import {tx} from "./util/tx"
import {invariant} from "@onflow/util-invariant"

const CODE = fcl.cdc`
  import MmntItemsMarket from 0x9ff86e7c9a7b1fd9

  transaction(itemID: UInt64) {
      prepare(account: AuthAccount) {
          let listing <- account
            .borrow<&MmntItemsMarket.Collection>(from: MmntItemsMarket.CollectionStoragePath)!
            .remove(itemID: itemID)
          destroy listing
      }
  }
`

// prettier-ignore
export function cancelMarketListing({ itemID }, opts = {}) {
  invariant(itemID != null, "cancelMarketListing({itemID}) -- itemID required")

  return tx([
    fcl.transaction(CODE),
    fcl.args([
      fcl.arg(Number(itemID), t.UInt64),
    ]),
    fcl.proposer(fcl.authz),
    fcl.payer(fcl.authz),
    fcl.authorizations([fcl.authz]),
    fcl.limit(1000),
  ], opts)
}
