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
    class="fixed inset-0 z-50 overflow-y-auto bg-slate-950/70 backdrop-blur-md flex items-center justify-center p-4"
    @click.self="emit('close')"
  >
    <div class="bg-[#0f131c] text-white rounded-3xl shadow-2xl border border-white/10 max-w-md w-full overflow-hidden animate-in fade-in zoom-in-95 duration-200">
      <div class="px-6 py-4 border-b border-white/10 flex items-center justify-between">
        <div class="flex items-center space-x-2.5">
          <span class="w-2.5 h-2.5 rounded-full bg-amber-400"></span>
          <h3 class="text-base font-extrabold text-white tracking-tight">Atur Limit Anggaran</h3>
        </div>
        <button @click="emit('close')" class="text-slate-400 hover:text-white p-1.5 rounded-lg hover:bg-white/10">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.2" d="M6 18L18 6M6 6l12 12"/>
          </svg>
        </button>
      </div>

      <form @submit.prevent="handleSave" class="p-6 space-y-5">
        <div v-if="errorMessage" class="bg-rose-500/15 border border-rose-500/30 text-rose-200 text-xs px-4 py-3 rounded-xl">
          {{ errorMessage }}
        </div>

        <div>
          <label class="block text-[11px] font-mono font-bold uppercase tracking-widest text-slate-400 mb-1.5">Periode Bulan</label>
          <input
            v-model="month"
            type="month"
            class="w-full px-4 py-3 bg-[#161c2a] border border-white/15 rounded-xl text-sm font-semibold text-white focus:outline-none focus:border-emerald-400"
          />
        </div>

        <div>
          <label class="block text-[11px] font-mono font-bold uppercase tracking-widest text-slate-400 mb-1.5">
            Kategori Pengeluaran
          </label>
          <select
            v-model="categoryId"
            class="w-full px-4 py-3 bg-[#161c2a] border border-white/15 rounded-xl text-sm font-semibold text-white focus:outline-none focus:border-emerald-400"
          >
            <option value="" disabled class="bg-slate-900 text-slate-400">-- Pilih Kategori --</option>
            <option v-for="cat in expenseCategories" :key="cat.id" :value="cat.id" class="bg-slate-900 text-white">
              {{ cat.name }}
            </option>
          </select>
        </div>

        <div>
          <label class="block text-[11px] font-mono font-bold uppercase tracking-widest text-slate-400 mb-1.5">
            Batas Maksimal Bulanan (IDR)
          </label>
          <div class="relative rounded-xl">
            <input
              type="text"
              inputmode="numeric"
              :value="rawLimit ? new Intl.NumberFormat('id-ID').format(rawLimit) : ''"
              @input="onLimitInput"
              placeholder="Contoh: 1.500.000"
              class="w-full text-xl font-black font-mono px-4 py-3 bg-[#161c2a] border border-white/15 rounded-xl text-white placeholder:text-slate-500 focus:outline-none focus:border-emerald-400"
            />
            <div class="absolute inset-y-0 right-0 pr-4 flex items-center pointer-events-none text-xs font-mono font-bold text-slate-400">
              IDR
            </div>
          </div>
          <p v-if="rawLimit" class="text-xs text-emerald-400 mt-1.5 font-mono font-semibold">
            Limit Aktif: {{ formatIDR(rawLimit) }}
          </p>
        </div>

        <div class="pt-3 border-t border-white/10 flex items-center justify-end space-x-3">
          <button
            type="button"
            @click="emit('close')"
            class="tactile-btn px-4 py-2.5 text-xs font-bold text-slate-300 hover:text-white bg-white/[0.06] hover:bg-white/10 rounded-xl"
          >
            Batal
          </button>
          <button
            type="submit"
            :disabled="isSubmitting"
            class="tactile-btn px-5 py-2.5 text-xs font-black text-slate-950 bg-emerald-400 hover:bg-emerald-300 disabled:opacity-50 rounded-xl shadow-md"
          >
            {{ isSubmitting ? 'Menyimpan...' : 'Simpan Limit' }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>
