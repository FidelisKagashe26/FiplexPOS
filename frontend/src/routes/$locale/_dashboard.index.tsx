import { createFileRoute, Link, useParams } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'
import { useDashboardSummaryQuery, useSalesReportQuery, useProductPerformanceQuery, usePaymentMethodPerformanceQuery } from '@/lib/api/query/reports'
import { useQueryClient } from '@tanstack/react-query'
import { meQueryOptions } from '@/lib/api/query/auth'
import { useState } from 'react'
import { Store } from 'lucide-react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { DashboardHeader } from "@/components/dashboard/DashboardHeader"
import { DashboardStats } from "@/components/dashboard/DashboardStats"
import { DashboardSalesChart } from "@/components/dashboard/DashboardSalesChart"
import { DashboardTopProducts } from "@/components/dashboard/DashboardTopProducts"
import { DashboardPaymentChart } from "@/components/dashboard/DashboardPaymentChart"
import { DashboardQuickActions } from "@/components/dashboard/DashboardQuickActions"
import { formatDate, formatCurrency } from '@/lib/utils'


export const Route = createFileRoute('/$locale/_dashboard/')(({
    component: DashboardIndex,
    loader: ({ context: { queryClient } }: any) => queryClient.ensureQueryData(meQueryOptions()),
} as any))

function DashboardIndex() {
    const { t } = useTranslation()
    const queryClient = useQueryClient()
    const user = queryClient.getQueryData(meQueryOptions().queryKey) as any
    const { locale } = useParams({ from: '/$locale/_dashboard' })
    const hasActiveShop = !!localStorage.getItem('activeShopId')

    if (user?.role === 'superadmin' && !hasActiveShop) {
        return (
            <div className="mx-auto flex min-h-[60vh] max-w-xl items-center">
                <Card className="w-full border-primary/30 bg-primary/5">
                    <CardHeader>
                        <div className="mb-2 flex h-11 w-11 items-center justify-center rounded-xl bg-primary/15 text-primary">
                            <Store className="h-6 w-6" />
                        </div>
                        <CardTitle>Super Admin Workspace</CardTitle>
                        <CardDescription>
                            Start by registering a shop or select an existing shop. Store products, sales and staff only appear after a shop is selected.
                        </CardDescription>
                    </CardHeader>
                    <CardContent>
                        <Button asChild>
                            <Link to="/$locale/shops" params={{ locale }}>Manage Shops</Link>
                        </Button>
                    </CardContent>
                </Card>
            </div>
        )
    }
    const [endDate, setEndDate] = useState<string>(new Date().toISOString().split('T')[0])
    const [startDate, setStartDate] = useState<string>(
        new Date(new Date().setDate(new Date().getDate() - 30)).toISOString().split('T')[0]
    )

    const { data: summary, isLoading: isLoadingSummary } = useDashboardSummaryQuery(startDate, endDate)
    const { data: salesData, isLoading: isLoadingSales } = useSalesReportQuery(startDate, endDate)
    const { data: productsData, isLoading: isLoadingProducts } = useProductPerformanceQuery(startDate, endDate)
    const { data: paymentsData, isLoading: isLoadingPayments } = usePaymentMethodPerformanceQuery(startDate, endDate)

    const topProducts = productsData?.products?.slice(0, 5) || []

    return (
        <div className="flex flex-col gap-8">
            <DashboardHeader
                t={t}
                username={user?.username || 'User'}
                startDate={startDate}
                endDate={endDate}
                onDateChange={({ from, to }) => {
                    setStartDate(from)
                    setEndDate(to)
                }}
            />

            <DashboardStats
                t={t}
                summary={summary}
                isLoading={isLoadingSummary}
                formatCurrency={formatCurrency}
            />

            <div className="grid gap-4 grid-cols-1 lg:grid-cols-7">
                <DashboardSalesChart
                    t={t}
                    isLoading={isLoadingSales}
                    salesData={salesData}
                    paymentsData={paymentsData}
                    formatCurrency={formatCurrency}
                    formatDate={formatDate}
                />

                <DashboardTopProducts
                    t={t}
                    isLoading={isLoadingProducts}
                    topProducts={topProducts}
                    formatCurrency={formatCurrency}
                />
            </div>

            <div className="grid gap-4 grid-cols-1 lg:grid-cols-7">
                <DashboardPaymentChart
                    t={t}
                    isLoading={isLoadingPayments}
                    salesData={salesData}
                    paymentsData={paymentsData}
                    formatCurrency={formatCurrency}
                    formatDate={formatDate}
                />

                <DashboardQuickActions t={t} />
            </div>
        </div>
    )
}