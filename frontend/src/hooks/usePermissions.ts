import { useQuery } from '@tanstack/react-query'
import { axiosInstance } from '@/lib/api/client'

export type MyPermissions = {
    role: string
    shop_id: string | null
    permissions: string[]
}

// usePermissions fetches the current user's effective permission list from the
// backend (/auth/me/permissions). Authorization itself is always enforced on the
// backend — this list is only used to show/hide UI. Admins receive the full
// catalog from the server, so `can()` returns true for them across the board.
export function usePermissions() {
    const { data, isLoading, isError } = useQuery<MyPermissions>({
        queryKey: ['auth', 'my-permissions'],
        queryFn: async () => {
            const res = await axiosInstance.get('/auth/me/permissions')
            return res.data.data as MyPermissions
        },
        staleTime: 1000 * 60 * 5,
        retry: false,
    })

    const permissions = data?.permissions ?? []
    const role = data?.role ?? ''

    const can = (permission: string) => permissions.includes(permission)
    const canAny = (...perms: string[]) => perms.some((p) => permissions.includes(p))
    const canAll = (...perms: string[]) => perms.every((p) => permissions.includes(p))

    return { role, permissions, shopId: data?.shop_id ?? null, can, canAny, canAll, isLoading, isError }
}
