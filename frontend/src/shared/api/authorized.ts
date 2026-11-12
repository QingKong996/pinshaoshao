export function authorized(token: string) {
  return {
    headers: { Authorization: `Bearer ${token}` },
  }
}
