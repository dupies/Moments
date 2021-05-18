import {config} from "@onflow/fcl"

config()
  .put("env", process.env.REACT_APP_CHAIN_ENV)
  .put("accessNode.api", process.env.REACT_APP_FLOW_ACCESS_NODE)
  .put("challenge.handshake", process.env.REACT_APP_WALLET_DISCOVERY)
  .put("0xFungibleToken", process.env.REACT_APP_CONTRACT_FUNGIBLE_TOKEN)
  .put("0xNonFungibleToken", process.env.REACT_APP_CONTRACT_NON_FUNGIBLE_TOKEN)
  .put("0xMoment", process.env.REACT_APP_CONTRACT_MOMENT)
  .put("0xMomentItemsMarket", process.env.REACT_APP_CONTRACT_MMNT_ITEMS_MARKET)
  .put("0xMomentItems", process.env.REACT_APP_CONTRACT_MMNT_ITEMS)
