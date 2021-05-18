import * as fcl from "@onflow/fcl"
import * as t from "@onflow/types"
import {tx} from "./util/tx"

const CODE = fcl.cdc`
  import FungibleToken from 0xFungibleToken
  import NonFungibleToken from 0xNonFungibleToken
  import Moment from 0x9ff86e7c9a7b1fd9
  import MmntItems from 0x9ff86e7c9a7b1fd9
  import MmntItemsMarket from 0x9ff86e7c9a7b1fd9

  transaction(itemID: UInt64, price: UFix64) {
    let momentVault: Capability<&Moment.Vault{FungibleToken.Receiver}>
    let mmntItemsCollection: Capability<&MmntItems.Collection{NonFungibleToken.Provider, MmntItems.MmntItemsCollectionPublic}>
    let marketCollection: &MmntItemsMarket.Collection

    prepare(signer: AuthAccount) {
        // we need a provider capability, but one is not provided by default so we create one.
        let MmntItemsCollectionProviderPrivatePath = /private/mmntItemsCollectionProvider

        self.momentVault = signer.getCapability<&Moment.Vault{FungibleToken.Receiver}>(Moment.ReceiverPublicPath)!
        assert(self.momentVault.borrow() != nil, message: "Missing or mis-typed Moment receiver")

        if !signer.getCapability<&MmntItems.Collection{NonFungibleToken.Provider, MmntItems.MmntItemsCollectionPublic}>(MmntItemsCollectionProviderPrivatePath)!.check() {
            signer.link<&MmntItems.Collection{NonFungibleToken.Provider, MmntItems.MmntItemsCollectionPublic}>(MmntItemsCollectionProviderPrivatePath, target: MmntItems.CollectionStoragePath)
        }

        self.mmntItemsCollection = signer.getCapability<&MmntItems.Collection{NonFungibleToken.Provider, MmntItems.MmntItemsCollectionPublic}>(MmntItemsCollectionProviderPrivatePath)!
        assert(self.mmntItemsCollection.borrow() != nil, message: "Missing or mis-typed MmntItemsCollection provider")

        self.marketCollection = signer.borrow<&MmntItemsMarket.Collection>(from: MmntItemsMarket.CollectionStoragePath)
            ?? panic("Missing or mis-typed MmntItemsMarket Collection")
    }

    execute {
        let offer <- MmntItemsMarket.createSaleOffer (
            sellerItemProvider: self.mmntItemsCollection,
            itemID: itemID,
            typeID: self.mmntItemsCollection.borrow()!.borrowMmntItem(id: itemID)!.typeID,
            sellerPaymentReceiver: self.momentVault,
            price: price
        )
        self.marketCollection.insert(offer: <-offer)
    }
}
`

export function createSaleOffer({itemID, price}, opts = {}) {
  if (itemID == null)
    throw new Error("createSaleOffer(itemID, price) -- itemID required")
  if (price == null)
    throw new Error("createSaleOffer(itemID, price) -- price required")

  // prettier-ignore
  return tx([
    fcl.transaction(CODE),
    fcl.args([
      fcl.arg(Number(itemID), t.UInt64),
      fcl.arg(String(price), t.UFix64),
    ]),
    fcl.proposer(fcl.authz),
    fcl.payer(fcl.authz),
    fcl.authorizations([
      fcl.authz
    ]),
    fcl.limit(1000)
  ], opts)
}
