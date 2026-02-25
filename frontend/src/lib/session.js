export function getToken() {
  return localStorage.getItem('wavetime_token') || ''
}

export function setToken(token) {
  localStorage.setItem('wavetime_token', token)
}

export function clearToken() {
  localStorage.removeItem('wavetime_token')
}

export function isAuthed() {
  return Boolean(getToken())
}

export function getTokenClaims() {
  const token = getToken()
  if (!token) return null

  const payloadPart = token.split('.')[0]
  if (!payloadPart) return null

  try {
    const normalized = payloadPart.replace(/-/g, '+').replace(/_/g, '/')
    const padded = normalized + '='.repeat((4 - (normalized.length % 4)) % 4)
    const json = atob(padded)
    return JSON.parse(json)
  } catch {
    return null
  }
}

export function getUserRole() {
  const claims = getTokenClaims()
  return claims?.role || ''
}

export function isAdmin() {
  return getUserRole() === 'admin'
}
