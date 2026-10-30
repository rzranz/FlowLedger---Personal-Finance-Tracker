<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { api } from '../api/client'
import type { BudgetReportItem, Category } from '../types'
import { formatIDR } from '../types'

const props = defineProps<{
  categories: Category[]
  currentMonth: string
}>()

const emit = defineEmits<{
  (e: 'open-budget-modal'): void
}>()

const selectedMonth = ref(props.currentMonth)
const budgets = ref<BudgetReportItem[]>([])
const loading = ref(false)

async function loadBudgets() {
  loading.value = true
  try {
    const res = await api.getBudgets(selectedMonth.value)
    budgets.value = res.items || []
  } catch (err) {
    console.error('Failed to load budgets:', err)
  } finally {
    loading.value = false
  }
}

watch(selectedMonth, () => {
  loadBudgets()
})

onMounted(() => {
  loadBudgets()
})

defineExpose({
  loadBudgets,
})
</script>

<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 bg-white p-4 sm:p-6 rounded-2xl border border-slate-200 shadow-sm">
      <div>
        <h1 class="text-xl sm:text-2xl font-extrabold text-slate-900 tracking-tight">Manajemen Anggaran (Budgets)</h1>
        <p class="text-xs sm:text-sm text-slate-500">Guardrails pengeluaran per kategori dengan indikator ambang batas otomatis</p>
      </div>

      <div class="flex items-center space-x-3">
        <input
          v-model="selectedMonth"
          type="month"
          class="px-3.5 py-2 border border-slate-300 rounded-xl text-sm font-semibold text-slate-800 bg-slate-50 focus:bg-white focus:ring-2 focus:ring-emerald-500"
        />
        <button
          @click="emit('open-budget-modal')"
          class="px-4 py-2 bg-emerald-600 hover:bg-emerald-700 text-white rounded-xl text-sm font-semibold shadow-sm transition-all"
        >
          + Pasang Anggaran
        </button>
      </div>
    </div>

    <!-- Budgets List Cards -->
    <div v-if="loading" class="grid grid-cols-1 md:grid-cols-2 gap-4 animate-pulse">
      <div v-for="i in 4" :key="i" class="h-40 bg-white rounded-2xl border border-slate-200 p-6"></div>
    </div>

    <div v-else-if="budgets.length === 0" class="bg-white rounded-2xl border border-slate-200 p-12 text-center">
      <p class="text-slate-500 font-medium mb-3">Belum ada anggaran pengeluaran yang diatur untuk periode {{ selectedMonth }}.</p>
      <button
        @click="emit('open-budget-modal')"
        class="px-4 py-2 bg-emerald-600 hover:bg-emerald-700 text-white rounded-xl text-sm font-semibold shadow-sm"
      >
        + Pasang Batas Anggaran Pertama
      </button>
    </div>

    <div v-else class="grid grid-cols-1 md:grid-cols-2 gap-4">
      <div
        v-for="item in budgets"
        :key="item.budget.id"
        class="bg-white p-6 rounded-2xl border transition-all shadow-sm hover:shadow-md"
        :class="[
          item.status === 'EXCEEDED' ? 'border-rose-300 ring-1 ring-rose-300/50' :
          item.status === 'WARNING' ? 'border-amber-300 ring-1 ring-amber-300/50' :
          'border-slate-200'
        ]"
      >
        <div class="flex items-center justify-between mb-3">
          <div>
            <h3 class="text-base font-bold text-slate-900">{{ item.category_name }}</h3>
            <p class="text-xs text-slate-500">Periode {{ item.budget.month }}</p>
          </div>
          <span
            class="text-[11px] font-extrabold uppercase px-2.5 py-1 rounded-full tracking-wider"
            :class="[
              item.status === 'EXCEEDED' ? 'bg-rose-100 text-rose-800' :
              item.status === 'WARNING' ? 'bg-amber-100 text-amber-800' :
              'bg-emerald-100 text-emerald-800'
            ]"
          >
            {{ item.status }}
          </span>
        </div>

        <!-- Numbers -->
        <div class="grid grid-cols-3 gap-2 py-3 border-y border-slate-100 text-xs">
          <div>
            <span class="text-slate-400 block text-[11px] uppercase font-bold">Terpakai</span>
            <span class="font-mono font-bold text-slate-900">{{ formatIDR(item.spent_idr) }}</span>
          </div>
          <div>
            <span class="text-slate-400 block text-[11px] uppercase font-bold">Batas Limit</span>
            <span class="font-mono font-semibold text-slate-700">{{ formatIDR(item.budget.limit_idr) }}</span>
          </div>
          <div class="text-right">
            <span class="text-slate-400 block text-[11px] uppercase font-bold">Sisa Budget</span>
            <span
              class="font-mono font-bold"
              :class="item.remaining_idr < 0 ? 'text-rose-600' : 'text-emerald-600'"
            >
              {{ formatIDR(item.remaining_idr) }}
            </span>
          </div>
        </div>

        <!-- Visual Bar -->
        <div class="mt-4 space-y-1.5">
          <div class="flex justify-between text-xs text-slate-500">
            <span>Utilisasi Anggaran</span>
            <span class="font-bold font-mono">{{ item.utilization_pct }}%</span>
          </div>
          <div class="w-full bg-slate-100 rounded-full h-3 overflow-hidden">
            <div
              class="h-3 rounded-full transition-all duration-300"
              :class="[
                item.status === 'EXCEEDED' ? 'bg-rose-500' :
                item.status === 'WARNING' ? 'bg-amber-500' :
                'bg-emerald-500'
              ]"
              :style="{ width: `${Math.min(item.utilization_pct, 100)}%` }"
            ></div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

