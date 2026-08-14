import { useState } from 'react'
import { createFileRoute } from '@tanstack/react-router'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Store, Plus, Search, MoreHorizontal, CheckCircle2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuLabel, DropdownMenuSeparator, DropdownMenuTrigger } from '@/components/ui/dropdown-menu'
import { Badge } from '@/components/ui/badge'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { axiosInstance } from '@/lib/api/client'

export const Route = createFileRoute('/$locale/_dashboard/shops')({
    component: ShopsPage,
})

function ShopsPage() {
    const [searchQuery, setSearchQuery] = useState('')
    const [isAddShopOpen, setIsAddShopOpen] = useState(false)
    const [newShop, setNewShop] = useState({ name: '', address: '', owner_email: '' })
    const [activeShopId, setActiveShopId] = useState<string | null>(() => localStorage.getItem('activeShopId'))
    const queryClient = useQueryClient()

    // Selecting a shop stores it as the active tenant; the axios client sends it as
    // the X-Shop-Id header on every request, so the backend scopes data to it.
    const selectShop = (shopId: string) => {
        localStorage.setItem('activeShopId', shopId)
        setActiveShopId(shopId)
        queryClient.invalidateQueries()
    }

    const { data: shopsResponse, isLoading } = useQuery({
        queryKey: ['shops'],
        queryFn: async () => {
            const res = await axiosInstance.get('/shops')
            return res.data.data
        }
    })

    const createShopMutation = useMutation({
        mutationFn: async (shop: any) => {
            const res = await axiosInstance.post('/shops', shop)
            return res.data
        },
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['shops'] })
            setIsAddShopOpen(false)
            setNewShop({ name: '', address: '', owner_email: '' })
        }
    })

    const handleCreateShop = () => {
        createShopMutation.mutate(newShop)
    }

    const shops = shopsResponse || []
    const filteredShops = shops.filter((shop: any) => 
        shop.name?.toLowerCase().includes(searchQuery.toLowerCase())
    )


    return (
        <div className="space-y-6">
            <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
                <div>
                    <h2 className="text-2xl font-bold tracking-tight">Shops Management</h2>
                    <p className="text-muted-foreground">Super Admin control panel to manage all tenants (shops).</p>
                </div>
                <div className="flex flex-col sm:flex-row items-center gap-2">
                    <div className="relative w-full sm:w-64">
                        <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
                        <Input
                            type="search"
                            placeholder="Search shops..."
                            className="pl-8 w-full bg-background"
                            value={searchQuery}
                            onChange={(e) => setSearchQuery(e.target.value)}
                        />
                    </div>
                    <Dialog open={isAddShopOpen} onOpenChange={setIsAddShopOpen}>
                        <DialogTrigger asChild>
                            <Button className="w-full sm:w-auto">
                                <Plus className="mr-2 h-4 w-4" />
                                Register New Shop
                            </Button>
                        </DialogTrigger>
                        <DialogContent>
                            <DialogHeader>
                                <DialogTitle>Register New Shop</DialogTitle>
                                <DialogDescription>
                                    Create a new tenant workspace. An owner account will also be required.
                                </DialogDescription>
                            </DialogHeader>
                            <div className="grid gap-4 py-4">
                                <div className="grid gap-2">
                                    <label htmlFor="name" className="text-sm font-medium">Shop Name</label>
                                    <Input id="name" placeholder="e.g. Fiplex Electronics" value={newShop.name} onChange={e => setNewShop({...newShop, name: e.target.value})} />
                                </div>
                                <div className="grid gap-2">
                                    <label htmlFor="address" className="text-sm font-medium">Location / Address</label>
                                    <Input id="address" placeholder="e.g. Makumbusho, Dar es Salaam" value={newShop.address} onChange={e => setNewShop({...newShop, address: e.target.value})} />
                                </div>
                                <div className="grid gap-2">
                                    <label htmlFor="owner_email" className="text-sm font-medium">Owner Email</label>
                                    <Input id="owner_email" type="email" placeholder="owner@shop.com" value={newShop.owner_email} onChange={e => setNewShop({...newShop, owner_email: e.target.value})} />
                                </div>
                            </div>
                            <DialogFooter>
                                <Button variant="outline" onClick={() => setIsAddShopOpen(false)}>Cancel</Button>
                                <Button type="submit" disabled={createShopMutation.isPending} onClick={handleCreateShop}>
                                    {createShopMutation.isPending ? 'Creating...' : 'Create Shop'}
                                </Button>
                            </DialogFooter>
                        </DialogContent>
                    </Dialog>
                </div>
            </div>

            <div className="grid gap-4 md:grid-cols-3">
                <Card className="bg-primary/5 border-primary/20">
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">Total Active Shops</CardTitle>
                        <CheckCircle2 className="h-4 w-4 text-primary" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">{shops.filter((s: any) => s.is_active).length}</div>
                        <p className="text-xs text-muted-foreground">{shops.length} total registered</p>
                    </CardContent>
                </Card>
            </div>

            <Card className="border-border">
                <Table>
                    <TableHeader>
                        <TableRow>
                            <TableHead>Shop Name</TableHead>
                            <TableHead>Location</TableHead>
                            <TableHead>Owner</TableHead>
                            <TableHead>Status</TableHead>
                            <TableHead>Registered</TableHead>
                            <TableHead className="w-[80px]"></TableHead>
                        </TableRow>
                    </TableHeader>
                    <TableBody>
                        {isLoading ? (
                            <TableRow>
                                <TableCell colSpan={6} className="text-center h-24 text-muted-foreground">Loading shops...</TableCell>
                            </TableRow>
                        ) : filteredShops.length === 0 ? (
                            <TableRow>
                                <TableCell colSpan={6} className="text-center h-24 text-muted-foreground">No shops found.</TableCell>
                            </TableRow>
                        ) : filteredShops.map((shop: any) => (
                            <TableRow key={shop.id} className={activeShopId === shop.id ? 'bg-primary/5' : ''}>
                                <TableCell className="font-medium">
                                    <div className="flex items-center gap-2">
                                        <div className="p-2 rounded-md bg-secondary text-secondary-foreground">
                                            <Store className="h-4 w-4" />
                                        </div>
                                        {shop.name}
                                        {activeShopId === shop.id && (
                                            <Badge variant="outline" className="ml-1 border-primary text-primary">Active</Badge>
                                        )}
                                    </div>
                                </TableCell>
                                <TableCell className="text-muted-foreground">{shop.address || 'N/A'}</TableCell>
                                <TableCell>{shop.owner_email || 'No Owner'}</TableCell>
                                <TableCell>
                                    {shop.is_active ? (
                                        <Badge variant="default" className="bg-emerald-500/10 text-emerald-500 hover:bg-emerald-500/20">Active</Badge>
                                    ) : (
                                        <Badge variant="secondary" className="text-muted-foreground">Inactive</Badge>
                                    )}
                                </TableCell>
                                <TableCell className="text-muted-foreground">{new Date(shop.created_at).toLocaleDateString()}</TableCell>
                                <TableCell>
                                    <DropdownMenu>
                                        <DropdownMenuTrigger asChild>
                                            <Button variant="ghost" size="icon" className="h-8 w-8">
                                                <MoreHorizontal className="h-4 w-4" />
                                                <span className="sr-only">Open menu</span>
                                            </Button>
                                        </DropdownMenuTrigger>
                                        <DropdownMenuContent align="end">
                                            <DropdownMenuLabel>Actions</DropdownMenuLabel>
                                            <DropdownMenuItem onClick={() => selectShop(shop.id)} disabled={activeShopId === shop.id}>
                                                {activeShopId === shop.id ? 'Currently Active' : 'Set as Active Shop'}
                                            </DropdownMenuItem>
                                            <DropdownMenuItem>Edit Settings</DropdownMenuItem>
                                            <DropdownMenuSeparator />
                                            {shop.is_active ? (
                                                <DropdownMenuItem className="text-destructive">Deactivate Shop</DropdownMenuItem>
                                            ) : (
                                                <DropdownMenuItem className="text-emerald-500">Activate Shop</DropdownMenuItem>
                                            )}
                                        </DropdownMenuContent>
                                    </DropdownMenu>
                                </TableCell>
                            </TableRow>
                        ))}
                    </TableBody>
                </Table>
            </Card>
        </div>
    )
}
