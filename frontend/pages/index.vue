<script setup lang="ts">
import {
  formatCardNumber,
  formatIban,
  useCardToIban,
} from '~/composables/useCardToIban'

const card = ref('')
const copied = ref(false)
const { result, errorMessage, isLoading, convert, reset } = useCardToIban()

const cardDisplayValue = computed({
  get: () => formatCardNumber(card.value),
  set: (value: string) => {
    card.value = value
  },
})

async function submit() {
  copied.value = false
  await convert(card.value)
}

function clearForm() {
  card.value = ''
  reset()
}

async function copyIban() {
  if (!result.value?.iban || !navigator.clipboard) return

  await navigator.clipboard.writeText(result.value.iban)
  copied.value = true
  window.setTimeout(() => {
    copied.value = false
  }, 2200)
}
</script>

<template>
  <div class="relative min-h-screen overflow-hidden">
    <div class="mesh pointer-events-none absolute inset-0 h-[34rem] opacity-70" />

    <header class="relative z-10 border-b border-emerald-950/5 bg-white/65 backdrop-blur-xl">
      <div class="mx-auto flex max-w-7xl items-center justify-between px-5 py-5 sm:px-8 lg:px-10">
        <NuxtLink to="/" class="flex items-center gap-3" aria-label="شبایار">
          <span class="grid h-11 w-11 place-items-center rounded-2xl bg-ink text-white shadow-card">
            <svg viewBox="0 0 24 24" class="h-6 w-6" fill="none" aria-hidden="true">
              <path d="M5 8.5h14M5 15.5h14M8 5v14M16 5v14" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" />
              <circle cx="12" cy="12" r="9" stroke="#67e8b0" stroke-width="1.5" />
            </svg>
          </span>
          <span>
            <span class="block text-lg font-extrabold tracking-tight text-ink">شبایار</span>
            <span class="block text-[11px] font-medium text-emerald-700">ساده، امن، شفاف</span>
          </span>
        </NuxtLink>

        <div class="hidden items-center gap-2 text-sm text-slate-500 sm:flex">
          <span class="h-2 w-2 rounded-full bg-emerald-500" />
          سرویس آنلاین است
        </div>
      </div>
    </header>

    <main class="relative z-10 mx-auto max-w-7xl px-5 pb-16 pt-10 sm:px-8 sm:pt-16 lg:px-10 lg:pt-20">
      <div class="grid items-center gap-12 lg:grid-cols-[1.05fr_0.95fr] lg:gap-20">
        <section class="order-2 lg:order-1">
          <div class="mb-6 inline-flex items-center gap-2 rounded-full border border-emerald-200 bg-emerald-50 px-3 py-1.5 text-xs font-semibold text-emerald-700">
            <svg viewBox="0 0 20 20" class="h-4 w-4" fill="currentColor" aria-hidden="true">
              <path fill-rule="evenodd" d="M10 1.667a6.667 6.667 0 0 1 6.667 6.666c0 4.667-6.667 10-6.667 10s-6.667-5.333-6.667-10A6.667 6.667 0 0 1 10 1.667Zm0 9.166a2.5 2.5 0 1 0 0-5 2.5 2.5 0 0 0 0 5Z" clip-rule="evenodd" />
            </svg>
            استعلام امن و سریع
          </div>

          <h1 class="max-w-xl text-4xl font-extrabold leading-[1.3] tracking-tight text-ink sm:text-5xl lg:text-[4rem]">
            شماره کارتت را به
            <span class="relative mx-2 inline-block text-emerald-600">
              شبا
              <span class="absolute -bottom-1 right-0 h-2 w-full rounded-full bg-gold/25" />
            </span>
            تبدیل کن.
          </h1>
          <p class="mt-6 max-w-lg text-base leading-8 text-slate-500 sm:text-lg">
            فقط شماره کارت را وارد کن؛ شبای حساب را از مسیر رسمی و قابل اعتماد دریافت کن.
            اطلاعات کارتت در این سرویس ذخیره نمی‌شود.
          </p>

          <div class="mt-9 grid max-w-lg grid-cols-3 gap-3 sm:gap-5">
            <div class="rounded-2xl border border-emerald-950/5 bg-white/75 p-4 shadow-sm">
              <span class="mb-3 grid h-9 w-9 place-items-center rounded-xl bg-emerald-100 text-emerald-700">
                <svg viewBox="0 0 24 24" class="h-5 w-5" fill="none" aria-hidden="true"><path d="M12 3 5 6v5c0 4.5 2.9 8.3 7 10 4.1-1.7 7-5.5 7-10V6l-7-3Z" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round" /><path d="m9 12 2 2 4-4" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" /></svg>
              </span>
              <strong class="block text-sm text-ink">امن</strong>
              <span class="mt-1 block text-[11px] text-slate-400">حفاظت از داده</span>
            </div>
            <div class="rounded-2xl border border-emerald-950/5 bg-white/75 p-4 shadow-sm">
              <span class="mb-3 grid h-9 w-9 place-items-center rounded-xl bg-amber-100 text-amber-700">
                <svg viewBox="0 0 24 24" class="h-5 w-5" fill="none" aria-hidden="true"><path d="M13 2 4 14h7l-1 8 9-12h-7l1-8Z" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round" /></svg>
              </span>
              <strong class="block text-sm text-ink">سریع</strong>
              <span class="mt-1 block text-[11px] text-slate-400">نتیجه در چند ثانیه</span>
            </div>
            <div class="rounded-2xl border border-emerald-950/5 bg-white/75 p-4 shadow-sm">
              <span class="mb-3 grid h-9 w-9 place-items-center rounded-xl bg-sky-100 text-sky-700">
                <svg viewBox="0 0 24 24" class="h-5 w-5" fill="none" aria-hidden="true"><path d="M4 18V9m5 9V5m5 13v-6m5 6V3" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" /></svg>
              </span>
              <strong class="block text-sm text-ink">شفاف</strong>
              <span class="mt-1 block text-[11px] text-slate-400">نمایش کامل نتیجه</span>
            </div>
          </div>
        </section>

        <section class="order-1 lg:order-2">
          <div class="glass relative overflow-hidden rounded-[2rem] border border-white p-6 shadow-soft sm:p-8">
            <div class="absolute -left-16 -top-16 h-44 w-44 rounded-full bg-emerald-100/80 blur-2xl" />
            <div class="relative">
              <div class="flex items-start justify-between gap-4">
                <div>
                  <p class="text-xs font-bold uppercase tracking-[0.18em] text-emerald-600">Card to IBAN</p>
                  <h2 class="mt-2 text-2xl font-extrabold text-ink">تبدیل شماره کارت</h2>
                </div>
                <span class="grid h-12 w-12 place-items-center rounded-2xl bg-ink text-emerald-200">
                  <svg viewBox="0 0 24 24" class="h-6 w-6" fill="none" aria-hidden="true"><rect x="3" y="5" width="18" height="14" rx="2.5" stroke="currentColor" stroke-width="1.8" /><path d="M3 10h18M7 15h3" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" /></svg>
                </span>
              </div>

              <form class="mt-8" @submit.prevent="submit">
                <label for="card-number" class="mb-2 block text-sm font-semibold text-ink">شماره کارت</label>
                <div class="relative">
                  <input
                    id="card-number"
                    v-model="cardDisplayValue"
                    dir="ltr"
                    inputmode="numeric"
                    autocomplete="off"
                    maxlength="19"
                    placeholder="۶۰۳۷ ۹۹۷۵ ۱۲۳۴ ۵۶۷۰"
                    class="w-full rounded-2xl border border-slate-200 bg-white px-4 py-4 text-center text-lg tracking-[0.14em] text-ink shadow-sm transition placeholder:text-slate-300 focus:border-emerald-500 focus:ring-4 focus:ring-emerald-100"
                    :disabled="isLoading"
                    aria-describedby="card-hint"
                  />
                  <span class="pointer-events-none absolute left-4 top-1/2 -translate-y-1/2 text-slate-300">
                    <svg viewBox="0 0 24 24" class="h-5 w-5" fill="none" aria-hidden="true"><path d="M4 7.5h16v9H4zM4 10.5h16M7 14h3" stroke="currentColor" stroke-width="1.6" stroke-linejoin="round" /></svg>
                  </span>
                </div>
                <p id="card-hint" class="mt-2 text-xs text-slate-400">شماره کارت ۱۶ رقمی را بدون نگرانی وارد کن.</p>

                <div v-if="errorMessage" role="alert" class="mt-5 flex items-start gap-3 rounded-2xl border border-rose-100 bg-rose-50 p-4 text-sm text-rose-700">
                  <svg viewBox="0 0 24 24" class="mt-0.5 h-5 w-5 shrink-0" fill="none" aria-hidden="true"><circle cx="12" cy="12" r="9" stroke="currentColor" stroke-width="1.8" /><path d="M12 8v5M12 16h.01" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" /></svg>
                  <span>{{ errorMessage }}</span>
                </div>

                <button
                  type="submit"
                  class="mt-6 flex w-full items-center justify-center gap-3 rounded-2xl bg-emerald-600 px-5 py-4 font-bold text-white shadow-lg shadow-emerald-600/20 transition hover:bg-emerald-700 disabled:cursor-not-allowed disabled:opacity-70"
                  :disabled="isLoading"
                >
                  <span v-if="isLoading" class="h-5 w-5 animate-spin rounded-full border-2 border-white/30 border-t-white" />
                  <span>{{ isLoading ? 'در حال استعلام...' : 'دریافت شماره شبا' }}</span>
                  <svg v-if="!isLoading" viewBox="0 0 24 24" class="h-5 w-5" fill="none" aria-hidden="true"><path d="M5 12h14M13 6l6 6-6 6" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" /></svg>
                </button>
              </form>

              <div v-if="result" class="mt-6 rounded-2xl border border-emerald-200 bg-emerald-50/80 p-5" aria-live="polite">
                <div class="flex items-center justify-between gap-3">
                  <div class="flex items-center gap-2 text-sm font-bold text-emerald-800">
                    <span class="grid h-7 w-7 place-items-center rounded-full bg-emerald-600 text-white">
                      <svg viewBox="0 0 24 24" class="h-4 w-4" fill="none" aria-hidden="true"><path d="m6 12 4 4 8-8" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" /></svg>
                    </span>
                    استعلام با موفقیت انجام شد
                  </div>
                  <span class="text-xs text-emerald-700">{{ result.bankName }}</span>
                </div>
                <p class="mt-4 text-xs font-medium text-emerald-700">شماره شبا</p>
                <div class="mt-1 flex items-center justify-between gap-3 rounded-xl bg-white px-3 py-3" dir="ltr">
                  <span class="min-w-0 truncate text-sm font-bold tracking-wider text-ink">{{ formatIban(result.iban) }}</span>
                  <button type="button" class="shrink-0 rounded-lg p-2 text-emerald-700 transition hover:bg-emerald-100" :aria-label="copied ? 'کپی شد' : 'کپی شماره شبا'" @click="copyIban">
                    <svg v-if="!copied" viewBox="0 0 24 24" class="h-5 w-5" fill="none" aria-hidden="true"><rect x="8" y="8" width="11" height="11" rx="2" stroke="currentColor" stroke-width="1.7" /><path d="M16 8V6a2 2 0 0 0-2-2H6a2 2 0 0 0-2 2v8a2 2 0 0 0 2 2h2" stroke="currentColor" stroke-width="1.7" /></svg>
                    <svg v-else viewBox="0 0 24 24" class="h-5 w-5" fill="none" aria-hidden="true"><path d="m5 12 4 4L19 6" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" /></svg>
                  </button>
                </div>
                <p v-if="copied" class="mt-2 text-xs text-emerald-700">شماره شبا کپی شد.</p>
                <button type="button" class="mt-4 text-xs font-semibold text-slate-500 underline decoration-slate-300 underline-offset-4 transition hover:text-ink" @click="clearForm">استعلام جدید</button>
              </div>

              <div class="mt-6 flex items-center justify-center gap-2 text-xs text-slate-400">
                <svg viewBox="0 0 24 24" class="h-4 w-4" fill="none" aria-hidden="true"><path d="M12 3 5 6v5c0 4.5 2.9 8.3 7 10 4.1-1.7 7-5.5 7-10V6l-7-3Z" stroke="currentColor" stroke-width="1.6" stroke-linejoin="round" /><path d="m9 12 2 2 4-4" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round" /></svg>
                اطلاعات کارت شما در پاسخ یا رابط کاربری نمایش داده نمی‌شود
              </div>
            </div>
          </div>
        </section>
      </div>
    </main>

    <footer class="relative z-10 border-t border-emerald-950/5 bg-white/45">
      <div class="mx-auto flex max-w-7xl flex-col gap-2 px-5 py-6 text-xs text-slate-400 sm:flex-row sm:items-center sm:justify-between sm:px-8 lg:px-10">
        <span>© {{ new Date().getFullYear() }} شبایار</span>
        <span>استعلام بانکی با حفظ حریم خصوصی</span>
      </div>
    </footer>
  </div>
</template>
