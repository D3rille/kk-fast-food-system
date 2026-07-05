import { apiFetch, setToken, setRefreshToken } from "./client"

interface LoginResponse {
  access_token: string
  refresh_token: string
}

export async function cashierLogin(username: string, password: string): Promise<string> {
  const data = await apiFetch<LoginResponse>("/api/v1/auth/login", {
    method: "POST",
    body: JSON.stringify({ username, password }),
  })
  setToken(data.access_token)
  setRefreshToken(data.refresh_token)
  return data.access_token
}
