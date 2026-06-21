import { NavLink, useMatch } from 'react-router-dom'
import {
  Sidebar,
  SidebarContent,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from '@/components/ui/sidebar'

const NAV_ITEMS = [
  { to: '/datasets', label: 'Datasets' },
  { to: '/experiments', label: 'Experiments' },
]

function NavItem({ to, label }: { to: string; label: string }) {
  const isActive = !!useMatch(to)
  return (
    <SidebarMenuItem>
      <SidebarMenuButton asChild isActive={isActive}>
        <NavLink to={to}>{label}</NavLink>
      </SidebarMenuButton>
    </SidebarMenuItem>
  )
}

export function AppSidebar() {
  return (
    <Sidebar>
      <SidebarHeader className="px-4 py-3 text-sm font-semibold">
        LLM Experiments
      </SidebarHeader>
      <SidebarContent>
        <SidebarMenu>
          {NAV_ITEMS.map((item) => (
            <NavItem key={item.to} {...item} />
          ))}
        </SidebarMenu>
      </SidebarContent>
    </Sidebar>
  )
}
