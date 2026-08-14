import { useMemo, useState } from 'react'
import { createFileRoute } from '@tanstack/react-router'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Shield, Plus, KeyRound, Users, Edit2, Trash2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog'
import { Badge } from '@/components/ui/badge'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { axiosInstance } from '@/lib/api/client'
import { toast } from 'sonner'

export const Route = createFileRoute('/$locale/_dashboard/roles')({
    component: RolesAndPermissionsPage,
})

type Role = {
    id: string
    name: string
    description: string | null
    shop_id: string | null
}

type Permission = {
    id: string
    name: string
    module: string
    description: string
}

// Authorization itself is enforced entirely on the backend. This page only
// reads the permission catalog and a role's current permissions from the API and
// writes the chosen set back — it never decides access on the client.
function RolesAndPermissionsPage() {
    const [isAddRoleOpen, setIsAddRoleOpen] = useState(false)
    const [newRole, setNewRole] = useState({ name: '', description: '' })
    const [selectedRoleId, setSelectedRoleId] = useState<string | null>(null)
    const [draft, setDraft] = useState<Set<string>>(new Set())
    const queryClient = useQueryClient()

    const { data: roles = [], isLoading } = useQuery<Role[]>({
        queryKey: ['roles'],
        queryFn: async () => {
            const res = await axiosInstance.get('/roles')
            return res.data.data ?? []
        },
    })

    // Full permission catalog (grouped by module for the matrix).
    const { data: permissions = [] } = useQuery<Permission[]>({
        queryKey: ['permissions'],
        queryFn: async () => {
            const res = await axiosInstance.get('/permissions')
            return res.data.data ?? []
        },
    })

    // The currently-selected role's permissions, used to seed the draft.
    const { isFetching: isLoadingRolePerms } = useQuery({
        queryKey: ['role-permissions', selectedRoleId],
        enabled: !!selectedRoleId,
        queryFn: async () => {
            const res = await axiosInstance.get(`/roles/${selectedRoleId}/permissions`)
            const perms: Permission[] = res.data.data ?? []
            setDraft(new Set(perms.map((p) => p.id)))
            return perms
        },
    })

    const createRoleMutation = useMutation({
        mutationFn: async (role: { name: string; description: string }) => {
            const res = await axiosInstance.post('/roles', role)
            return res.data
        },
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['roles'] })
            setIsAddRoleOpen(false)
            setNewRole({ name: '', description: '' })
            toast.success('Role created')
        },
        onError: () => toast.error('Failed to create role'),
    })

    const savePermissionsMutation = useMutation({
        mutationFn: async () => {
            if (!selectedRoleId) return
            await axiosInstance.put(`/roles/${selectedRoleId}/permissions`, {
                permission_ids: Array.from(draft),
            })
        },
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['role-permissions', selectedRoleId] })
            toast.success('Permissions updated')
        },
        onError: () => toast.error('Failed to save permissions'),
    })

    const grouped = useMemo(() => {
        const map = new Map<string, Permission[]>()
        for (const p of permissions) {
            const list = map.get(p.module) ?? []
            list.push(p)
            map.set(p.module, list)
        }
        return Array.from(map.entries()).sort((a, b) => a[0].localeCompare(b[0]))
    }, [permissions])

    const selectedRole = roles.find((r) => r.id === selectedRoleId) || null

    const toggle = (permId: string) => {
        setDraft((prev) => {
            const next = new Set(prev)
            if (next.has(permId)) next.delete(permId)
            else next.add(permId)
            return next
        })
    }

    const actionLabel = (name: string) => {
        const parts = name.split('.')
        const action = parts.length > 1 ? parts[1] : name
        return action.charAt(0).toUpperCase() + action.slice(1)
    }

    return (
        <div className="space-y-6">
            <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
                <div>
                    <h2 className="text-2xl font-bold tracking-tight">Roles & Permissions</h2>
                    <p className="text-muted-foreground">Create custom roles and assign them specific permissions.</p>
                </div>
                <div className="flex items-center gap-2">
                    <Dialog open={isAddRoleOpen} onOpenChange={setIsAddRoleOpen}>
                        <DialogTrigger asChild>
                            <Button>
                                <Plus className="mr-2 h-4 w-4" />
                                Create New Role
                            </Button>
                        </DialogTrigger>
                        <DialogContent className="max-w-md">
                            <DialogHeader>
                                <DialogTitle>Create New Role</DialogTitle>
                                <DialogDescription>
                                    Define a new role and later assign permissions to it.
                                </DialogDescription>
                            </DialogHeader>
                            <div className="grid gap-4 py-4">
                                <div className="grid gap-2">
                                    <label htmlFor="role_name" className="text-sm font-medium">Role Name</label>
                                    <Input id="role_name" placeholder="e.g. Supervisor" value={newRole.name} onChange={e => setNewRole({ ...newRole, name: e.target.value })} />
                                </div>
                                <div className="grid gap-2">
                                    <label htmlFor="role_desc" className="text-sm font-medium">Description (Optional)</label>
                                    <Textarea id="role_desc" placeholder="Briefly describe what this role does..." value={newRole.description} onChange={e => setNewRole({ ...newRole, description: e.target.value })} />
                                </div>
                            </div>
                            <DialogFooter>
                                <Button variant="outline" onClick={() => setIsAddRoleOpen(false)}>Cancel</Button>
                                <Button type="submit" disabled={createRoleMutation.isPending || !newRole.name} onClick={() => createRoleMutation.mutate(newRole)}>
                                    {createRoleMutation.isPending ? 'Creating...' : 'Create Role'}
                                </Button>
                            </DialogFooter>
                        </DialogContent>
                    </Dialog>
                </div>
            </div>

            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
                {isLoading ? (
                    <div className="col-span-3 text-center py-10 text-muted-foreground">Loading roles...</div>
                ) : roles.length === 0 ? (
                    <div className="col-span-3 text-center py-10 text-muted-foreground">No roles found.</div>
                ) : roles.map((role) => (
                    <Card key={role.id} className={`flex flex-col transition-colors ${selectedRoleId === role.id ? 'border-primary' : 'border-border/60 hover:border-border'}`}>
                        <CardHeader className="pb-3">
                            <div className="flex items-center justify-between">
                                <div className="flex items-center gap-2">
                                    <div className="p-2 rounded-lg bg-primary/10 text-primary">
                                        <Shield className="h-5 w-5" />
                                    </div>
                                    <CardTitle className="text-lg">{role.name}</CardTitle>
                                </div>
                                <div className="flex gap-1">
                                    <Button variant="ghost" size="icon" className="h-7 w-7 text-muted-foreground hover:text-foreground">
                                        <Edit2 className="h-3.5 w-3.5" />
                                    </Button>
                                    {role.name !== 'admin' && (
                                        <Button variant="ghost" size="icon" className="h-7 w-7 text-muted-foreground hover:text-destructive">
                                            <Trash2 className="h-3.5 w-3.5" />
                                        </Button>
                                    )}
                                </div>
                            </div>
                        </CardHeader>
                        <CardContent className="flex-1 pb-4">
                            <p className="text-sm text-muted-foreground mb-4">
                                {role.description || 'No description provided.'}
                            </p>
                            <div className="flex items-center gap-4 text-sm">
                                <div className="flex items-center gap-1.5 text-muted-foreground">
                                    <Users className="h-4 w-4" />
                                    <span><span className="font-medium text-foreground">--</span> users</span>
                                </div>
                                <div className="flex items-center gap-1.5 text-muted-foreground">
                                    <KeyRound className="h-4 w-4" />
                                    <span><span className="font-medium text-foreground">--</span> perms</span>
                                </div>
                            </div>
                        </CardContent>
                        <div className="px-6 py-4 bg-muted/30 border-t flex justify-between items-center rounded-b-xl">
                            <Badge variant="outline" className="bg-background">{role.shop_id ? 'Shop Role' : 'Global Role'}</Badge>
                            <Button variant="link" className="h-auto p-0 text-primary" onClick={() => setSelectedRoleId(role.id)}>
                                Manage Permissions →
                            </Button>
                        </div>
                    </Card>
                ))}
            </div>

            <div className="mt-8 rounded-xl border bg-card text-card-foreground shadow-sm overflow-hidden">
                <div className="p-6 border-b bg-muted/10">
                    <h3 className="text-lg font-semibold flex items-center gap-2">
                        <KeyRound className="h-5 w-5 text-primary" />
                        Permission Matrix
                    </h3>
                    <p className="text-sm text-muted-foreground mt-1">
                        {selectedRole
                            ? <>Configuring access for <span className="font-medium text-foreground">{selectedRole.name}</span>.</>
                            : 'Select a role above to configure its access controls.'}
                    </p>
                </div>

                {!selectedRole ? (
                    <div className="p-10 text-center text-muted-foreground">No role selected.</div>
                ) : selectedRole.name === 'admin' ? (
                    <div className="p-10 text-center text-muted-foreground">
                        The <span className="font-medium text-foreground">admin</span> role has full access to everything and cannot be restricted.
                    </div>
                ) : isLoadingRolePerms ? (
                    <div className="p-10 text-center text-muted-foreground">Loading permissions...</div>
                ) : (
                    <div className="divide-y">
                        {grouped.map(([module, perms]) => (
                            <div key={module} className="p-6">
                                <h4 className="text-sm font-semibold capitalize mb-3">{module.replace(/_/g, ' ')}</h4>
                                <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
                                    {perms.map((p) => (
                                        <label key={p.id} className="flex items-center gap-2 text-sm cursor-pointer">
                                            <input
                                                type="checkbox"
                                                checked={draft.has(p.id)}
                                                onChange={() => toggle(p.id)}
                                                className="h-4 w-4 rounded border-gray-300 text-primary focus:ring-primary accent-primary"
                                            />
                                            <span className="text-foreground">{actionLabel(p.name)}</span>
                                            <span className="text-muted-foreground text-xs">{p.description}</span>
                                        </label>
                                    ))}
                                </div>
                            </div>
                        ))}
                    </div>
                )}

                {selectedRole && selectedRole.name !== 'admin' && (
                    <div className="p-4 bg-muted/10 flex items-center justify-end gap-2 border-t">
                        <Button variant="outline" onClick={() => setSelectedRoleId(null)}>Cancel</Button>
                        <Button disabled={savePermissionsMutation.isPending} onClick={() => savePermissionsMutation.mutate()}>
                            {savePermissionsMutation.isPending ? 'Saving...' : 'Save Permissions'}
                        </Button>
                    </div>
                )}
            </div>
        </div>
    )
}
