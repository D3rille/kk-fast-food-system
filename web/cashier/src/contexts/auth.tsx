/* eslint-disable react-refresh/only-export-components */
import { createContext, useContext, useEffect, useState } from "react"
import { cashierLogin } from "@/api/auth"
import { getToken, clearTokens, setAuthFailureHandler } from "@/api/client"

interface AuthContextValue {
  token: string | null
  login: (username: string, password: string) => Promise<void>
  logout: () => void
}

const AuthContext = createContext<AuthContextValue | null>(null)

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [token, setTokenState] = useState<string | null>(() => getToken())

  useEffect(() => {
    // If a background token refresh ultimately fails (refresh token expired
    // or rejected), drop the session and send the cashier back to sign in.
    setAuthFailureHandler(() => {
      setTokenState(null)
      window.location.href = "/"
    })
  }, [])

  async function login(username: string, password: string) {
    const t = await cashierLogin(username, password)
    setTokenState(t)
  }

  function logout() {
    clearTokens()
    setTokenState(null)
  }

  return <AuthContext.Provider value={{ token, login, logout }}>{children}</AuthContext.Provider>
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error("useAuth must be used within AuthProvider")
  return ctx
}
