import { useState } from 'react'
import { Input } from '@/components/ui/input'
import { PasswordInput } from "@/components/ui/password-input"
import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
import { useForm } from '@tanstack/react-form'
import { cn } from '@/lib/utils'
import * as z from 'zod'

type LoginTheme = 'light' | 'dark'

interface LoginFormProps {
    t: any
    auth: any
    mode: LoginTheme
    onSubmitSuccess: () => void
}

export function LoginForm({ t, auth, mode, onSubmitSuccess }: LoginFormProps) {
    const [serverError, setServerError] = useState<string | null>(null)
    const isDark = mode === 'dark'

    const form = useForm({
        defaultValues: {
            email: '',
            password: '',
        },
        validators: {
            onChange: z.object({
                email: z.string().email(),
                password: z.string().min(1)
            })
        },
        onSubmit: async ({ value }) => {
            setServerError(null)
            try {
                await auth.login({
                    email: value.email,
                    password: value.password,
                })
                onSubmitSuccess()
            } catch (error: any) {
                console.error('Login Failed:', error)
                const msg = error?.response?.data?.message ?? error?.message ?? t('auth.login_failed')
                setServerError(msg)
            }
        }
    })

    const labelClassName = cn(
        'text-xs font-semibold uppercase tracking-[0.18em]',
        isDark ? 'text-white/90' : 'text-slate-700'
    )

    const inputClassName = cn(
        'h-12 rounded-full px-5 shadow-none',
        isDark
            ? 'border-white/20 bg-white/12 text-white placeholder:text-white/45 focus-visible:border-amber-400 focus-visible:ring-amber-400/25'
            : 'border-slate-300/80 bg-white/80 text-slate-950 placeholder:text-slate-400 focus-visible:border-amber-500 focus-visible:ring-amber-500/20'
    )

    return (
        <section
            className={cn(
                'relative z-10 w-full max-w-md rounded-3xl border p-7 shadow-2xl backdrop-blur-md transition-colors duration-500 sm:p-10',
                isDark
                    ? 'border-white/15 bg-slate-950/50 shadow-black/45'
                    : 'border-white/70 bg-white/72 shadow-slate-900/20'
            )}
        >
            <header className="mb-9 text-center">
                <h1 className={cn(
                    'text-3xl font-semibold uppercase tracking-[0.22em]',
                    isDark ? 'text-white' : 'text-slate-950'
                )}>
                    {t('auth.sign_in')}
                </h1>
                <div className="mx-auto mt-4 h-px w-16 bg-amber-400" />
            </header>

            <form
                onSubmit={(event) => {
                    event.preventDefault()
                    event.stopPropagation()
                    form.handleSubmit()
                }}
                className="space-y-6"
            >
                <form.Field
                    name="email"
                    children={(field) => (
                        <div className="space-y-2">
                            <Label htmlFor={field.name} className={labelClassName}>
                                {t('auth.email')}
                            </Label>
                            <Input
                                id={field.name}
                                name={field.name}
                                value={field.state.value}
                                onBlur={field.handleBlur}
                                onChange={(event) => field.handleChange(event.target.value)}
                                placeholder={t('auth.email_placeholder')}
                                type="email"
                                autoComplete="email"
                                className={inputClassName}
                            />
                            {field.state.meta.errors.length > 0 && (
                                <em role="alert" className={cn('text-sm font-medium', isDark ? 'text-red-300' : 'text-red-700')}>
                                    {field.state.meta.errors.map((error) => typeof error === 'object' ? ((error as any).message ?? JSON.stringify(error)) : String(error)).join(', ')}
                                </em>
                            )}
                        </div>
                    )}
                />

                <form.Field
                    name="password"
                    children={(field) => (
                        <div className="space-y-2">
                            <Label htmlFor={field.name} className={labelClassName}>
                                {t('auth.password')}
                            </Label>
                            <PasswordInput
                                id={field.name}
                                name={field.name}
                                value={field.state.value}
                                onBlur={field.handleBlur}
                                onChange={(event) => field.handleChange(event.target.value)}
                                placeholder={t('auth.password_placeholder')}

                                autoComplete="current-password"
                                className={inputClassName}
                            />
                            {field.state.meta.errors.length > 0 && (
                                <em role="alert" className={cn('text-sm font-medium', isDark ? 'text-red-300' : 'text-red-700')}>
                                    {field.state.meta.errors.map((error) => typeof error === 'object' ? ((error as any).message ?? JSON.stringify(error)) : String(error)).join(', ')}
                                </em>
                            )}
                        </div>
                    )}
                />

                {serverError && (
                    <div className={cn(
                        'rounded-xl border p-3 text-center text-sm',
                        isDark
                            ? 'border-red-300/25 bg-red-950/45 text-red-200'
                            : 'border-red-200 bg-red-50/90 text-red-700'
                    )}>
                        {serverError}
                    </div>
                )}

                <form.Subscribe
                    selector={(state) => [state.canSubmit, state.isSubmitting]}
                    children={([canSubmit, isSubmitting]) => (
                        <div className="pt-2">
                            <Button
                                type="submit"
                                className={cn(
                                    'h-12 w-full rounded-full border bg-transparent text-sm font-semibold uppercase tracking-[0.16em] shadow-none transition-colors',
                                    isDark
                                        ? 'border-white/55 text-white hover:border-amber-400 hover:bg-amber-400 hover:text-slate-950'
                                        : 'border-slate-900/55 text-slate-950 hover:border-amber-500 hover:bg-amber-400 hover:text-slate-950'
                                )}
                                disabled={!canSubmit || isSubmitting || auth.isLoading}
                            >
                                {isSubmitting || auth.isLoading ? t('auth.signing_in') : t('auth.sign_in')}
                            </Button>
                        </div>
                    )}
                />
            </form>
        </section>
    )
}