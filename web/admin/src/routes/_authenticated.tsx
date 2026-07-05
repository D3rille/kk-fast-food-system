import { createFileRoute, redirect } from '@tanstack/react-router'
import { isAuthenticated } from '@/api/auth'
import { Layout } from '@/components/Layout'

export const Route = createFileRoute('/_authenticated')({
  beforeLoad: () => {
    if (!isAuthenticated()) throw redirect({ to: '/login' })
  },
  component: Layout,
})
