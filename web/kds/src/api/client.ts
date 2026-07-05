const BASE_URL = import.meta.env.VITE_API_BASE_URL ?? "http://localhost:8080"

let token: string | null = null
let refreshToken: string | null = null

export function setToken(t: string) {
  token = t
}

export function setRefreshToken(t: string) {
  refreshToken = t
}

export function clearToken() {
  token = null
  refreshToken = null
}

interface TokenPair {
  access_token: string
  refresh_token: string
}

let refreshPromise: Promise<string> | null = null

// The KDS has no login page/staff user to send back to — it authenticates
// as a device. If the refresh token is missing or rejected, fall back to a
// fresh device login using the configured kitchen credentials.
async function refreshAccessToken(): Promise<string> {
  if (!refreshPromise) {
    refreshPromise = doRefresh().finally(() => {
      refreshPromise = null
    })
  }
  return refreshPromise
}

async function doRefresh(): Promise<string> {
  if (refreshToken) {
    const res = await fetch(`${BASE_URL}/api/v1/auth/refresh`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ refresh_token: refreshToken }),
    })
    if (res.ok) {
      const data = (await res.json()) as TokenPair
      setToken(data.access_token)
      setRefreshToken(data.refresh_token)
      return data.access_token
    }
  }

  const username = import.meta.env.VITE_KDS_USERNAME as string | undefined
  const password = import.meta.env.VITE_KDS_PASSWORD as string | undefined
  if (!username || !password) {
    clearToken()
    throw new Error("Session expired and no device credentials configured")
  }

  const loginRes = await fetch(`${BASE_URL}/api/v1/auth/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ username, password }),
  })
  if (!loginRes.ok) {
    clearToken()
    throw new Error("Failed to re-authenticate kitchen display")
  }
  const data = (await loginRes.json()) as TokenPair
  setToken(data.access_token)
  setRefreshToken(data.refresh_token)
  return data.access_token
}

export async function apiFetch<T>(path: string, init?: RequestInit): Promise<T> {
  const isAuthEndpoint = path.startsWith("/api/v1/auth/")

  const send = async (): Promise<Response> => {
    const headers: Record<string, string> = {
      "Content-Type": "application/json",
    }
    if (token) {
      headers["Authorization"] = `Bearer ${token}`
    }

    return fetch(`${BASE_URL}${path}`, {
      ...init,
      headers: { ...headers, ...(init?.headers as Record<string, string>) },
    })
  }

  let res = await send()

  if (res.status === 401 && !isAuthEndpoint) {
    try {
      await refreshAccessToken()
      res = await send()
    } catch {
      // Fall through — the original 401 response is handled below.
    }
  }

  if (!res.ok) {
    const text = await res.text()
    throw new Error(`API error ${res.status}: ${text}`)
  }

  return res.json() as Promise<T>
}

export function getWsUrl(): string {
  const wsBase =
    import.meta.env.VITE_WS_BASE_URL ?? BASE_URL.replace(/^http/, "ws")
  return `${wsBase}/ws/kitchen`
}
