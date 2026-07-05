const BASE_URL = import.meta.env.VITE_API_BASE_URL ?? "http://localhost:8080"

// Product image_url values are stored as relative paths (e.g. "/files/images/x.png");
// resolve them against the API origin so they load correctly regardless of which origin
// this app is served from.
export function resolveImageUrl(path?: string | null): string | undefined {
  if (!path) return undefined
  return new URL(path, BASE_URL).toString()
}

export function getToken(): string | null {
  return localStorage.getItem("kiosk_access_token")
}

export function setToken(token: string): void {
  localStorage.setItem("kiosk_access_token", token)
}

export function getRefreshToken(): string | null {
  return localStorage.getItem("kiosk_refresh_token")
}

export function setRefreshToken(token: string): void {
  localStorage.setItem("kiosk_refresh_token", token)
}

export function clearToken() {
  localStorage.removeItem("kiosk_access_token")
  localStorage.removeItem("kiosk_refresh_token")
}

interface TokenPair {
  access_token: string
  refresh_token: string
}

let refreshPromise: Promise<string> | null = null

// The kiosk has no login page/staff user to send back to — it authenticates
// as a device. If the refresh token is missing or rejected, fall back to a
// fresh device login using the configured kiosk credentials.
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
  if (refreshToken) {
    const res = await fetch(new URL("/api/v1/auth/refresh", BASE_URL).toString(), {
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

  const username = import.meta.env.VITE_KIOSK_USERNAME as string | undefined
  const password = import.meta.env.VITE_KIOSK_PASSWORD as string | undefined
  if (!username || !password) {
    clearToken()
    throw new Error("Session expired and no device credentials configured")
  }

  const loginRes = await fetch(new URL("/api/v1/auth/login", BASE_URL).toString(), {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ username, password }),
  })
  if (!loginRes.ok) {
    clearToken()
    throw new Error("Failed to re-authenticate kiosk device")
  }
  const data = (await loginRes.json()) as TokenPair
  setToken(data.access_token)
  setRefreshToken(data.refresh_token)
  return data.access_token
}

type QueryValue =
  | string
  | number
  | boolean
  | null
  | undefined
  | (string | number | boolean)[]

export async function apiFetch<T>(
  path: string,
  init?: RequestInit,
  params?: Record<string, QueryValue>,
): Promise<T> {
  const url = new URL(path, BASE_URL)

  if (params) {
    for (const [key, value] of Object.entries(params)) {
      if (value == null) continue

      if (Array.isArray(value)) {
        value.forEach((v) => url.searchParams.append(key, String(v)))
      } else {
        url.searchParams.append(key, String(value))
      }
    }
  }

  const isAuthEndpoint = path.startsWith("/api/v1/auth/")

  const send = async (): Promise<Response> => {
    const token = getToken()
    const headers: Record<string, string> = {
      "Content-Type": "application/json",
    }
    if (token) {
      headers["Authorization"] = `Bearer ${token}`
    }
    return fetch(url.toString(), {
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
