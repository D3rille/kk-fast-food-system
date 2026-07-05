import { apiFetch, setToken, setRefreshToken, clearToken } from "./client"
import type { LoginResponse, User } from "@/types/api"

export async function login(username: string, password: string): Promise<User> {
  const data = await apiFetch<LoginResponse>("/api/v1/auth/login", {
    method: "POST",
    body: JSON.stringify({ username, password }),
  })
  setToken(data.access_token)
  setRefreshToken(data.refresh_token)
  localStorage.setItem("admin_user", JSON.stringify(data.user))
  return data.user
}

export function logout(): void {
  clearToken()
}

export function getStoredUser(): User | null {
  try {
    const raw = localStorage.getItem("admin_user")
    return raw ? (JSON.parse(raw) as User) : null
  } catch {
    return null
  }
}

export function isAuthenticated(): boolean {
  return !!localStorage.getItem("admin_access_token")
}
