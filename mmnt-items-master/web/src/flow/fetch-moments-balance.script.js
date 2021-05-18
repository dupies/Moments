import {send, decode, script, args, arg, cdc} from "@onflow/fcl"
import {Address} from "@onflow/types"

const CODE = cdc`
  import FungibleToken from 0xFungibleToken
  import Moment from 0x9ff86e7c9a7b1fd9

  pub fun main(address: Address): UFix64? {
    if let vault = getAccount(address).getCapability<&{FungibleToken.Balance}>(Moment.BalancePublicPath).borrow() {
      return vault.balance
    }
    return nil
  }

`

export function fetchMomentsBalance(address) {
  if (address == null) return Promise.resolve(false)

  // prettier-ignore
  return send([
    script(CODE),
    args([
      arg(address, Address)
    ])
  ]).then(decode)
}
