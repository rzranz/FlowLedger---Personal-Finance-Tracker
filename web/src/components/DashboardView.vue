<script setup lang="ts">
import { computed } from 'vue'
import type { MonthlyReport } from '../types'
import { formatIDR } from '../types'

const props = defineProps<{
  report: MonthlyReport | null
  loading: boolean
  currentMonth: string
}>()

const emit = defineEmits<{
  (e: 'update:currentMonth', month: string): void
  (e: 'open-add-modal'): void
  (e: 'open-budget-modal'): void
  (e: 'navigate-tab', tab: 'transactions' | 'budgets'): void
}>()

const savingsColorClass = computed(() => {
  if (!props.report) return 'text-slate-900'
  if (props.report.savings_rate_pct >= 20) return 'text-emerald-600'
  if (props.report.savings_rate_pct > 0) return 'text-amber-600'
  return 'text-rose-600'
})
</script>

<template>
  <div class="space-y-6">
    <!-- Top Bar: Month Picker & Summary Title -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 bg-white p-4 sm:p-6 rounded-2xl border border-slate-200 shadow-sm">
      <div>
        <h1 class="text-xl sm:text-2xl font-extrabold text-slate-900 tracking-tight">Executive Dashboard</h1>
        <p class="text-xs sm:text-sm text-slate-500">Agregasi performa finansial deterministik dengan kalkulasi paralel errgroup</p>
      </div>

      <div class="flex items-center space-x-3">
        <label class="text-xs font-bold uppercase tracking-wider text-slate-500">Periode:</label>
        <input
          type="month"
          :value="currentMonth"
          @change="emit('update:currentMonth', ($event.target as HTMLInputElement).value)"
          class="px-3.5 py-2 border border-slate-300 rounded-xl text-sm font-semibold text-slate-800 bg-slate-50 focus:bg-white focus:ring-2 focus:ring-emerald-500 focus:border-emerald-500"
        />
      </div>
    </div>

    <!-- Loading Skeleton -->
    <div v-if="loading" class="grid grid-cols-1 md:grid-cols-4 gap-4 animate-pulse">
      <div v-for="i in 4" :key="i" class="h-32 bg-white rounded-2xl border border-slate-200 p-6"></div>
    </div>

    <!-- 4 KPI Cards -->
    <div v-else-if="report" class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
      <!-- Total Income -->
      <div class="bg-white p-5 rounded-2xl border border-slate-200 shadow-sm hover:shadow-md transition-shadow">
        <div class="flex items-center justify-between">
          <span class="text-xs font-bold uppercase tracking-wider text-slate-500">Total Pemasukan</span>
          <span class="w-8 h-8 rounded-xl bg-emerald-50 text-emerald-600 flex items-center justify-center font-bold text-sm">
            <svg class="w-4 h-4 stroke-[2.2]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" d="M17 7L7 17M7 17H16M7 17V8"/>
            </svg>
          </span>
        </div>
        <div class="mt-3">
          <p class="text-2xl font-extrabold text-slate-900 font-mono tracking-tight">
            {{ formatIDR(report.total_income_idr) }}
          </p>
          <div class="mt-2 flex items-center text-xs">
            <span
              :class="report.mom_comparison.income_delta_pct >= 0 ? 'text-emerald-600 font-semibold' : 'text-slate-500'"
            >
              {{ report.mom_comparison.income_delta_pct >= 0 ? '▲ +' : '▼ ' }}{{ report.mom_comparison.income_delta_pct }}%
            </span>
            <span class="text-slate-400 ml-1.5">vs bulan lalu</span>
          </div>
        </div>
      </div>

      <!-- Total Expense -->
      <div class="bg-white p-5 rounded-2xl border border-slate-200 shadow-sm hover:shadow-md transition-shadow">
        <div class="flex items-center justify-between">
          <span class="text-xs font-bold uppercase tracking-wider text-slate-500">Total Pengeluaran</span>
          <span class="w-8 h-8 rounded-xl bg-rose-50 text-rose-600 flex items-center justify-center font-bold text-sm">
            <svg class="w-4 h-4 stroke-[2.2]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" d="M7 17L17 7M17 7H8M17 7V16"/>
            </svg>
          </span>
        </div>
        <div class="mt-3">
          <p class="text-2xl font-extrabold text-slate-900 font-mono tracking-tight">
            {{ formatIDR(report.total_expense_idr) }}
          </p>
          <div class="mt-2 flex items-center text-xs">
            <span
              :class="report.mom_comparison.expense_delta_pct <= 0 ? 'text-emerald-600 font-semibold' : 'text-rose-600 font-semibold'"
            >
              {{ report.mom_comparison.expense_delta_pct >= 0 ? '▲ +' : '▼ ' }}{{ report.mom_comparison.expense_delta_pct }}%
            </span>
            <span class="text-slate-400 ml-1.5">vs bulan lalu</span>
          </div>
        </div>
      </div>

      <!-- Net Cash Flow -->
      <div class="bg-white p-5 rounded-2xl border border-slate-200 shadow-sm hover:shadow-md transition-shadow">
        <div class="flex items-center justify-between">
          <span class="text-xs font-bold uppercase tracking-wider text-slate-500">Arus Kas Bersih</span>
          <span class="w-8 h-8 rounded-xl bg-blue-50 text-blue-600 flex items-center justify-center font-bold text-sm">
            <svg class="w-4 h-4 stroke-[2.2]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" d="M3 6l3 1m0 0l-3 9a5.002 5.002 0 006.001 0M6 7l3 9M6 7l6-2m6 2l3-1m-3 1l-3 9a5.002 5.002 0 006.001 0M18 7l3 9m-3-9l-6-2m0-2v2m0 16V5m0 16H9m3 0h3"/>
            </svg>
          </span>
        </div>
        <div class="mt-3">
          <p
            class="text-2xl font-extrabold font-mono tracking-tight"
            :class="report.net_cash_flow_idr >= 0 ? 'text-slate-900' : 'text-rose-600'"
          >
            {{ formatIDR(report.net_cash_flow_idr) }}
          </p>
          <div class="mt-2 flex items-center text-xs text-slate-500">
            <span>Pemasukan − Pengeluaran</span>
          </div>
        </div>
      </div>

      <!-- Savings Rate % -->
      <div class="bg-white p-5 rounded-2xl border border-slate-200 shadow-sm hover:shadow-md transition-shadow">
        <div class="flex items-center justify-between">
          <span class="text-xs font-bold uppercase tracking-wider text-slate-500">Rasio Tabungan</span>
          <span class="w-8 h-8 rounded-xl bg-teal-50 text-teal-600 flex items-center justify-center font-bold text-sm">
            <svg class="w-4 h-4 stroke-[2.2]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6"/>
            </svg>
          </span>
        </div>
        <div class="mt-3">
          <p class="text-2xl font-extrabold font-mono tracking-tight" :class="savingsColorClass">
            {{ report.savings_rate_pct }}%
          </p>
          <div class="mt-2 flex items-center text-xs text-slate-500">
            <span>Target sehat: ≥ 20%</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Analytics Detail Section: Category Breakdown & Budget Watchlist -->
    <div v-if="report" class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <!-- Category Spend Breakdown -->
      <div class="bg-white p-6 rounded-2xl border border-slate-200 shadow-sm">
        <div class="flex items-center justify-between mb-4">
          <div>
            <h2 class="text-base font-bold text-slate-900">Distribusi Pengeluaran</h2>
            <p class="text-xs text-slate-500">Rincian pengeluaran per kategori terurut dari terbesar</p>
          </div>
          <button
            @click="emit('navigate-tab', 'transactions')"
            class="text-xs font-semibold text-emerald-600 hover:text-emerald-700"
          >
            Lihat Transaksi →
          </button>
        </div>

        <div v-if="report.category_breakdown.length === 0" class="text-center py-10 text-slate-400 text-sm">
          Belum ada transaksi pengeluaran pada bulan ini.
        </div>

        <div v-else class="space-y-4">
          <div v-for="item in report.category_breakdown" :key="item.category_id" class="space-y-1.5">
            <div class="flex items-center justify-between text-xs sm:text-sm">
              <span class="font-medium text-slate-800">{{ item.category_name }}</span>
              <div class="flex items-center space-x-2">
                <span class="font-mono font-bold text-slate-900">{{ formatIDR(item.amount_idr) }}</span>
                <span class="text-xs font-semibold text-slate-500 bg-slate-100 px-1.5 py-0.5 rounded">{{ item.percentage }}%</span>
              </div>
            </div>
            <!-- Progress Bar -->
            <div class="w-full bg-slate-100 rounded-full h-2 overflow-hidden">
              <div
                class="bg-emerald-500 h-2 rounded-full transition-all duration-300"
                :style="{ width: `${item.percentage}%` }"
              ></div>
            </div>
          </div>
        </div>
      </div>

      <!-- Budget Watchlist Cards -->
      <div class="bg-white p-6 rounded-2xl border border-slate-200 shadow-sm">
        <div class="flex items-center justify-between mb-4">
          <div>
            <h2 class="text-base font-bold text-slate-900">Radar Anggaran (Budget Watchlist)</h2>
            <p class="text-xs text-slate-500">Guardrails peringatan SAFE (&lt;80%), WARNING (80-99%), EXCEEDED (&ge;100%)</p>
          </div>
          <button
            @click="emit('navigate-tab', 'budgets')"
            class="text-xs font-semibold text-emerald-600 hover:text-emerald-700"
          >
            Kelola Anggaran →
          </button>
        </div>

        <div v-if="report.budget_watchlist.length === 0" class="text-center py-10">
          <p class="text-slate-400 text-sm mb-3">Belum ada batas anggaran yang diatur untuk bulan ini.</p>
          <button
            @click="emit('open-budget-modal')"
            class="text-xs font-semibold bg-emerald-50 text-emerald-700 px-3 py-1.5 rounded-lg border border-emerald-200 hover:bg-emerald-100"
          >
            + Buat Anggaran Pertama
          </button>
        </div>

        <div v-else class="space-y-4">
          <div
            v-for="item in report.budget_watchlist"
            :key="item.budget.id"
            class="p-4 rounded-xl border transition-all"
            :class="[
              item.status === 'EXCEEDED' ? 'bg-rose-50/50 border-rose-200' :
              item.status === 'WARNING' ? 'bg-amber-50/50 border-amber-200' :
              'bg-slate-50/50 border-slate-200'
            ]"
          >
            <div class="flex items-center justify-between mb-2">
              <span class="text-sm font-bold text-slate-900">{{ item.category_name }}</span>
              <span
                class="text-[10px] font-extrabold uppercase px-2 py-0.5 rounded-full tracking-wider"
                :class="[
                  item.status === 'EXCEEDED' ? 'bg-rose-100 text-rose-800' :
                  item.status === 'WARNING' ? 'bg-amber-100 text-amber-800' :
                  'bg-emerald-100 text-emerald-800'
                ]"
              >
                {{ item.status }}
              </span>
            </div>

            <!-- Spend vs Limit -->
            <div class="flex items-center justify-between text-xs font-mono mb-2">
              <span class="text-slate-600">Terpakai: <b>{{ formatIDR(item.spent_idr) }}</b></span>
              <span class="text-slate-400">Limit: {{ formatIDR(item.budget.limit_idr) }}</span>
            </div>

            <!-- Progress Bar -->
            <div class="w-full bg-slate-200/80 rounded-full h-2.5 overflow-hidden">
              <div
                class="h-2.5 rounded-full transition-all duration-300"
                :class="[
                  item.status === 'EXCEEDED' ? 'bg-rose-500' :
                  item.status === 'WARNING' ? 'bg-amber-500' :
                  'bg-emerald-500'
                ]"
                :style="{ width: `${Math.min(item.utilization_pct, 100)}%` }"
              ></div>
            </div>

            <div class="mt-2 flex items-center justify-between text-xs text-slate-500">
              <span>Utilisasi: <b>{{ item.utilization_pct }}%</b></span>
              <span>Sisa: <b :class="item.remaining_idr < 0 ? 'text-rose-600 font-mono' : 'text-emerald-700 font-mono'">{{ formatIDR(item.remaining_idr) }}</b></span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

