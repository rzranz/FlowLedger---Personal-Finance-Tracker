<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue'
import { api } from '../api/client'
import type { Category, TransactionType, ProblemDetails, PaymentMethod } from '../types'
import { formatIDR, PAYMENT_METHODS } from '../types'

const props = defineProps<{
  isOpen: boolean
  categories: Category[]
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'created'): void
}>()

const amountInputRef = ref<HTMLInputElement | null>(null)

const txType = ref<TransactionType>('expense')
const rawAmount = ref<number | null>(null)
const categoryId = ref<number | ''>('')
const paymentMethod = ref<PaymentMethod>('qris')
const occurredOn = ref<string>(new Date().toISOString().split('T')[0])
const note = ref<string>('')

const isSubmitting = ref(false)
const errorMessage = ref<string | null>(null)
const fieldErrors = ref<Record<string, string>>({})

// Filter categories matching current transaction type
const availableCategories = computed(() => {
  return props.categories.filter((c) => c.type === txType.value)
})

// Auto select first category when switching type
watch(txType, () => {
  const matching = availableCategories.value
  if (matching.length > 0) {
    categoryId.value = matching[0].id
  } else {
    categoryId.value = ''
  }
})

// Focus amount input when modal opens
watch(
  () => props.isOpen,
  (open) => {
    if (open) {
      errorMessage.value = null
      fieldErrors.value = {}
      rawAmount.value = null
      note.value = ''
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
}

async function handleSubmit() {
  errorMessage.value = null
  fieldErrors.value = {}

  if (!rawAmount.value || rawAmount.value <= 0) {
    fieldErrors.value['amount_idr'] = 'Nominal harus lebih besar dari 0'
    return
  }
  if (!categoryId.value) {
    fieldErrors.value['category_id'] = 'Pilih kategori transaksi'
    return
  }
  if (!occurredOn.value) {
    fieldErrors.value['occurred_on'] = 'Tanggal transaksi wajib diisi'
    return
  }

  isSubmitting.value = true
  const methodTag = `[${paymentMethod.value.toUpperCase()}]`
  const cleanUserNote = note.value.replace(/\[.*?\]/g, '').trim()
  const finalNote = cleanUserNote ? `${methodTag} ${cleanUserNote}` : methodTag

  try {
    await api.createTransaction({
      category_id: Number(categoryId.value),
      type: txType.value,
      amount_idr: rawAmount.value,
      occurred_on: occurredOn.value,
      note: finalNote,
    })
    emit('created')
    emit('close')
  } catch (err: any) {
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
</script>

<template>
  <div
    v-if="isOpen"
    class="fixed inset-0 z-50 overflow-y-auto bg-slate-900/50 backdrop-blur-sm flex items-center justify-center p-4"
    @click.self="emit('close')"
  >
    <div class="bg-white rounded-2xl shadow-xl border border-slate-200 max-w-lg w-full overflow-hidden transition-all animate-in fade-in zoom-in-95 duration-150">
      <!-- Modal Header -->
      <div class="px-6 py-4 border-b border-slate-100 flex items-center justify-between bg-slate-50/50">
        <div class="flex items-center space-x-2">
          <span class="w-2.5 h-2.5 rounded-full bg-emerald-500"></span>
          <h3 class="text-base font-bold text-slate-900">Catat Transaksi Baru</h3>
        </div>
        <button
          @click="emit('close')"
          class="text-slate-400 hover:text-slate-600 p-1 rounded-lg hover:bg-slate-100"
        >
          ✕
        </button>
      </div>

      <!-- Form Body -->
      <form @submit.prevent="handleSubmit" class="p-6 space-y-4">
        <!-- Error Alert -->
        <div v-if="errorMessage" class="bg-rose-50 border border-rose-200 text-rose-700 text-sm px-4 py-3 rounded-xl">
          {{ errorMessage }}
        </div>

        <!-- Type Selector (Pengeluaran / Pemasukan) -->
        <div class="grid grid-cols-2 gap-2 bg-slate-100 p-1 rounded-xl">
          <button
            type="button"
            @click="txType = 'expense'"
            :class="[
              'py-2 text-sm font-semibold rounded-lg transition-all',
              txType === 'expense'
                ? 'bg-rose-500 text-white shadow-sm'
                : 'text-slate-600 hover:text-slate-900'
            ]"
          >
            💸 Pengeluaran
          </button>
          <button
            type="button"
            @click="txType = 'income'"
            :class="[
              'py-2 text-sm font-semibold rounded-lg transition-all',
              txType === 'income'
                ? 'bg-emerald-600 text-white shadow-sm'
                : 'text-slate-600 hover:text-slate-900'
            ]"
          >
            💰 Pemasukan
          </button>
        </div>

        <!-- Amount IDR Input with live formatting -->
        <div>
          <label class="block text-xs font-bold uppercase tracking-wider text-slate-500 mb-1">
            Nominal (Rupiah)
          </label>
          <div class="relative rounded-xl shadow-sm">
            <input
              ref="amountInputRef"
              type="text"
              inputmode="numeric"
              :value="rawAmount ? new Intl.NumberFormat('id-ID').format(rawAmount) : ''"
              @input="onAmountInput"
              placeholder="Contoh: 50.000"
              class="w-full text-2xl font-bold font-mono px-4 py-3 border border-slate-300 rounded-xl focus:ring-2 focus:ring-emerald-500 focus:border-emerald-500 transition-colors"
              :class="fieldErrors['amount_idr'] ? 'border-rose-400 bg-rose-50/20' : ''"
            />
            <div class="absolute inset-y-0 right-0 pr-4 flex items-center pointer-events-none text-xs font-semibold text-slate-400">
              IDR
            </div>
          </div>
          <p v-if="fieldErrors['amount_idr']" class="text-xs text-rose-600 mt-1">
            {{ fieldErrors['amount_idr'] }}
          </p>
          <p v-else-if="rawAmount" class="text-xs text-emerald-600 mt-1 font-semibold">
            Terbaca: {{ formattedAmountDisplay }}
          </p>
        </div>

        <!-- Category Selector -->
        <div>
          <label class="block text-xs font-bold uppercase tracking-wider text-slate-500 mb-1">
            Kategori ({{ txType === 'expense' ? 'Pengeluaran' : 'Pemasukan' }})
          </label>
          <select
            v-model="categoryId"
            class="w-full px-3.5 py-2.5 border border-slate-300 rounded-xl text-sm font-medium bg-white focus:ring-2 focus:ring-emerald-500 focus:border-emerald-500"
            :class="fieldErrors['category_id'] ? 'border-rose-400' : ''"
          >
            <option value="" disabled>-- Pilih Kategori --</option>
            <option v-for="cat in availableCategories" :key="cat.id" :value="cat.id">
              {{ cat.name }}
            </option>
          </select>
          <p v-if="fieldErrors['category_id']" class="text-xs text-rose-600 mt-1">
            {{ fieldErrors['category_id'] }}
          </p>
        </div>

        <!-- Payment Method Selector (QRIS, Tunai, ATM, Transfer) -->
        <div>
          <label class="block text-xs font-bold uppercase tracking-wider text-slate-500 mb-1">
            Metode Transaksi / Kartu
          </label>
          <div class="grid grid-cols-2 sm:grid-cols-4 gap-2">
            <button
              v-for="m in PAYMENT_METHODS"
              :key="m.id"
              type="button"
              @click="paymentMethod = m.id"
              :class="[
                'px-2.5 py-2 rounded-xl text-xs font-bold border transition-all flex items-center justify-center space-x-1.5',
                paymentMethod === m.id
                  ? 'bg-slate-900 text-white border-slate-900 shadow-md ring-2 ring-emerald-500'
                  : 'bg-slate-50 text-slate-700 border-slate-200 hover:bg-slate-100'
              ]"
            >
              <span>{{ m.icon }}</span>
              <span>{{ m.shortName }}</span>
            </button>
          </div>
        </div>

        <!-- Date Picker -->
        <div>
          <label class="block text-xs font-bold uppercase tracking-wider text-slate-500 mb-1">
            Tanggal Transaksi
          </label>
          <input
            v-model="occurredOn"
            type="date"
            class="w-full px-3.5 py-2.5 border border-slate-300 rounded-xl text-sm font-medium bg-white focus:ring-2 focus:ring-emerald-500 focus:border-emerald-500"
            :class="fieldErrors['occurred_on'] ? 'border-rose-400' : ''"
          />
          <p v-if="fieldErrors['occurred_on']" class="text-xs text-rose-600 mt-1">
            {{ fieldErrors['occurred_on'] }}
          </p>
        </div>

        <!-- Note / Keterangan -->
        <div>
          <label class="block text-xs font-bold uppercase tracking-wider text-slate-500 mb-1">
            Catatan (Opsional)
          </label>
          <input
            v-model="note"
            type="text"
            maxlength="500"
            placeholder="Contoh: Makan siang bersama tim"
            class="w-full px-3.5 py-2.5 border border-slate-300 rounded-xl text-sm focus:ring-2 focus:ring-emerald-500 focus:border-emerald-500"
          />
        </div>

        <!-- Footer Actions -->
        <div class="pt-4 border-t border-slate-100 flex items-center justify-end space-x-3">
          <button
            type="button"
            @click="emit('close')"
            class="px-4 py-2 text-sm font-semibold text-slate-600 hover:text-slate-800 hover:bg-slate-100 rounded-xl transition-colors"
          >
            Batal
          </button>
          <button
            type="submit"
            :disabled="isSubmitting"
            class="px-5 py-2 text-sm font-semibold text-white bg-emerald-600 hover:bg-emerald-700 disabled:opacity-50 rounded-xl shadow-sm transition-all"
          >
            {{ isSubmitting ? 'Menyimpan...' : 'Simpan Transaksi' }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

