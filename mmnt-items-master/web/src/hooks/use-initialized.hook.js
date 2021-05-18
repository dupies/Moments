import {
  atomFamily,
  selectorFamily,
  useRecoilState,
  useRecoilValue,
} from "recoil"
import {
  LOADING,
  IDLE,
  PROCESSING,
  SUCCESS,
  ERROR,
  IDLE_DELAY,
} from "../global/constants"
import {isAccountInitialized} from "../flow/is-account-initialized.script"
import {initializeAccount} from "../flow/initialize-account.tx"
import {sleep} from "../util/sleep"
import {useFlowBalance} from "./use-flow-balance.hook"
import {useMomentsBalance} from "./use-moments-balance.hook"

export const $status = atomFamily({
  key: "init::status",
  default: IDLE,
})

export const $init = atomFamily({
  key: "init::state",
  default: selectorFamily({
    key: "init::default",
    get: address => () => isAccountInitialized(address),
  }),
})

export const $computedInit = selectorFamily({
  key: "init::computed",
  get: address => async ({get}) => {
    const all = get($init(address))
    return all.Mmnt && all.MmntItems && all.MmntItemsMarket
  },
})

export function useInitialized(address) {
  const [init, setInit] = useRecoilState($init(address))
  const isInitialized = useRecoilValue($computedInit(address))
  const [status, setStatus] = useRecoilState($status(address))
  const flow = useFlowBalance(address)
  const mmnt = useMomentsBalance(address)

  function recheck() {
    isAccountInitialized(address).then(setInit)
  }

  return {
    ...init,
    isInitialized,
    status: isInitialized == null ? LOADING : status,
    recheck,
    async initialize() {
      initializeAccount(address, {
        onStart() {
          setStatus(PROCESSING)
        },
        async onSuccess() {
          recheck()
          flow.refresh()
          mmnt.refresh()
          setStatus(SUCCESS)
        },
        onError() {
          setStatus(ERROR)
        },
        async onComplete() {
          await sleep(IDLE_DELAY)
          setStatus(IDLE)
        },
      })
    },
  }
}
