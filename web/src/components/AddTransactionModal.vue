<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue'
import { api } from '../api/client'
import type { Category, TransactionType, ProblemDetails, PaymentMethod } from '../types'
import { formatIDR, PAYMENT_METHODS } from '../types'

const props = defineProps<{
  isOpen: boolean
  categories: Category[]
  defaultType?: TransactionType
  defaultMethod?: PaymentMethod
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'created'): void
}>()

const amountInputRef = ref<HTMLInputElement | null>(null)

// Stages: 'form' -> 'processing' -> 'receipt'
const stage = ref<'form' | 'processing' | 'receipt'>('form')

const txType = ref<TransactionType>('expense')
const rawAmount = ref<number | null>(null)
const categoryId = ref<number | ''>('')
const paymentMethod = ref<PaymentMethod>('qris')
const occurredOn = ref<string>(new Date().toISOString().split('T')[0])
const note = ref<string>('')

const isSubmitting = ref(false)
const errorMessage = ref<string | null>(null)
const fieldErrors = ref<Record<string, string>>({})

// Receipt snapshot data for the full-screen M-Banking animation
const receiptData = ref<{
  refId: string
  type: TransactionType
  amount: number
  categoryName: string
  methodName: string
  cardNumber: string
  date: string
  note: string
  timestamp: string
} | null>(null)

// Quick preset amounts (distinct for expense vs income)
const quickPresets = computed(() => {
  if (txType.value === 'expense') {
    return [25000, 50000, 100000, 250000, 500000]
  }
  return [500000, 1000000, 2500000, 5000000, 10000000]
})

// Filter categories matching current transaction type
const availableCategories = computed(() => {
  return props.categories.filter((c) => c.type === txType.value)
})

const selectedCategoryName = computed(() => {
  const found = availableCategories.value.find((c) => c.id === Number(categoryId.value))
  return found ? found.name : 'Kategori'
})

const selectedCardInfo = computed(() => {
  return PAYMENT_METHODS.find((m) => m.id === paymentMethod.value) || PAYMENT_METHODS[0]
})

// Auto select first category when switching type
watch(txType, () => {
  fieldErrors.value = {}
  errorMessage.value = null
  const matching = availableCategories.value
  if (matching.length > 0) {
    categoryId.value = matching[0].id
  } else {
    categoryId.value = ''
  }
})

// Initialize state when modal opens
watch(
  () => props.isOpen,
  (open) => {
    if (open) {
      stage.value = 'form'
      errorMessage.value = null
      fieldErrors.value = {}
      rawAmount.value = null
      note.value = ''
      txType.value = props.defaultType || 'expense'
      if (props.defaultMethod && props.defaultMethod !== 'all') {
        paymentMethod.value = props.defaultMethod
      } else {
        paymentMethod.value = 'qris'
      }
      occurredOn.value = new Date().toISOString().split('T')[0]
      const matching = availableCategories.value
      if (matching.length > 0) {
        categoryId.value = matching[0].id
      }
      nextTick(() => {
        amountInputRef.value?.focus()
      })
    }
  }
)

const formattedAmountDisplay = computed(() => {
  if (!rawAmount.value || rawAmount.value <= 0) return 'Rp 0'
  return formatIDR(rawAmount.value)
})

function onAmountInput(e: Event) {
  const target = e.target as HTMLInputElement
  const digitsOnly = target.value.replace(/\D/g, '')
  if (digitsOnly === '') {
    rawAmount.value = null
  } else {
    rawAmount.value = parseInt(digitsOnly, 10)
  }
  if (fieldErrors.value['amount_idr']) {
    delete fieldErrors.value['amount_idr']
  }
}

function selectPresetAmount(val: number) {
  rawAmount.value = val
  if (fieldErrors.value['amount_idr']) {
    delete fieldErrors.value['amount_idr']
  }
}

function switchTxType(newType: TransactionType) {
  if (txType.value === newType) return
  txType.value = newType
}

async function handleSubmit() {
  errorMessage.value = null
  fieldErrors.value = {}

  if (!rawAmount.value || rawAmount.value <= 0) {
    fieldErrors.value['amount_idr'] = 'Masukkan nominal lebih besar dari Rp 0'
    return
  }
  if (!categoryId.value) {
    fieldErrors.value['category_id'] = 'Pilih kategori terlebih dahulu'
    return
  }
  if (!occurredOn.value) {
    fieldErrors.value['occurred_on'] = 'Tanggal transaksi wajib diisi'
    return
  }

  isSubmitting.value = true
  stage.value = 'processing'

  const methodTag = `[${paymentMethod.value.toUpperCase()}]`
  const cleanUserNote = note.value.replace(/\[.*?\]/g, '').trim()
  const finalNote = cleanUserNote ? `${methodTag} ${cleanUserNote}` : methodTag

  const now = new Date()
  const refPrefix = txType.value === 'expense' ? 'FL-OUT' : 'FL-IN'
  const randomSuffix = Math.floor(100000 + Math.random() * 900000)

  try {
    const [created] = await Promise.all([
      api.createTransaction({
        category_id: Number(categoryId.value),
        type: txType.value,
        amount_idr: rawAmount.value,
        occurred_on: occurredOn.value,
        note: finalNote,
      }),
      // M-Banking authorization animation duration
      new Promise((resolve) => setTimeout(resolve, 750)),
    ])

    receiptData.value = {
      refId: `${refPrefix}-${created?.id || randomSuffix}`,
      type: txType.value,
      amount: rawAmount.value,
      categoryName: selectedCategoryName.value,
      methodName: selectedCardInfo.value.name,
      cardNumber: selectedCardInfo.value.cardNumber,
      date: occurredOn.value,
      note: cleanUserNote || selectedCategoryName.value,
      timestamp: now.toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit', second: '2-digit' }),
    }

    stage.value = 'receipt'
    emit('created')
  } catch (err: any) {
    stage.value = 'form'
    const prob = err as ProblemDetails
    if (prob.validation_errors && prob.validation_errors.length > 0) {
      for (const v of prob.validation_errors) {
        fieldErrors.value[v.field] = v.reason
      }
    } else {
      errorMessage.value = prob.detail || 'Gagal menyimpan transaksi'
    }
  } finally {
    isSubmitting.value = false
  }
}

function resetForAnother() {
  rawAmount.value = null
  note.value = ''
  errorMessage.value = null
  fieldErrors.value = {}
  stage.value = 'form'
  nextTick(() => {
    amountInputRef.value?.focus()
  })
}

function closeAll() {
  stage.value = 'form'
  emit('close')
}
</script>

<template>
  <!-- FULL-SCREEN M-BANKING TRANSACTION OVERLAY -->
  <Transition
    enter-active-class="transition-all duration-400 ease-[cubic-bezier(0.22,1,0.36,1)]"
    enter-from-class="opacity-0 translate-y-8 scale-[0.98]"
    enter-to-class="opacity-100 translate-y-0 scale-100"
    leave-active-class="transition-all duration-300 ease-[cubic-bezier(0.32,0.72,0,1)]"
    leave-from-class="opacity-100 translate-y-0 scale-100"
    leave-to-class="opacity-0 translate-y-8 scale-[0.98]"
  >
    <div
      v-if="isOpen"
      class="fixed inset-0 z-50 bg-[#07090e] text-slate-100 overflow-y-auto flex flex-col justify-between select-none"
    >
      <!-- Dynamic Ambient M-Banking Aurora (Changes between Expense Coral-Rose & Income Emerald-Teal) -->
      <div
        class="pointer-events-none fixed inset-x-0 top-0 h-96 transition-all duration-700 opacity-90"
        :style="{
          background:
            txType === 'expense'
              ? 'radial-gradient(ellipse 80% 60% at 50% -10%, rgba(244, 63, 94, 0.22), rgba(251, 146, 60, 0.08), transparent 75%)'
              : 'radial-gradient(ellipse 80% 60% at 50% -10%, rgba(16, 185, 129, 0.24), rgba(45, 212, 191, 0.1), transparent 75%)',
        }"
      ></div>

      <!-- ============================================================== -->
      <!-- STAGE 1: FULL-SCREEN M-BANKING TRANSACTION INPUT               -->
      <!-- ============================================================== -->
      <div v-if="stage === 'form'" class="relative z-10 w-full max-w-xl mx-auto px-4 sm:px-6 py-5 sm:py-8 flex-1 flex flex-col justify-between">
        <!-- Top M-Banking Navigation Bar -->
        <div>
          <div class="flex items-center justify-between pb-5 border-b border-white/[0.08]">
            <button
              type="button"
              @click="closeAll"
              class="tactile-btn w-10 h-10 rounded-full bg-white/[0.06] hover:bg-white/[0.12] border border-white/[0.1] flex items-center justify-center text-slate-300 hover:text-white"
              title="Kembali"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.2" d="M15 19l-7-7 7-7"/>
              </svg>
            </button>

            <div class="text-center">
              <div class="flex items-center justify-center space-x-1.5">
                <span
                  class="w-2 h-2 rounded-full transition-colors duration-300"
                  :class="txType === 'expense' ? 'bg-rose-500' : 'bg-emerald-400'"
                ></span>
                <span class="text-[10px] font-mono font-bold uppercase tracking-widest text-slate-400">
                  {{ txType === 'expense' ? 'Outbound Payment' : 'Inbound Deposit' }}
                </span>
              </div>
              <h2 class="text-sm sm:text-base font-extrabold text-white tracking-tight mt-0.5">
                {{ txType === 'expense' ? 'Transaksi Pengeluaran' : 'Pencatatan Pemasukan' }}
              </h2>
            </div>

            <button
              type="button"
              @click="closeAll"
              class="tactile-btn w-10 h-10 rounded-full bg-white/[0.06] hover:bg-white/[0.12] border border-white/[0.1] flex items-center justify-center text-slate-300 hover:text-white"
              title="Tutup"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.2" d="M6 18L18 6M6 6l12 12"/>
              </svg>
            </button>
          </div>

          <!-- Mode Switcher Pill (Pengeluaran vs Pemasukan — Clean Vector Icons, No Emojis) -->
          <div class="mt-5 grid grid-cols-2 gap-2 bg-white/[0.05] p-1.5 rounded-2xl border border-white/[0.08]">
            <button
              type="button"
              @click="switchTxType('expense')"
              :class="[
                'tactile-btn py-2.5 px-4 rounded-xl text-xs sm:text-sm font-bold flex items-center justify-center space-x-2 transition-all duration-300',
                txType === 'expense'
                  ? 'bg-gradient-to-r from-rose-500 to-orange-500 text-white shadow-[0_8px_20px_-4px_rgba(244,63,94,0.5)]'
                  : 'text-slate-400 hover:text-white'
              ]"
            >
              <svg class="w-4 h-4 stroke-[2.4] transition-transform duration-300" :class="txType === 'expense' ? '-translate-y-0.5 translate-x-0.5' : ''" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" d="M7 17L17 7M17 7H8M17 7V16"/>
              </svg>
              <span>Pengeluaran</span>
            </button>

            <button
              type="button"
              @click="switchTxType('income')"
              :class="[
                'tactile-btn py-2.5 px-4 rounded-xl text-xs sm:text-sm font-bold flex items-center justify-center space-x-2 transition-all duration-300',
                txType === 'income'
                  ? 'bg-gradient-to-r from-emerald-500 to-teal-500 text-slate-950 shadow-[0_8px_20px_-4px_rgba(16,185,129,0.5)]'
                  : 'text-slate-400 hover:text-white'
              ]"
            >
              <svg class="w-4 h-4 stroke-[2.4] transition-transform duration-300" :class="txType === 'income' ? 'translate-y-0.5 -translate-x-0.5' : ''" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" d="M17 7L7 17M7 17H16M7 17V8"/>
              </svg>
              <span>Pemasukan</span>
            </button>
          </div>

          <!-- M-Banking Flow Trajectory Banner (Visualizing Outbound vs Inbound Flow) -->
          <div class="mt-4 bg-white/[0.03] border border-white/[0.07] rounded-2xl px-4 py-3 flex items-center justify-between text-xs">
            <div class="flex items-center space-x-2.5">
              <div
                class="w-8 h-8 rounded-xl flex items-center justify-center font-bold"
                :class="txType === 'expense' ? 'bg-rose-500/15 text-rose-400 border border-rose-500/30' : 'bg-emerald-500/15 text-emerald-400 border border-emerald-500/30'"
              >
                <svg v-if="txType === 'expense'" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <rect x="2" y="5" width="20" height="14" rx="2"/>
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2 10h20"/>
                </svg>
                <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
                </svg>
              </div>
              <div>
                <div class="text-[10px] font-mono uppercase text-slate-400">
                  {{ txType === 'expense' ? 'Sumber Kartu' : 'Sumber Dana' }}
                </div>
                <div class="font-bold text-white text-xs">
                  {{ txType === 'expense' ? selectedCardInfo.shortName : selectedCategoryName }}
                </div>
              </div>
            </div>

            <!-- Animated Flow Dots -->
            <div class="flex items-center space-x-1.5 px-2">
              <span
                class="w-1.5 h-1.5 rounded-full animate-ping"
                :class="txType === 'expense' ? 'bg-rose-400' : 'bg-emerald-400'"
              ></span>
              <svg class="w-4 h-4 text-slate-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14 5l7 7m0 0l-7 7m7-7H3"/>
              </svg>
            </div>

            <div class="flex items-center space-x-2.5 text-right">
              <div>
                <div class="text-[10px] font-mono uppercase text-slate-400">
                  {{ txType === 'expense' ? 'Alokasi Kategori' : 'Masuk ke Kartu' }}
                </div>
                <div class="font-bold text-white text-xs">
                  {{ txType === 'expense' ? selectedCategoryName : selectedCardInfo.shortName }}
                </div>
              </div>
            </div>
          </div>

          <!-- Error Banner -->
          <div v-if="errorMessage" class="mt-4 bg-rose-500/15 border border-rose-500/30 text-rose-200 text-xs px-4 py-3 rounded-2xl">
            {{ errorMessage }}
          </div>

          <!-- Form Inputs -->
          <form @submit.prevent="handleSubmit" class="mt-5 space-y-5">
            <!-- Hero Amount Input (M-Banking Centerpiece) -->
            <div
              class="rounded-3xl p-5 border transition-all duration-300"
              :class="[
                fieldErrors['amount_idr']
                  ? 'bg-rose-950/20 border-rose-500/50'
                  : txType === 'expense'
                    ? 'bg-white/[0.04] border-white/[0.1] focus-within:border-rose-400/60 focus-within:bg-white/[0.06]'
                    : 'bg-white/[0.04] border-white/[0.1] focus-within:border-emerald-400/60 focus-within:bg-white/[0.06]'
              ]"
            >
              <div class="flex items-center justify-between mb-2">
                <label class="text-[11px] font-mono font-bold uppercase tracking-widest text-slate-400">
                  Nominal {{ txType === 'expense' ? 'Pengeluaran' : 'Pemasukan' }} (IDR)
                </label>
                <span
                  class="text-[10px] font-mono font-bold px-2 py-0.5 rounded-full"
                  :class="txType === 'expense' ? 'bg-rose-500/15 text-rose-300' : 'bg-emerald-500/15 text-emerald-300'"
                >
                  {{ formattedAmountDisplay }}
                </span>
              </div>

              <div class="relative flex items-center">
                <span
                  class="text-2xl sm:text-3xl font-black font-mono mr-2.5 select-none"
                  :class="txType === 'expense' ? 'text-rose-400' : 'text-emerald-400'"
                >
                  {{ txType === 'expense' ? '-Rp' : '+Rp' }}
                </span>
                <input
                  ref="amountInputRef"
                  type="text"
                  inputmode="numeric"
                  :value="rawAmount ? new Intl.NumberFormat('id-ID').format(rawAmount) : ''"
                  @input="onAmountInput"
                  placeholder="0"
                  class="w-full bg-transparent text-3xl sm:text-4xl font-black font-mono text-white placeholder:text-slate-600 focus:outline-none tracking-tight"
                />
              </div>

              <p v-if="fieldErrors['amount_idr']" class="text-xs text-rose-400 mt-2 font-medium">
                {{ fieldErrors['amount_idr'] }}
              </p>

              <!-- Quick M-Banking Nominal Chips -->
              <div class="mt-4 pt-3 border-t border-white/[0.07] flex items-center gap-2 overflow-x-auto pb-1">
                <button
                  v-for="preset in quickPresets"
                  :key="preset"
                  type="button"
                  @click="selectPresetAmount(preset)"
                  :class="[
                    'tactile-btn text-[11px] font-mono font-bold px-3 py-1.5 rounded-full border whitespace-nowrap transition-all',
                    rawAmount === preset
                      ? txType === 'expense'
                        ? 'bg-rose-500 text-white border-rose-400 shadow-sm'
                        : 'bg-emerald-400 text-slate-950 border-emerald-300 shadow-sm'
                      : 'bg-white/[0.05] hover:bg-white/[0.1] text-slate-300 border-white/[0.1]'
                  ]"
                >
                  {{ formatIDR(preset) }}
                </button>
              </div>
            </div>

            <!-- Payment Method / Card Selector (Crisp Vector Icons, No Emojis) -->
            <div>
              <label class="block text-[11px] font-mono font-bold uppercase tracking-widest text-slate-400 mb-2">
                Metode Kartu / Dompet
              </label>
              <div class="grid grid-cols-2 sm:grid-cols-4 gap-2.5">
                <button
                  v-for="m in PAYMENT_METHODS"
                  :key="m.id"
                  type="button"
                  @click="paymentMethod = m.id"
                  :class="[
                    'tactile-btn p-3 rounded-2xl text-xs font-bold border flex flex-col items-start justify-between space-y-2 transition-all',
                    paymentMethod === m.id
                      ? 'bg-white text-slate-950 border-white shadow-[0_10px_25px_rgba(255,255,255,0.15)] scale-[1.02]'
                      : 'bg-white/[0.04] text-slate-300 border-white/[0.09] hover:bg-white/[0.08] hover:text-white'
                  ]"
                >
                  <div class="flex items-center justify-between w-full">
                    <!-- Vector Icon per Payment Type -->
                    <span
                      class="w-7 h-7 rounded-lg flex items-center justify-center"
                      :style="{
                        backgroundColor: paymentMethod === m.id ? '#0f172a' : 'rgba(255,255,255,0.08)',
                        color: m.color,
                      }"
                    >
                      <!-- QR Icon -->
                      <svg v-if="m.iconType === 'qr'" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 3h6v6H3V3zm12 0h6v6h-6V3zM3 15h6v6H3v-6zm13 0h2v2h-2v-2zm3 3h2v3h-2v-3zm-3 3h2v-2h-2v2z"/>
                      </svg>
                      <!-- Cash / Banknote Icon -->
                      <svg v-else-if="m.iconType === 'cash'" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <rect x="2" y="6" width="20" height="12" rx="2" stroke-width="2"/>
                        <circle cx="12" cy="12" r="2.5" stroke-width="2"/>
                      </svg>
                      <!-- ATM Card Icon -->
                      <svg v-else-if="m.iconType === 'card'" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <rect x="2" y="5" width="20" height="14" rx="2" stroke-width="2"/>
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2 10h20"/>
                      </svg>
                      <!-- Bank Vault Icon -->
                      <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 10l9-6 9 6v2H3v-2zm2 4h2v5H5v-5zm6 0h2v5h-2v-5zm6 0h2v5h-2v-5zM3 20h18"/>
                      </svg>
                    </span>

                    <span class="text-[10px] font-mono opacity-60">{{ m.cardNumber.replace('•••• ', '') }}</span>
                  </div>
                  <span class="tracking-tight font-extrabold">{{ m.shortName }}</span>
                </button>
              </div>
            </div>

            <!-- Category & Date Grid -->
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <!-- Category Selector -->
              <div>
                <label class="block text-[11px] font-mono font-bold uppercase tracking-widest text-slate-400 mb-1.5">
                  Kategori ({{ txType === 'expense' ? 'Pengeluaran' : 'Pemasukan' }})
                </label>
                <select
                  v-model="categoryId"
                  class="w-full px-4 py-3 bg-[#111622] border border-white/[0.12] rounded-2xl text-sm font-semibold text-white focus:outline-none focus:border-emerald-400 transition-colors"
                >
                  <option value="" disabled class="bg-slate-900 text-slate-400">-- Pilih Kategori --</option>
                  <option
                    v-for="cat in availableCategories"
                    :key="cat.id"
                    :value="cat.id"
                    class="bg-slate-900 text-white"
                  >
                    {{ cat.name }}
                  </option>
                </select>
                <p v-if="fieldErrors['category_id']" class="text-xs text-rose-400 mt-1">
                  {{ fieldErrors['category_id'] }}
                </p>
              </div>

              <!-- Date Picker -->
              <div>
                <label class="block text-[11px] font-mono font-bold uppercase tracking-widest text-slate-400 mb-1.5">
                  Tanggal Transaksi
                </label>
                <input
                  v-model="occurredOn"
                  type="date"
                  class="w-full px-4 py-3 bg-[#111622] border border-white/[0.12] rounded-2xl text-sm font-semibold font-mono text-white focus:outline-none focus:border-emerald-400 transition-colors"
                />
                <p v-if="fieldErrors['occurred_on']" class="text-xs text-rose-400 mt-1">
                  {{ fieldErrors['occurred_on'] }}
                </p>
              </div>
            </div>

            <!-- Note / Keterangan -->
            <div>
              <label class="block text-[11px] font-mono font-bold uppercase tracking-widest text-slate-400 mb-1.5">
                Catatan / Berita Acara (Opsional)
              </label>
              <input
                v-model="note"
                type="text"
                maxlength="500"
                :placeholder="txType === 'expense' ? 'Contoh: Kopi & makan siang tim' : 'Contoh: Pembayaran invoice proyek'"
                class="w-full px-4 py-3 bg-[#111622] border border-white/[0.12] rounded-2xl text-sm font-medium text-white placeholder:text-slate-500 focus:outline-none focus:border-emerald-400 transition-colors"
              />
            </div>

            <!-- Submit M-Banking Action Bar -->
            <div class="pt-3 flex items-center space-x-3">
              <button
                type="button"
                @click="closeAll"
                class="tactile-btn px-5 py-3.5 rounded-2xl text-xs sm:text-sm font-bold text-slate-300 hover:text-white bg-white/[0.06] hover:bg-white/[0.1] border border-white/[0.1]"
              >
                Batal
              </button>

              <button
                type="submit"
                :disabled="isSubmitting"
                :class="[
                  'tactile-btn flex-1 py-3.5 px-6 rounded-2xl text-xs sm:text-sm font-black tracking-wide flex items-center justify-center space-x-2 shadow-xl',
                  txType === 'expense'
                    ? 'bg-gradient-to-r from-rose-500 via-rose-500 to-orange-500 text-white shadow-rose-500/25'
                    : 'bg-gradient-to-r from-emerald-400 via-teal-400 to-emerald-500 text-slate-950 shadow-emerald-500/25'
                ]"
              >
                <span>{{ txType === 'expense' ? 'Otorisasi Pengeluaran' : 'Konfirmasi Pemasukan' }}</span>
                <svg v-if="txType === 'expense'" class="w-4 h-4 stroke-[2.5]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M7 17L17 7M17 7H8M17 7V16"/>
                </svg>
                <svg v-else class="w-4 h-4 stroke-[2.5]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7"/>
                </svg>
              </button>
            </div>
          </form>
        </div>
      </div>

      <!-- ============================================================== -->
      <!-- STAGE 2: FULL-SCREEN M-BANKING AUTHORIZATION ANIMATION         -->
      <!-- (Distinct Motion for Pengeluaran Outbound vs Pemasukan Inbound) -->
      <!-- ============================================================== -->
      <div
        v-else-if="stage === 'processing'"
        class="relative z-10 flex-1 flex flex-col items-center justify-center px-6 text-center"
      >
        <!-- EXPENSE OUTBOUND ANIMATION (Shooting Upward Waves) -->
        <div v-if="txType === 'expense'" class="relative w-36 h-36 flex items-center justify-center mb-8">
          <div class="absolute inset-0 rounded-full border-2 border-rose-500/40 animate-outbound-wave-1"></div>
          <div class="absolute inset-2 rounded-full border-2 border-orange-400/30 animate-outbound-wave-2"></div>
          <div class="w-24 h-24 rounded-full bg-gradient-to-tr from-rose-500 to-orange-500 flex items-center justify-center shadow-[0_0_50px_rgba(244,63,94,0.55)] animate-outbound-flight">
            <svg class="w-10 h-10 text-white stroke-[2.5]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" d="M7 17L17 7M17 7H8M17 7V16"/>
            </svg>
          </div>
        </div>

        <!-- INCOME INBOUND ANIMATION (Converging Downward Vault Waves) -->
        <div v-else class="relative w-36 h-36 flex items-center justify-center mb-8">
          <div class="absolute inset-0 rounded-full border-2 border-emerald-400/40 animate-inbound-wave-1"></div>
          <div class="absolute inset-2 rounded-full border-2 border-teal-300/30 animate-inbound-wave-2"></div>
          <div class="w-24 h-24 rounded-full bg-gradient-to-tr from-emerald-500 to-teal-400 flex items-center justify-center shadow-[0_0_50px_rgba(16,185,129,0.55)] animate-inbound-drop">
            <svg class="w-10 h-10 text-slate-950 stroke-[2.5]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" d="M17 7L7 17M7 17H16M7 17V8"/>
            </svg>
          </div>
        </div>

        <div class="text-xs font-mono font-bold uppercase tracking-widest" :class="txType === 'expense' ? 'text-rose-400' : 'text-emerald-400'">
          {{ txType === 'expense' ? 'Mengotorisasi Pembayaran Keluar' : 'Mengkreditkan Dana Masuk' }}
        </div>
        <div class="text-2xl sm:text-3xl font-black font-mono text-white mt-2">
          {{ formattedAmountDisplay }}
        </div>
        <p class="text-xs text-slate-400 mt-2">
          Menyinkronkan jurnal transaksi ke SQLite WAL Engine...
        </p>
      </div>

      <!-- ============================================================== -->
      <!-- STAGE 3: FULL-SCREEN M-BANKING DIGITAL RECEIPT                 -->
      <!-- (Distinct Visual Signature for Expense vs Income)              -->
      <!-- ============================================================== -->
      <div
        v-else-if="stage === 'receipt' && receiptData"
        class="relative z-10 w-full max-w-md mx-auto px-4 sm:px-6 py-8 flex-1 flex flex-col items-center justify-center"
      >
        <!-- Animated Status Seal -->
        <div class="relative mb-5">
          <!-- Expense Outbound Seal -->
          <div
            v-if="receiptData.type === 'expense'"
            class="w-20 h-20 rounded-full bg-gradient-to-tr from-rose-500 to-orange-400 flex items-center justify-center shadow-[0_12px_40px_rgba(244,63,94,0.5)] animate-outbound-flight"
          >
            <svg class="w-9 h-9 text-white stroke-[2.6]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" d="M7 17L17 7M17 7H8M17 7V16"/>
            </svg>
          </div>

          <!-- Income Inbound Seal -->
          <div
            v-else
            class="w-20 h-20 rounded-full bg-gradient-to-tr from-emerald-400 to-teal-300 flex items-center justify-center shadow-[0_12px_40px_rgba(16,185,129,0.5)] animate-inbound-drop"
          >
            <svg class="w-9 h-9 text-slate-950 stroke-[2.8]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7"/>
            </svg>
          </div>
        </div>

        <!-- Status Badge -->
        <span
          class="text-[10px] font-mono font-black uppercase tracking-widest px-3.5 py-1 rounded-full border mb-2"
          :class="
            receiptData.type === 'expense'
              ? 'bg-rose-500/15 text-rose-300 border-rose-500/30'
              : 'bg-emerald-500/15 text-emerald-300 border-emerald-500/30'
          "
        >
          {{ receiptData.type === 'expense' ? 'Pembayaran Keluar Berhasil' : 'Dana Masuk Berhasil Diterima' }}
        </span>

        <!-- Hero Receipt Amount -->
        <div
          class="text-3xl sm:text-4xl font-black font-mono tracking-tight my-1"
          :class="receiptData.type === 'expense' ? 'text-white' : 'text-emerald-400'"
        >
          {{ receiptData.type === 'expense' ? '- ' : '+ ' }}{{ formatIDR(receiptData.amount) }}
        </div>
        <div class="text-xs text-slate-400 font-mono mb-6">
          {{ receiptData.date }} • {{ receiptData.timestamp }} WIB
        </div>

        <!-- M-Banking Receipt Card -->
        <div class="w-full bg-[#101520] border border-white/[0.1] rounded-3xl p-5 sm:p-6 shadow-2xl space-y-3.5 text-xs">
          <div class="flex items-center justify-between pb-3 border-b border-dashed border-white/[0.1]">
            <span class="text-slate-400 font-mono">No. Referensi</span>
            <span class="font-mono font-bold text-white">{{ receiptData.refId }}</span>
          </div>

          <div class="flex items-center justify-between">
            <span class="text-slate-400">Jenis Aliran Kas</span>
            <span
              class="font-bold"
              :class="receiptData.type === 'expense' ? 'text-rose-400' : 'text-emerald-400'"
            >
              {{ receiptData.type === 'expense' ? 'Debit / Pengeluaran (↗)' : 'Kredit / Pemasukan (↙)' }}
            </span>
          </div>

          <div class="flex items-center justify-between">
            <span class="text-slate-400">Metode Kartu</span>
            <span class="font-bold text-white font-mono">
              {{ receiptData.methodName }} ({{ receiptData.cardNumber }})
            </span>
          </div>

          <div class="flex items-center justify-between">
            <span class="text-slate-400">Kategori</span>
            <span class="font-bold text-white">{{ receiptData.categoryName }}</span>
          </div>

          <div class="flex items-center justify-between pt-3 border-t border-white/[0.08]">
            <span class="text-slate-400">Catatan</span>
            <span class="font-semibold text-slate-200 text-right max-w-[200px] truncate">{{ receiptData.note }}</span>
          </div>
        </div>

        <!-- Receipt Actions -->
        <div class="w-full mt-6 flex items-center space-x-3">
          <button
            type="button"
            @click="resetForAnother"
            class="tactile-btn flex-1 py-3.5 px-4 rounded-2xl text-xs font-bold text-slate-200 bg-white/[0.06] hover:bg-white/[0.12] border border-white/[0.1]"
          >
            Catat Transaksi Lain
          </button>
          <button
            type="button"
            @click="closeAll"
            class="tactile-btn flex-1 py-3.5 px-4 rounded-2xl text-xs font-black text-slate-950 bg-white hover:bg-slate-100 shadow-lg"
          >
            Selesai
          </button>
        </div>
      </div>
    </div>
  </Transition>
</template>
