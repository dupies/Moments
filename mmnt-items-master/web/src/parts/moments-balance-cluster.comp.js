import {Suspense} from "react"
import {useInitialized} from "../hooks/use-initialized.hook"
import {useMomentsBalance} from "../hooks/use-moments-balance.hook"
import {Bar, Label, Button} from "../display/bar.comp"
import {IDLE} from "../global/constants"
import {Loading} from "../parts/loading.comp"
import {fmtMoments} from "../util/fmt-moments"

export function MomentsBalanceCluster({address}) {
  const init = useInitialized(address)
  const moment = useMomentsBalance(address)
  if (address == null || !init.isInitialized) return null

  return (
    <Bar>
      <Label>Moments Balance:</Label>
      <Label strong good={moment.balance > 0} bad={moment.balance <= 0}>
        {fmtMoments(moment.balance)}
      </Label>
      <Button disabled={moment.status !== IDLE} onClick={moment.refresh}>
        Refresh
      </Button>
      <Button disabled={moment.status !== IDLE} onClick={moment.mint}>
        Mint
      </Button>
      {moment.status !== IDLE && <Loading label={moment.status} />}
    </Bar>
  )
}

export default function WrappedMomentsBalanceCluster({address}) {
  return (
    <Suspense
      fallback={
        <Bar>
          <Loading label="Fetching Moments Balance" />
        </Bar>
      }
    >
      <MomentsBalanceCluster address={address} />
    </Suspense>
  )
}
