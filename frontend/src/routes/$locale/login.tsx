import { createFileRoute, useRouter, useParams } from '@tanstack/react-router'
import { redirect } from '@tanstack/react-router'
import { useEffect, useState } from 'react'
import { useTheme } from 'next-themes'
import { meQueryOptions } from "@/lib/api/query/auth.ts";
import { useAuth } from "@/context/AuthContext";
import { useTranslation } from 'react-i18next'
import { LoginForm } from "@/components/auth/LoginForm"

type LoginTheme = 'light' | 'dark'

export function getLoginThemeForHour(hour: number): LoginTheme {
    return hour >= 6 && hour < 18 ? 'light' : 'dark'
}

export const Route = createFileRoute('/$locale/login')({
    ssr: false,
    loader: async ({ context: { queryClient }, params }) => {
        try {
            const me = await queryClient.ensureQueryData(meQueryOptions())
            if (me) {
                throw redirect({
                    to: '/$locale',
                    params: { locale: params.locale }
                })
            }
        } catch (error: any) {
            const status = error?.response?.status ?? error?.status ?? error?.cause?.status
            if (status === 401) {
                return
            }
        }
    },
    component: LoginPage,
})

function LoginPage() {
    const { locale } = useParams({ from: '/$locale/login' })
    const { t } = useTranslation()
    const auth = useAuth()
    const router = useRouter()
    const { setTheme } = useTheme()
    const [loginTheme, setLoginTheme] = useState<LoginTheme>(() => getLoginThemeForHour(new Date().getHours()))

    useEffect(() => {
        const updateTheme = () => setLoginTheme(getLoginThemeForHour(new Date().getHours()))
        const intervalId = window.setInterval(updateTheme, 60_000)
        window.addEventListener('focus', updateTheme)

        return () => {
            window.clearInterval(intervalId)
            window.removeEventListener('focus', updateTheme)
        }
    }, [])

    return (
        <main
            className="relative flex min-h-screen items-center justify-center overflow-hidden bg-cover bg-center px-5 py-10"
            style={{ backgroundImage: "url('/login-background.png')", colorScheme: loginTheme }}
            data-login-theme={loginTheme}
        >
            <div
                className={loginTheme === 'dark' ? 'absolute inset-0 bg-slate-950/68' : 'absolute inset-0 bg-white/18'}
                aria-hidden="true"
            />
            <LoginForm
                t={t}
                auth={auth}
                mode={loginTheme}
                onSubmitSuccess={async () => {
                    if (!window.localStorage.getItem('theme')) {
                        setTheme(loginTheme)
                    }
                    await router.invalidate()
                    await router.navigate({
                        to: '/$locale',
                        params: { locale },
                        replace: true
                    })
                }}
            />
        </main>
    )
}