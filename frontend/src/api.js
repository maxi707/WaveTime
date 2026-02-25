import { getToken } from './lib/session'

const API_BASE = '/api'

function buildHeaders(extraHeaders = {}, needsAuth = false) {
  const headers = {
    'Content-Type': 'application/json',
    ...extraHeaders,
  }

  if (needsAuth) {
    const token = getToken()
    if (token) {
      headers.Authorization = `Bearer ${token}`
    }
  }

  return headers
}

async function request(path, options = {}, needsAuth = false) {
  const response = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers: buildHeaders(options.headers, needsAuth),
  })

  if (response.status === 204) {
    return null
  }

  const payload = await response.json().catch(() => ({}))

  if (!response.ok) {
    const message = payload?.error || `Request failed: ${response.status}`
    throw new Error(message)
  }

  return payload
}

export async function healthz() {
  return request('/healthz')
}

export async function registerUser(data) {
  return request('/auth/register', {
    method: 'POST',
    body: JSON.stringify(data),
  })
}

export async function loginUser(data) {
  return request('/auth/login', {
    method: 'POST',
    body: JSON.stringify(data),
  })
}

export async function getMe() {
  return request('/me', { method: 'GET' }, true)
}

export async function patchMe(data) {
  return request('/me', {
    method: 'PATCH',
    body: JSON.stringify(data),
  }, true)
}

export async function listPools() {
  return request('/pools', { method: 'GET' }, true)
}

export async function listTrainingTypes() {
  return request('/training-types', { method: 'GET' }, true)
}

export async function listSchedule(poolId, dateFrom, dateTo) {
  const params = new URLSearchParams({
    pool_id: String(poolId),
    date_from: dateFrom,
    date_to: dateTo,
  })

  return request(`/schedule?${params.toString()}`, { method: 'GET' }, true)
}

export async function createBooking(slotId) {
  return request('/bookings', {
    method: 'POST',
    body: JSON.stringify({ slot_id: slotId }),
  }, true)
}

export async function listMyBookings() {
  return request('/bookings/my', { method: 'GET' }, true)
}

export async function cancelBooking(id) {
  return request(`/bookings/${id}`, { method: 'DELETE' }, true)
}

export async function adjustBookingSeats(id, action) {
  return request(`/bookings/${id}/seats`, {
    method: 'PATCH',
    body: JSON.stringify({ action }),
  }, true)
}

export async function initPayment(bookingId) {
  return request('/payments/init', {
    method: 'POST',
    body: JSON.stringify({ booking_id: bookingId, provider: 'mock' }),
  }, true)
}

export async function applyWebhook(externalId, status) {
  return request('/payments/webhook', {
    method: 'POST',
    body: JSON.stringify({ external_id: externalId, status }),
  })
}

export async function listAdminBookings(status = '', limit = 100) {
  const params = new URLSearchParams()
  if (status) params.set('status', status)
  params.set('limit', String(limit))
  return request(`/admin/bookings?${params.toString()}`, { method: 'GET' }, true)
}

export async function updateAdminBooking(id, status, reason = '') {
  return request(`/admin/bookings/${id}`, {
    method: 'PATCH',
    body: JSON.stringify({ status, reason }),
  }, true)
}

export async function createAdminSlot(data) {
  return request('/admin/slots', {
    method: 'POST',
    body: JSON.stringify(data),
  }, true)
}

export async function updateAdminSlot(id, data) {
  return request(`/admin/slots/${id}`, {
    method: 'PATCH',
    body: JSON.stringify(data),
  }, true)
}
