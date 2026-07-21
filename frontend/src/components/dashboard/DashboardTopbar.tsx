import { Link } from '@tanstack/react-router'
import { ChevronDown, LogOut, User as UserIcon } from 'lucide-react'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { LanguageSwitcher } from '@/components/LanguageSwitcher'
import { ThemeToggle } from '@/components/ThemeToggle'
import { ShiftControl } from './ShiftControl'

interface DashboardTopbarProps {
  t: any
  locale: string
  user: any
  handleLogout: () => void
}

export function DashboardTopbar({ t, locale, user, handleLogout }: DashboardTopbarProps) {
  const userName = user?.username ?? 'User'
  const userRole = user?.role ?? ''

  return (
    <header data-dashboard-topbar className="flex h-16 shrink-0 items-center gap-3 px-3 pl-14 md:px-5 md:pl-5">
      <div className="hidden min-w-0 sm:block">
        <p className="text-[11px] font-semibold uppercase tracking-[0.16em] text-primary">POS Kasir</p>
        <p className="truncate text-sm font-semibold text-foreground">{t('dashboard.workspace', 'Store workspace')}</p>
      </div>

      <div className="ml-auto flex min-w-0 items-center gap-2">
        <div className="hidden w-48 lg:block">
          <ShiftControl />
        </div>

        <div className="topbar-language w-11 sm:w-40">
          <LanguageSwitcher />
        </div>

        <div className="topbar-icon-control">
          <ThemeToggle />
        </div>

        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="outline" className="topbar-user-trigger gap-2 px-2.5 sm:px-3">
              <Avatar className="h-7 w-7">
                <AvatarImage src={user?.avatar || undefined} alt={userName} />
                <AvatarFallback><UserIcon className="h-3.5 w-3.5" /></AvatarFallback>
              </Avatar>
              <span className="hidden min-w-0 text-left md:block">
                <span className="block max-w-28 truncate text-sm font-semibold leading-tight">{userName}</span>
                <span className="block max-w-28 truncate text-[11px] font-normal leading-tight text-muted-foreground">{userRole}</span>
              </span>
              <ChevronDown className="hidden h-3.5 w-3.5 text-muted-foreground md:block" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" className="w-56">
            <DropdownMenuLabel>
              <span className="block truncate">{userName}</span>
              <span className="block text-xs font-normal text-muted-foreground">{userRole}</span>
            </DropdownMenuLabel>
            <DropdownMenuSeparator />
            <DropdownMenuItem asChild>
              <Link to="/$locale/account" params={{ locale } as any}>
                <UserIcon className="mr-2 h-4 w-4" />
                {t('navigation.account', 'Account')}
              </Link>
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>

        <Button
          variant="outline"
          onClick={handleLogout}
          className="topbar-logout gap-2 px-3 text-destructive hover:bg-destructive/10 hover:text-destructive"
          title={t('common.logout')}
        >
          <LogOut className="h-4 w-4" />
          <span className="hidden xl:inline">{t('common.logout')}</span>
        </Button>
      </div>
    </header>
  )
}