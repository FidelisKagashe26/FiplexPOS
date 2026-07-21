import { Label } from '@/components/ui/label'
import { DateRangePicker } from '@/components/ui/date-range-picker'
import { Sparkles } from 'lucide-react'

interface DashboardHeaderProps {
    t: any
    username: string
    startDate: string
    endDate: string
    onDateChange: (range: { from: string; to: string }) => void
}

export function DashboardHeader({ t, username, startDate, endDate, onDateChange }: DashboardHeaderProps) {
    return (
        <header className="dashboard-hero relative overflow-hidden rounded-3xl px-6 py-6 text-white md:px-8 md:py-7">
            <div className="dashboard-hero-orb dashboard-hero-orb-primary" aria-hidden="true" />
            <div className="dashboard-hero-orb dashboard-hero-orb-secondary" aria-hidden="true" />

            <div className="relative z-10 flex flex-col gap-6 lg:flex-row lg:items-center lg:justify-between">
                <div className="min-w-0">
                    <div className="dashboard-hero-kicker mb-3 inline-flex items-center gap-2 rounded-full px-3 py-1.5 text-xs font-semibold uppercase tracking-[0.14em]">
                        <Sparkles className="h-3.5 w-3.5" />
                        {t('dashboard.store_overview', 'Store overview')}
                    </div>
                    <h1 className="font-heading text-2xl font-bold tracking-[-0.035em] md:text-3xl">
                        {t('auth.welcome_back', 'Welcome back')}, {username}!
                    </h1>
                    <p className="mt-1.5 max-w-xl text-sm text-white/62 md:text-base">
                        {t('dashboard.welcome_subtitle', 'Here is what is happening with your store today.')}
                    </p>
                </div>

                <div className="dashboard-hero-controls flex w-full flex-col gap-3 rounded-2xl p-3 sm:flex-row sm:items-end lg:w-auto">
                    <div className="flex items-center gap-2 px-1 pb-0.5 sm:pb-2">
                        <span className="relative flex h-2.5 w-2.5">
                            <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-emerald-300 opacity-60" />
                            <span className="relative inline-flex h-2.5 w-2.5 rounded-full bg-emerald-300" />
                        </span>
                        <span className="whitespace-nowrap text-xs font-medium text-white/70">
                            {t('dashboard.store_live', 'Store live')}
                        </span>
                    </div>
                    <div className="grid min-w-0 gap-1.5">
                        <Label className="px-1 text-[11px] font-semibold uppercase tracking-[0.13em] text-white/48">
                            {t('reports.date_range', 'Date Range')}
                        </Label>
                        <DateRangePicker
                            className="dashboard-date-picker"
                            date={{ from: startDate, to: endDate }}
                            onDateChange={onDateChange}
                        />
                    </div>
                </div>
            </div>
        </header>
    )
}