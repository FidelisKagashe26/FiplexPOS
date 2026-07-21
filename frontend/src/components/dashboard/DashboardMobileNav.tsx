import { Link } from '@tanstack/react-router'
import { Zap, Menu } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Sheet, SheetContent, SheetTrigger, SheetHeader, SheetTitle } from '@/components/ui/sheet'
import { cn } from '@/lib/utils'

interface DashboardMobileNavProps {
    t: any
    locale: string
    branding: any
    filteredMenu: any[]
}

export function DashboardMobileNav({
    t,
    locale,
    branding,
    filteredMenu
}: DashboardMobileNavProps) {
    return (
        <Sheet>
            <SheetTrigger asChild>
                <Button
                    variant="outline"
                    size="icon"
                    className="shrink-0 md:hidden absolute left-4 top-4 z-10 rounded-xl"
                >
                    <Menu className="h-5 w-5" />
                    <span className="sr-only">{t('dashboard.toggle_nav')}</span>
                </Button>
            </SheetTrigger>
            <SheetContent data-app-sidebar side="left" className="flex w-[280px] flex-col gap-0 border-r border-sidebar-border bg-sidebar/95 p-0 backdrop-blur-2xl">
                {/* Header */}
                <SheetHeader className="p-4 pb-3 border-b border-border/50">
                    <SheetTitle className="sr-only">Navigation</SheetTitle>
                    <Link
                        to="/$locale"
                        params={{ locale } as any}
                        className="flex items-center gap-3 font-heading font-bold"
                    >
                        {branding?.app_logo ? (
                            <img src={branding.app_logo} alt={t('settings.branding.logo')} className="h-9 w-9 object-contain rounded-xl" />
                        ) : (
                            <div className="h-9 w-9 rounded-xl bg-primary flex items-center justify-center">
                                <Zap className="h-5 w-5 text-primary-foreground" />
                            </div>
                        )}
                        <span className="text-sm tracking-tight">{branding?.app_name || t('dashboard.brand_name')}</span>
                    </Link>
                </SheetHeader>

                {/* Navigation - scrollable */}
                <nav className="flex-1 overflow-y-auto py-3 px-3">
                    <div className="grid gap-1">
                        {filteredMenu.map((item) => (
                            <Link
                                key={item.to}
                                to={item.to}
                                params={{ locale } as any}
                                activeOptions={{ exact: item.to === '/$locale' }}
                                className={cn(
                                    "flex items-center gap-3 rounded-xl px-3 py-2.5 text-sm font-medium text-muted-foreground transition-all duration-200",
                                    "hover:bg-primary/10 hover:text-primary",
                                    "[&.active]:bg-primary [&.active]:text-primary-foreground [&.active]:shadow-md [&.active]:shadow-primary/20"
                                )}
                            >
                                <item.icon className="h-[18px] w-[18px]" />
                                {item.label}
                            </Link>
                        ))}
                    </div>
                </nav>
            </SheetContent>
        </Sheet>
    )
}
