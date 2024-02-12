export function compareUInt64Str(a: string, b: string) {
  const lenDiff = a.length - b.length
  if (lenDiff !== 0) return lenDiff
  if (a < b) return -1
  if (a > b) return 1
  return 0
}
