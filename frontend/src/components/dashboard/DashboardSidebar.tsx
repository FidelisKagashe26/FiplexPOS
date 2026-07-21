import { Link } from '@tanstack/react-router'
import { Zap, Menu } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip'
import { cn } from '@/lib/utils'

interface DashboardSidebarProps {
    t: any
    locale: string
    branding: any
    isSidebarCollapsed: boolean
    setIsSidebarCollapsed: (collapsed: boolean) => void
    filteredMenu: any[]
}

export function DashboardSidebar({
    t,
    locale,
    branding,
    isSidebarCollapsed,
    setIsSidebarCollapsed,
    filteredMenu
}: DashboardSidebarProps) {
    return (
        <div className="hidden md:block">
            <div data-app-sidebar className="flex h-full max-h-screen flex-col gap-2 border-r border-sidebar-border bg-sidebar/85 shadow-xl shadow-slate-950/5 backdrop-blur-2xl">
                {/* Logo */}
                <div className={cn("flex h-16 items-center px-5", isSidebarCollapsed ? "justify-center" : "justify-between")}>
                    <Link
                        to="/$locale"
                        params={{ locale } as any}
                        className={cn("flex items-center gap-3 font-heading font-bold", isSidebarCollapsed && "hidden")}
                    >
                        {branding?.app_logo ? (
                            <img src={branding.app_logo} alt={t('settings.branding.logo')} className="h-9 w-9 object-contain rounded-xl" />
                        ) : (
                            <div className="h-9 w-9 rounded-xl bg-primary flex items-center justify-center">
                                <Zap className="h-5 w-5 text-primary-foreground" />
                            </div>
                        )}
                        <span className="text-xl truncate tracking-tight">{branding?.app_name || t('dashboard.brand_name')}</span>
                    </Link>
                    <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => setIsSidebarCollapsed(!isSidebarCollapsed)}
                        className="h-8 w-8 rounded-lg"
                    >
                        <Menu className="h-4 w-4" />
                        <span className="sr-only">Toggle Sidebar</span>
                    </Button>
                </div>

                {/* Navigation */}
                <div className="flex-1 py-2">
                    <TooltipProvider delayDuration={0}>
                        <nav className="grid items-start px-3 text-sm font-medium gap-1">
                            {filteredMenu.map((item) => (
                                <Tooltip key={item.to}>
                                    <TooltipTrigger asChild>
                                        <Link
                                            to={item.to}
                                            params={{ locale } as any}
                                            activeOptions={{ exact: item.to === '/$locale' }}
                                            className={cn(
                                                "flex items-center gap-3 rounded-xl px-3 py-2.5 text-muted-foreground transition-all duration-200",
                                                "hover:bg-primary/10 hover:text-primary",
                                                "[&.active]:bg-primary [&.active]:text-primary-foreground [&.active]:shadow-lg [&.active]:shadow-primary/25",
                                                isSidebarCollapsed && "justify-center px-2"
                                            )}
                                        >
                                            <item.icon className={cn("h-[18px] w-[18px]", isSidebarCollapsed ? "h-5 w-5" : "")} />
                                            {!isSidebarCollapsed && <span>{item.label}</span>}
                                        </Link>
                                    </TooltipTrigger>
                                    {isSidebarCollapsed && (
                                        <TooltipContent side="right">
                                            {item.label}
                                        </TooltipContent>
                                    )}
                                </Tooltip>
                            ))}
                        </nav>
                    </TooltipProvider>
                </div>
            </div>
        </div>
    )
}
