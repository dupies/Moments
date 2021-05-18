import {atomFamily, selectorFamily, useRecoilState} from "recoil"
import {fetchMomentsBalance} from "../flow/fetch-moments-balance.script"
import {IDLE, PROCESSING} from "../global/constants"

export const valueAtom = atomFamily({
  key: "moments-balance::state",
  default: selectorFamily({
    key: "moments-balance::default",
    get: address => async () => fetchMomentsBalance(address),
  }),
})

export const statusAtom = atomFamily({
  key: "moments-balance::status",
  default: IDLE,
})

export function useMomentsBalance(address) {
  const [balance, setBalance] = useRecoilState(valueAtom(address))
  const [status, setStatus] = useRecoilState(statusAtom(address))

  async function refresh() {
    setStatus(PROCESSING)
    await fetchMomentsBalance(address).then(setBalance)
    setStatus(IDLE)
  }

  return {
    balance,
    status,
    refresh,
    async mint() {
      setStatus(PROCESSING)
      await fetch(process.env.REACT_APP_API_MOMENT_MINT, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          recipient: address,
          amount: 50.0,
        }),
      })
      await fetchMomentsBalance(address).then(setBalance)
      setStatus(IDLE)
    },
  }
}
