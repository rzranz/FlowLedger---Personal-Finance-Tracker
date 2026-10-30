<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { api } from '../api/client'
import type { Category, ProblemDetails } from '../types'
import { formatIDR } from '../types'

const props = defineProps<{
  isOpen: boolean
  categories: Category[]
  initialMonth: string
  selectedCategoryId?: number
  initialLimit?: number
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'saved'): void
}>()

const month = ref(props.initialMonth)
const categoryId = ref<number | ''>(props.selectedCategoryId || '')
const rawLimit = ref<number | null>(props.initialLimit || null)

const isSubmitting = ref(false)
const errorMessage = ref<string | null>(null)

const expenseCategories = computed(() => {
  return props.categories.filter((c) => c.type === 'expense')
})

watch(
  () => props.isOpen,
  (open) => {
    if (open) {
      month.value = props.initialMonth
      categoryId.value = props.selectedCategoryId || (expenseCategories.value[0]?.id || '')
      rawLimit.value = props.initialLimit || null
      errorMessage.value = null
    }
  }
)

function onLimitInput(e: Event) {
  const target = e.target as HTMLInputElement
  const digitsOnly = target.value.replace(/\D/g, '')
  rawLimit.value = digitsOnly ? parseInt(digitsOnly, 10) : null
}

async function handleSave() {
  if (!categoryId.value || !rawLimit.value || rawLimit.value <= 0) {
    errorMessage.value = 'Pilih kategori dan masukkan batas nominal lebih dari 0'
    return
  }

  isSubmitting.value = true
  errorMessage.value = null

  try {
    await api.upsertBudget(Number(categoryId.value), month.value, rawLimit.value)
    emit('saved')
    emit('close')
  } catch (err: any) {
    const prob = err as ProblemDetails
    errorMessage.value = prob.detail || 'Gagal menyimpan anggaran'
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
    <div class="bg-white rounded-2xl shadow-xl border border-slate-200 max-w-md w-full overflow-hidden animate-in fade-in zoom-in-95 duration-150">
      <div class="px-6 py-4 border-b border-slate-100 flex items-center justify-between bg-slate-50/50">
        <h3 class="text-base font-bold text-slate-900">Pasang Batas Anggaran (Budget)</h3>
        <button @click="emit('close')" class="text-slate-400 hover:text-slate-600 p-1">✕</button>
      </div>

      <form @submit.prevent="handleSave" class="p-6 space-y-4">
        <div v-if="errorMessage" class="bg-rose-50 border border-rose-200 text-rose-700 text-sm px-4 py-3 rounded-xl">
          {{ errorMessage }}
        </div>

        <div>
          <label class="block text-xs font-bold uppercase tracking-wider text-slate-500 mb-1">Bulan</label>
          <input
            v-model="month"
            type="month"
            class="w-full px-3.5 py-2.5 border border-slate-300 rounded-xl text-sm font-medium bg-white focus:ring-2 focus:ring-emerald-500"
          />
        </div>

        <div>
          <label class="block text-xs font-bold uppercase tracking-wider text-slate-500 mb-1">
            Kategori Pengeluaran
          </label>
          <select
            v-model="categoryId"
            class="w-full px-3.5 py-2.5 border border-slate-300 rounded-xl text-sm font-medium bg-white focus:ring-2 focus:ring-emerald-500"
          >
            <option value="" disabled>-- Pilih Kategori --</option>
            <option v-for="cat in expenseCategories" :key="cat.id" :value="cat.id">
              {{ cat.name }}
            </option>
          </select>
        </div>

        <div>
          <label class="block text-xs font-bold uppercase tracking-wider text-slate-500 mb-1">
            Batas Maksimal Bulanan (Limit IDR)
          </label>
          <div class="relative rounded-xl shadow-sm">
            <input
              type="text"
              inputmode="numeric"
              :value="rawLimit ? new Intl.NumberFormat('id-ID').format(rawLimit) : ''"
              @input="onLimitInput"
              placeholder="Contoh: 1.500.000"
              class="w-full text-xl font-bold font-mono px-4 py-3 border border-slate-300 rounded-xl focus:ring-2 focus:ring-emerald-500"
            />
            <div class="absolute inset-y-0 right-0 pr-4 flex items-center pointer-events-none text-xs font-semibold text-slate-400">
              IDR
            </div>
          </div>
          <p v-if="rawLimit" class="text-xs text-emerald-600 mt-1 font-semibold">
            Terbaca: {{ formatIDR(rawLimit) }}
          </p>
        </div>

        <div class="pt-4 border-t border-slate-100 flex items-center justify-end space-x-3">
          <button
            type="button"
            @click="emit('close')"
            class="px-4 py-2 text-sm font-semibold text-slate-600 hover:bg-slate-100 rounded-xl"
          >
            Batal
          </button>
          <button
            type="submit"
            :disabled="isSubmitting"
            class="px-5 py-2 text-sm font-semibold text-white bg-emerald-600 hover:bg-emerald-700 disabled:opacity-50 rounded-xl shadow-sm"
          >
            {{ isSubmitting ? 'Menyimpan...' : 'Simpan Anggaran' }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

