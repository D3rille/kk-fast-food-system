import { Outlet, Link, useNavigate } from "@tanstack/react-router"
import { LayoutGrid, PackageX, LogOut, ChefHat } from "lucide-react"
import { logout, getStoredUser } from "@/api/auth"

const navItems = [
  { to: "/items" as const, icon: PackageX, label: "Menu Items" },
  { to: "/categories" as const, icon: LayoutGrid, label: "Categories" },
]

export function Layout() {
  const navigate = useNavigate()
  const user = getStoredUser()

  function handleLogout() {
    logout()
    void navigate({ to: "/login" })
  }

  return (
    <div className="flex min-h-screen">
      {/* Sidebar */}
      <aside className="w-56 shrink-0 flex flex-col bg-sidebar border-r border-sidebar-border">
        {/* Brand */}
        <div className="flex items-center gap-2.5 px-4 py-5 border-b border-sidebar-border">
          <div className="flex items-center justify-center w-7 h-7 rounded-md bg-sidebar-primary/20 shrink-0">
            <ChefHat size={14} className="text-sidebar-primary" />
          </div>
          <div>
            <div className="font-bold text-sm leading-none text-sidebar-foreground">Admin</div>
            <div className="text-[10px] text-muted-foreground mt-0.5 leading-none">Kitchen Management</div>
          </div>
        </div>

        {/* Nav */}
        <nav className="flex-1 p-3 space-y-0.5">
          {navItems.map(({ to, icon: Icon, label }) => (
            <Link
              key={to}
              to={to}
              className="flex items-center gap-2.5 px-3 py-2 rounded-md text-sm text-sidebar-foreground hover:bg-sidebar-accent hover:text-sidebar-accent-foreground transition-colors"
              activeProps={{
                className:
                  "flex items-center gap-2.5 px-3 py-2 rounded-md text-sm bg-sidebar-accent text-sidebar-accent-foreground font-medium",
              }}
            >
              <Icon size={15} />
              {label}
            </Link>
          ))}
        </nav>

        {/* Footer */}
        <div className="p-3 border-t border-sidebar-border space-y-0.5">
          {user && (
            <div className="px-3 py-2">
              <div className="text-sm font-medium text-sidebar-foreground truncate">{user.username}</div>
              <div className="text-xs text-muted-foreground capitalize">{user.role}</div>
            </div>
          )}
          <button
            onClick={handleLogout}
            className="flex items-center gap-2.5 w-full px-3 py-2 rounded-md text-sm text-muted-foreground hover:bg-destructive/15 hover:text-destructive transition-colors"
          >
            <LogOut size={15} />
            Sign out
          </button>
        </div>
      </aside>

      {/* Main content */}
      <main className="flex-1 min-w-0 overflow-auto">
        <Outlet />
      </main>
    </div>
  )
}
