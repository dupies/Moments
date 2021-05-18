import {send, decode, script, args, arg, cdc} from "@onflow/fcl"
import {Address} from "@onflow/types"

const CODE = cdc`
  import FungibleToken from 0xFungibleToken
  import NonFungibleToken from 0xNonFungibleToken
  import Moment from 0x9ff86e7c9a7b1fd9
  import MmntItems from 0x9ff86e7c9a7b1fd9
  import MmntItemsMarket from 0x9ff86e7c9a7b1fd9

  pub fun hasMoment(_ address: Address): Bool {
    let receiver: Bool = getAccount(address)
      .getCapability<&Moment.Vault{FungibleToken.Receiver}>(Moment.ReceiverPublicPath)
      .check()

    let balance: Bool = getAccount(address)
      .getCapability<&Moment.Vault{FungibleToken.Balance}>(Moment.BalancePublicPath)
      .check()

    return receiver && balance
  }

  pub fun hasMmntItems(_ address: Address): Bool {
    return getAccount(address)
      .getCapability<&MmntItems.Collection{NonFungibleToken.CollectionPublic, MmntItems.MmntItemsCollectionPublic}>(MmntItems.CollectionPublicPath)
      .check()
  }

  pub fun hasMmntItemsMarket(_ address: Address): Bool {
    return getAccount(address)
      .getCapability<&MmntItemsMarket.Collection{MmntItemsMarket.CollectionPublic}>(MmntItemsMarket.CollectionPublicPath)
      .check()
  }

  pub fun main(address: Address): {String: Bool} {
    let ret: {String: Bool} = {}
    ret["Moment"] = hasMoment(address)
    ret["MmntItems"] = hasMmntItems(address)
    ret["MmntItemsMarket"] = hasMmntItemsMarket(address)
    return ret
  }
`

export function isAccountInitialized(address) {
  if (address == null) return Promise.resolve(false)

  // prettier-ignore
  return send([
    script(CODE),
    args([
      arg(address, Address)
    ])
  ]).then(decode)
}
