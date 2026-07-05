const BASE_URL = import.meta.env.VITE_API_BASE_URL ?? "http://localhost:8080"

// Product image_url values are stored as relative paths (e.g. "/files/images/x.png");
// resolve them against the API origin so they load correctly regardless of which origin
// this app is served from.
export function resolveImageUrl(path?: string | null): string | undefined {
  if (!path) return undefined
  return new URL(path, BASE_URL).toString()
}

export function getToken(): string | null {
  return localStorage.getItem("admin_access_token")
}

export function setToken(token: string): void {
  localStorage.setItem("admin_access_token", token)
}

export function getRefreshToken(): string | null {
  return localStorage.getItem("admin_refresh_token")
}

export function setRefreshToken(token: string): void {
  localStorage.setItem("admin_refresh_token", token)
}

export function clearToken(): void {
  localStorage.removeItem("admin_access_token")
  localStorage.removeItem("admin_refresh_token")
  localStorage.removeItem("admin_user")
}

interface TokenPair {
  access_token: string
  refresh_token: string
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

  const res = await fetch(new URL("/api/v1/auth/refresh", BASE_URL).toString(), {
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

type QueryValue =
  | string
  | number
  | boolean
  | null
  | undefined
  | (string | number | boolean)[]

export async function apiFetch<T>(path: string, init?: RequestInit, params?: Record<string, QueryValue>,): Promise<T> {
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
    const headers: HeadersInit = {
      ...(init?.body instanceof FormData ? {} : { "Content-Type": "application/json" }),
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...(init?.headers ?? {}),
    }
    return fetch(url.toString(), { ...init, headers })
  }

  let res = await send()

  if (res.status === 401 && !isAuthEndpoint) {
    try {
      await refreshAccessToken()
      res = await send()
    } catch {
      clearToken()
      window.location.href = "/login"
    }
  }

  if (!res.ok) {
    const body = await res.text()
    throw new Error(body || res.statusText)
  }
  if (res.status === 204) return undefined as T
  return res.json() as Promise<T>
}
