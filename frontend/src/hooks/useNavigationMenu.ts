import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import {
    FileText,
    LayoutDashboard,
    Package,
    Settings,
    ShoppingCart,
    User as UserIcon,
    Receipt,
    Tag,
    ActivityIcon,
    Users,
    Shield,
    Store
} from 'lucide-react'

const SUPER_ADMIN_ROLE = 'superadmin'

type DashboardMenuItem = {
    label: string
    icon: any
    to: string
    allowedRoles: string[]
}

// Super Admin works at platform level until a shop is selected. Once a shop is
// active, the normal store workspace becomes available for that selected tenant.
export function useNavigationMenu(userRole?: string) {
    const { t } = useTranslation()
    const [activeShopId, setActiveShopId] = useState(() => localStorage.getItem('activeShopId'))

    useEffect(() => {
        const syncActiveShop = () => setActiveShopId(localStorage.getItem('activeShopId'))
        window.addEventListener('active-shop-changed', syncActiveShop)
        window.addEventListener('storage', syncActiveShop)
        return () => {
            window.removeEventListener('active-shop-changed', syncActiveShop)
            window.removeEventListener('storage', syncActiveShop)
        }
    }, [])

    const storeRoles = ['admin', 'manager', 'cashier']
    const adminRoles = ['admin']
    const menuItems: DashboardMenuItem[] = [
        { label: t('sidebar.summary'), icon: LayoutDashboard, to: '/$locale', allowedRoles: storeRoles },
        { label: t('sidebar.pos'), icon: ShoppingCart, to: '/$locale/order', allowedRoles: storeRoles },
        { label: t('sidebar.transactions'), icon: Receipt, to: '/$locale/transactions', allowedRoles: storeRoles },
        { label: t('sidebar.product'), icon: Package, to: '/$locale/product', allowedRoles: storeRoles },
        { label: t('sidebar.promotions'), icon: Tag, to: '/$locale/promotions', allowedRoles: storeRoles },
        { label: t('sidebar.reports'), icon: FileText, to: '/$locale/reports', allowedRoles: adminRoles },
        { label: t('sidebar.customers', 'Customers'), icon: Users, to: '/$locale/customers', allowedRoles: storeRoles },
        { label: t('sidebar.users'), icon: UserIcon, to: '/$locale/users', allowedRoles: ['admin', 'manager'] },
        { label: 'Roles & Permissions', icon: Shield, to: '/$locale/roles', allowedRoles: adminRoles },
        { label: t('sidebar.settings'), icon: Settings, to: '/$locale/settings', allowedRoles: storeRoles },
        { label: t('sidebar.activity_logs'), icon: ActivityIcon, to: '/$locale/activity-logs', allowedRoles: adminRoles },
        { label: t('sidebar.account'), icon: UserIcon, to: '/$locale/account', allowedRoles: storeRoles },
    ]

    const shopsItem: DashboardMenuItem = {
        label: 'Shops', icon: Store, to: '/$locale/shops', allowedRoles: [SUPER_ADMIN_ROLE],
    }
    const accountItem: DashboardMenuItem = {
        label: t('sidebar.account'), icon: UserIcon, to: '/$locale/account', allowedRoles: [SUPER_ADMIN_ROLE],
    }

    if (userRole === SUPER_ADMIN_ROLE) {
        const storeWorkspace = activeShopId
            ? menuItems.filter(item => !['/$locale/users', '/$locale/roles', '/$locale/activity-logs', '/$locale/account'].includes(item.to) && item.allowedRoles.includes('admin'))
            : []
        return { menuItems: [shopsItem, ...storeWorkspace, accountItem], filteredMenu: [shopsItem, ...storeWorkspace, accountItem], activeShopId }
    }

    const filteredMenu = menuItems.filter(item => !!userRole && item.allowedRoles.includes(userRole))
    return { menuItems, filteredMenu, activeShopId }
}