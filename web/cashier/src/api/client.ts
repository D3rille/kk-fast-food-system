const BASE_URL = (import.meta.env.VITE_API_BASE_URL as string | undefined) ?? "http://localhost:8080"

const ACCESS_TOKEN_KEY = "cashier_token"
const REFRESH_TOKEN_KEY = "cashier_refresh_token"

export function getToken(): string | null {
  return localStorage.getItem(ACCESS_TOKEN_KEY)
}

export function setToken(t: string | null) {
  if (t) localStorage.setItem(ACCESS_TOKEN_KEY, t)
  else localStorage.removeItem(ACCESS_TOKEN_KEY)
}

export function getRefreshToken(): string | null {
  return localStorage.getItem(REFRESH_TOKEN_KEY)
}

export function setRefreshToken(t: string | null) {
  if (t) localStorage.setItem(REFRESH_TOKEN_KEY, t)
  else localStorage.removeItem(REFRESH_TOKEN_KEY)
}

export function clearTokens() {
  setToken(null)
  setRefreshToken(null)
}

export function getWsUrl(): string {
  const base = (import.meta.env.VITE_WS_BASE_URL as string | undefined) ?? BASE_URL
  return base.replace(/^http/, "ws") + "/ws/kitchen"
}

interface TokenPair {
  access_token: string
  refresh_token: string
}

let authFailureHandler: (() => void) | null = null

// Registered by AuthProvider so a rejected refresh token can send the
// cashier back to the login screen.
export function setAuthFailureHandler(handler: () => void) {
  authFailureHandler = handler
}

let refreshPromise: Promise<string> | null = null

async function refreshAccessToken(): Promise<string> {
  if (!refreshPromise) {
    refreshPromise = doRefresh().finally(() => {
      refreshPromise = null
    })
  }
  return refreshPromise
}

async function doRefresh(): Promise<string> {
  const refreshToken = getRefreshToken()
  if (!refreshToken) {
    throw new Error("No refresh token available")
  }

  const res = await fetch(`${BASE_URL}/api/v1/auth/refresh`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ refresh_token: refreshToken }),
  })
  if (!res.ok) {
    throw new Error("Refresh token rejected")
  }
  const data = (await res.json()) as TokenPair
  setToken(data.access_token)
  setRefreshToken(data.refresh_token)
  return data.access_token
}

export async function apiFetch<T>(path: string, init?: RequestInit): Promise<T> {
  const isAuthEndpoint = path.startsWith("/api/v1/auth/")

  const send = async (): Promise<Response> => {
    const headers: Record<string, string> = {
      "Content-Type": "application/json",
      ...(init?.headers as Record<string, string>),
    }
    const token = getToken()
    if (token) {
      headers["Authorization"] = `Bearer ${token}`
    }
    return fetch(`${BASE_URL}${path}`, { ...init, headers })
  }

  let res = await send()

  if (res.status === 401 && !isAuthEndpoint) {
    try {
      await refreshAccessToken()
      res = await send()
    } catch {
      clearTokens()
      authFailureHandler?.()
    }
  }

  if (!res.ok) {
    const body = await res.text()
    throw new Error(`API error ${res.status}: ${body}`)
  }
  return res.json() as Promise<T>
}
