export function fmtMoments(balance, cur = false) {
  if (balance == null) return null
  return [
    String(balance).replace(/0+$/, "").replace(/\.$/, ""),
    cur && "MOMENT",
  ]
    .filter(Boolean)
    .join(" ")
}
