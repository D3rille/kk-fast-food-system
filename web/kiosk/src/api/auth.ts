import { apiFetch, setToken, setRefreshToken } from "./client"

interface LoginResponse {
  access_token: string
  refresh_token: string
}

export async function kioskLogin(): Promise<string> {
  const username = import.meta.env.VITE_KIOSK_USERNAME as string | undefined
  const password = import.meta.env.VITE_KIOSK_PASSWORD as string | undefined

  if (!username || !password) {
    throw new Error("VITE_KIOSK_USERNAME and VITE_KIOSK_PASSWORD must be set")
  }

  const data = await apiFetch<LoginResponse>("/api/v1/auth/login", {
    method: "POST",
    body: JSON.stringify({ username, password }),
  })

  setToken(data.access_token)
  setRefreshToken(data.refresh_token)
  return data.access_token
}
