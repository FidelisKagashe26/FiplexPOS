/**
 * DashboardCharts.tsx
 *
 * Komponen ini sengaja dipisah dari dashboard index agar Recharts (~440 KB)
 * di-lazy load dan tidak masuk initial bundle.
 *
 * Di-import via: const DashboardCharts = lazy(() => import('@/components/dashboard/DashboardCharts'))
 */
import {
    ResponsiveContainer,
    BarChart,
    Bar,
    XAxis,
    YAxis,
    Tooltip as RechartsTooltip,
    CartesianGrid,
    PieChart,
    Pie,
    Cell,
    Legend,
} from 'recharts'
import { useTranslation } from 'react-i18next'

const COLORS = ['var(--chart-1)', 'var(--chart-4)', 'var(--chart-3)', 'var(--chart-2)', 'var(--chart-5)', 'var(--chart-3)']

interface SalesData {
    date?: string
    total_sales?: number
}

interface PaymentData {
    payment_method_name?: string
    order_count?: number
}

interface DashboardChartsProps {
    salesData: SalesData[] | undefined
    paymentsData: PaymentData[] | undefined
    formatCurrency: (value: number) => string
    formatDate: (dateString: string) => string
    chartType: 'bar' | 'pie'
}

export default function DashboardCharts({
    salesData,
    paymentsData,
    formatCurrency,
    formatDate,
    chartType,
}: DashboardChartsProps) {
    const { t } = useTranslation()

    if (chartType === 'bar') {
        return (
            <ResponsiveContainer width="100%" height={350}>
                <BarChart data={salesData}>
                    <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="var(--border)" />
                    <XAxis
                        dataKey="date"
                        stroke="var(--muted-foreground)"
                        fontSize={12}
                        tickLine={false}
                        axisLine={false}
                        tickFormatter={(value) => formatDate(value)}
                    />
                    <YAxis
                        stroke="var(--muted-foreground)"
                        fontSize={12}
                        tickLine={false}
                        axisLine={false}
                        tickFormatter={(value) => formatCurrency(Number(value || 0))}
                    />
                    <RechartsTooltip
                        formatter={(value: any) => formatCurrency(Number(value || 0))}
                        labelFormatter={(label) => formatDate(label)}
                    />
                    <Bar dataKey="total_sales" fill="var(--chart-1)" radius={[8, 8, 0, 0]} name={t('reports.sales.revenue')} />
                </BarChart>
            </ResponsiveContainer>
        )
    }

    // chartType === 'pie'
    return (
        <ResponsiveContainer width="100%" height={300}>
            <PieChart>
                <Pie
                    data={paymentsData}
                    cx="50%"
                    cy="50%"
                    innerRadius={60}
                    outerRadius={100}
                    fill="var(--chart-1)"
                    paddingAngle={5}
                    dataKey="order_count"
                    nameKey="payment_method_name"
                    label={({ percent }: any) => `${((percent || 0) * 100).toFixed(0)}%`}
                >
                    {(paymentsData || []).map((_entry, index) => (
                        <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                    ))}
                </Pie>
                <RechartsTooltip />
                <Legend verticalAlign="bottom" height={36} />
            </PieChart>
        </ResponsiveContainer>
    )
}
