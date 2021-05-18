import {Bar, Label} from "../display/bar.comp"
import {useConfig} from "../hooks/use-config.hook"

const Link = ({address, name}) => {
  const env = useConfig("env")

  return (
    <li>
      {name}:{" "}
      <a href={fvs(env, address, name)} target="_blank" rel="noreferrer">
        {address}
      </a>
    </li>
  )
}

export function ContractsCluster() {
  const moment = useConfig("0xMoment")
  const items = useConfig("0xMmntItems")
  const market = useConfig("0xMmntItemsMarket")

  return (
    <div>
      <Bar>
        <Label>Contracts</Label>
      </Bar>
      <ul>
        <Link address={moment} name="Moment" />
        <Link address={items} name="MmntItems" />
        <Link address={market} name="MmntItemsMarket" />
      </ul>
    </div>
  )
}

export default function WrappedContractsCluster() {
  return <ContractsCluster />
}

function fvs(env, addr, contract) {
  return `https://flow-view-source.com/${env}/account/${addr}/contract/${contract}`
}
