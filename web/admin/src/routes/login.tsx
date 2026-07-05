import { createFileRoute, redirect } from '@tanstack/react-router'
import { isAuthenticated } from '@/api/auth'
import LoginPage from '@/pages/LoginPage'

export const Route = createFileRoute('/login')({
  beforeLoad: () => {
    if (isAuthenticated()) throw redirect({ to: '/items' })
  },
  component: LoginPage,
})
