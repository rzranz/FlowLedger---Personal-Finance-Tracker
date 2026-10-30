<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { api } from './api/client'
import type { Category, MonthlyReport, Transaction, PaymentMethod } from './types'
import { getCurrentMonth, formatIDR, PAYMENT_METHODS, extractPaymentMethod, cleanNote } from './types'

import CardsStack from './components/CardsStack.vue'
import CardsExpandedView from './components/CardsExpandedView.vue'
import BottomDock from './components/BottomDock.vue'
import BudgetsView from './components/BudgetsView.vue'
import CategoriesView from './components/CategoriesView.vue'
import AddTransactionModal from './components/AddTransactionModal.vue'
import UpsertBudgetModal from './components/UpsertBudgetModal.vue'
import AddCategoryModal from './components/AddCategoryModal.vue'

// Active Tab in navigation
const activeTab = ref<'dashboard' | 'transactions' | 'budgets' | 'categories'>('dashboard')
const currentMonth = ref<string>(getCurrentMonth())
const isReady = ref(false)

// Card Expansion State: When true, dark hero slides down and reveals vertical cascading cards
const isDeckExpanded = ref(false)

// Active Card Payment Method filter ('all' | 'qris' | 'cash' | 'atm' | 'transfer')
const activeMethod = ref<PaymentMethod>('all')

// Selected Category filter (when clicking avatar circles in Spending widget)
const selectedCategoryFilter = ref<number | null>(null)

// Search in bottom transactions sheet
const showSearch = ref(false)
const searchQuery = ref('')
const searchInputRef = ref<HTMLInputElement | null>(null)

// Notification Bell drawer/popover
const showNotifications = ref(false)

// Data state
const categories = ref<Category[]>([])
const transactions = ref<Transaction[]>([])
const monthlyReport = ref<MonthlyReport | null>(null)
const loading = ref(false)

// Modals
const showAddTxModal = ref(false)
const defaultAddTxType = ref<'income' | 'expense'>('expense')
const showBudgetModal = ref(false)
const showCategoryModal = ref(false)

// Global notification toast
const toastMessage = ref<string | null>(null)
let toastTimer: any = null

function showToast(msg: string) {
  toastMessage.value = msg
  clearTimeout(toastTimer)
  toastTimer = setTimeout(() => {
    toastMessage.value = null
  }, 3500)
}

function openAddExpense() {
  defaultAddTxType.value = 'expense'
  showAddTxModal.value = true
}

function openAddIncome() {
  defaultAddTxType.value = 'income'
  showAddTxModal.value = true
}

// Calculate total spending per payment method for the current month
const methodSpendMap = computed(() => {
  const map: Record<PaymentMethod, number> = {
    all: 0,
    qris: 0,
    cash: 0,
    atm: 0,
    transfer: 0,
  }

  for (const tx of transactions.value) {
    if (tx.type === 'expense') {
      const m = extractPaymentMethod(tx.note, tx.id)
      map[m] = (map[m] || 0) + tx.amount_idr
      map.all += tx.amount_idr
    }
  }

  return map
})

// Current spending display (depends on active card filter)
const currentSpendingAmount = computed(() => {
  if (activeMethod.value === 'all') {
    return monthlyReport.value ? monthlyReport.value.total_expense_idr : methodSpendMap.value.all
  }
  return methodSpendMap.value[activeMethod.value] || 0
})

const activeCardLabel = computed(() => {
  if (activeMethod.value === 'all') return 'Total Pengeluaran'
  const found = PAYMENT_METHODS.find((m) => m.id === activeMethod.value)
  return found ? `Pengeluaran (${found.shortName})` : 'Pengeluaran'
})

// Budget alert items (for the notification bell)
const budgetAlerts = computed(() => {
  if (!monthlyReport.value?.budget_watchlist) return []
  return monthlyReport.value.budget_watchlist.filter((b) => b.status === 'WARNING' || b.status === 'EXCEEDED')
})

// Filtered transactions for the bottom sheet
const displayedTransactions = computed(() => {
  let list = transactions.value

  // Filter by active card method
  if (activeMethod.value !== 'all') {
    list = list.filter((tx) => extractPaymentMethod(tx.note, tx.id) === activeMethod.value)
  }

  // Filter by selected category (from spending widget circles)
  if (selectedCategoryFilter.value !== null) {
    list = list.filter((tx) => tx.category_id === selectedCategoryFilter.value)
  }

  // Filter by search query
  if (searchQuery.value.trim()) {
    const q = searchQuery.value.toLowerCase()
    list = list.filter(
      (tx) =>
        (tx.note && tx.note.toLowerCase().includes(q)) ||
        (tx.category && tx.category.name.toLowerCase().includes(q))
    )
  }

  return list
})

function onMethodSelectedFromDeck(method: PaymentMethod) {
  activeMethod.value = method
  isDeckExpanded.value = false
  if (method === 'all') {
    showToast('Tampilkan semua transaksi')
  } else {
    const found = PAYMENT_METHODS.find((m) => m.id === method)
    showToast(`Filter aktif: ${found?.name || method.toUpperCase()}`)
  }
}

function toggleCategoryFilter(catId: number) {
  if (selectedCategoryFilter.value === catId) {
    selectedCategoryFilter.value = null
  } else {
    selectedCategoryFilter.value = catId
  }
}

function toggleSearch() {
  showSearch.value = !showSearch.value
  if (showSearch.value) {
    setTimeout(() => {
      searchInputRef.value?.focus()
    }, 100)
  } else {
    searchQuery.value = ''
  }
}

async function checkReadiness() {
  try {
    const res = await api.getReady()
    isReady.value = res.status === 'ready'
  } catch {
    isReady.value = false
  }
}

async function loadCategories() {
  try {
    const res = await api.getCategories()
    categories.value = res.items || []
  } catch (err) {
    console.error('Failed to load categories:', err)
  }
}

async function loadMonthlyReport() {
  try {
    const res = await api.getMonthlyReport(currentMonth.value)
    monthlyReport.value = res
  } catch (err) {
    console.error('Failed to load monthly report:', err)
  }
}

async function loadTransactions() {
  loading.value = true
  try {
    const res = await api.getTransactions({
      month: currentMonth.value,
      limit: 100,
    })
    transactions.value = res.items || []
  } catch (err) {
    console.error('Failed to load transactions:', err)
  } finally {
    loading.value = false
  }
}

async function refreshAll() {
  await Promise.all([loadMonthlyReport(), loadTransactions(), loadCategories()])
}

watch(currentMonth, () => {
  refreshAll()
})

function onTransactionCreated() {
  showToast('✓ Transaksi berhasil dicatat!')
  refreshAll()
}

function onBudgetSaved() {
  showToast('✓ Batas anggaran berhasil disimpan!')
  refreshAll()
}

function onCategoryCreated() {
  showToast('✓ Kategori baru berhasil ditambahkan!')
  loadCategories()
}

async function handleDeleteTx(id: number) {
  if (!confirm('Hapus transaksi ini?')) return
  try {
    await api.deleteTransaction(id)
    showToast('✓ Transaksi berhasil dihapus')
    refreshAll()
  } catch (err: any) {
    alert(err?.detail || 'Gagal menghapus transaksi')
  }
}

// Global Keyboard Shortcut: 'N' opens Add Transaction modal
function handleKeyDown(e: KeyboardEvent) {
  const tag = (e.target as HTMLElement)?.tagName
  if (tag === 'INPUT' || tag === 'TEXTAREA' || tag === 'SELECT') {
    return
  }

  if (e.key === 'n' || e.key === 'N') {
    e.preventDefault()
    openAddExpense()
  }
}

// Category visual initial
function getCategoryInitial(name: string): string {
  if (!name) return '•'
  return name.trim().charAt(0).toUpperCase()
}

function getMethodBadgeColor(m: PaymentMethod): string {
  switch (m) {
    case 'qris':
      return 'bg-emerald-500/10 text-emerald-600 border-emerald-500/20'
    case 'cash':
      return 'bg-amber-500/10 text-amber-600 border-amber-500/20'
    case 'atm':
      return 'bg-cyan-500/10 text-cyan-600 border-cyan-500/20'
    case 'transfer':
      return 'bg-purple-500/10 text-purple-600 border-purple-500/20'
    default:
      return 'bg-slate-100 text-slate-600 border-slate-200'
  }
}

onMounted(() => {
  window.addEventListener('keydown', handleKeyDown)
  checkReadiness()
  refreshAll()
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeyDown)
})
</script>

<template>
  <!-- Main Studio Canvas (Refined Luxury Dark Theme) -->
  <div class="min-h-screen bg-[#07090e] text-slate-100 flex flex-col items-center justify-start relative selection:bg-emerald-500 selection:text-white">
    <!-- Top Bar: Minimalist Luxury Header -->
    <header class="w-full max-w-5xl px-4 sm:px-6 py-4 flex items-center justify-between z-30 border-b border-white/[0.06] text-xs">
      <div class="flex items-center space-x-3">
        <div class="w-8 h-8 rounded-xl bg-white/[0.08] border border-white/[0.12] flex items-center justify-center font-extrabold text-white text-xs shadow-inner">
          FL
        </div>
        <div class="flex items-center space-x-2">
          <span class="font-extrabold tracking-tight text-white text-sm">FlowLedger</span>
          <span class="text-[9px] font-mono font-bold uppercase tracking-widest bg-white/[0.06] text-slate-400 px-2 py-0.5 rounded-full border border-white/[0.08]">
            v2.0
          </span>
        </div>
      </div>

      <div class="flex items-center space-x-3">
        <div class="hidden sm:flex items-center space-x-2 text-slate-400 text-xs font-mono">
          <span class="w-1.5 h-1.5 rounded-full" :class="isReady ? 'bg-emerald-400' : 'bg-amber-400 animate-ping'"></span>
          <span>{{ isReady ? 'SQLite WAL' : 'Connecting...' }}</span>
        </div>
        <a
          href="/api/v1/transactions/export.csv"
          target="_blank"
          class="bg-white/[0.06] hover:bg-white/[0.1] text-slate-200 hover:text-white px-3.5 py-1.5 rounded-full border border-white/[0.1] font-medium transition-all flex items-center space-x-1.5 shadow-sm active:scale-95"
        >
          <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"/>
          </svg>
          <span>Export CSV</span>
        </a>
      </div>
    </header>

    <!-- Main Fluid Layout Container -->
    <div class="w-full max-w-5xl flex-1 flex flex-col justify-start relative pb-36">
      <!-- ======================================================== -->
      <!-- TOP DARK HERO SECTION                                   -->
      <!-- ======================================================== -->
      <section class="px-4 sm:px-8 pt-5 pb-6 space-y-6">
        <!-- Top Profile & Period Row -->
        <div class="flex items-center justify-between relative">
          <!-- User Profile & Status -->
          <div class="flex items-center space-x-3.5">
            <div class="relative">
              <div class="w-11 h-11 rounded-full bg-gradient-to-tr from-slate-800 to-slate-700 border border-white/[0.15] flex items-center justify-center text-white font-extrabold text-sm shadow-md">
                RZ
              </div>
              <span
                class="absolute bottom-0 right-0 w-3 h-3 rounded-full border-2 border-[#07090e]"
                :class="isReady ? 'bg-emerald-500' : 'bg-amber-500 animate-pulse'"
              ></span>
            </div>
            <div>
              <div class="text-[10px] font-mono font-bold uppercase tracking-widest text-slate-400">
                Personal Ledger
              </div>
              <div class="text-sm sm:text-base font-bold text-white tracking-tight">
                Randi Zakaria
              </div>
            </div>
          </div>

          <!-- Month Selector & Notification Bell -->
          <div class="flex items-center space-x-2">
            <input
              type="month"
              v-model="currentMonth"
              class="bg-white/[0.06] border border-white/[0.12] text-xs sm:text-sm font-semibold text-slate-200 px-3 py-1.5 rounded-full focus:ring-2 focus:ring-emerald-500 shadow-sm"
            />
            
            <!-- Bell Button with Active Alert Badge -->
            <div class="relative">
              <button
                @click="showNotifications = !showNotifications"
                class="w-9 h-9 rounded-full bg-white/[0.06] hover:bg-white/[0.12] border border-white/[0.12] text-slate-300 hover:text-white flex items-center justify-center transition-all active:scale-90"
                title="Pemberitahuan & Alert Anggaran"
              >
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9"/>
                </svg>
                <span
                  v-if="budgetAlerts.length > 0"
                  class="absolute -top-0.5 -right-0.5 w-2.5 h-2.5 bg-rose-500 border-2 border-[#07090e] rounded-full animate-pulse"
                ></span>
              </button>

              <!-- Notifications Dropdown Popover -->
              <div
                v-if="showNotifications"
                class="absolute right-0 top-11 w-72 sm:w-80 bg-slate-900 border border-white/[0.12] rounded-2xl shadow-2xl p-4 z-50 text-xs animate-in fade-in zoom-in-95 duration-150 backdrop-blur-xl"
              >
                <div class="flex items-center justify-between pb-2 border-b border-white/[0.08] mb-2">
                  <span class="font-bold text-white uppercase tracking-wider text-[10px] font-mono">Peringatan Anggaran</span>
                  <button @click="showNotifications = false" class="text-slate-400 hover:text-white">✕</button>
                </div>

                <div v-if="budgetAlerts.length === 0" class="py-4 text-center text-slate-400">
                  Semua anggaran dalam batas aman (SAFE).
                </div>
                <div v-else class="space-y-2 max-h-56 overflow-y-auto">
                  <div
                    v-for="b in budgetAlerts"
                    :key="b.budget.id"
                    class="p-2.5 rounded-xl border flex items-start space-x-2.5"
                    :class="b.status === 'EXCEEDED' ? 'bg-rose-950/40 border-rose-800/60 text-rose-200' : 'bg-amber-950/40 border-amber-800/60 text-amber-200'"
                  >
                    <div class="flex-1">
                      <div class="font-bold text-xs">{{ b.category_name }}</div>
                      <div class="text-[11px] opacity-80 mt-0.5">
                        {{ b.status === 'EXCEEDED' ? 'Melebihi batas anggaran' : 'Mendekati batas anggaran' }} ({{ b.utilization_pct }}%)
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Hero Middle: Total Balance & Interactive Cards Stack -->
        <div class="grid grid-cols-1 md:grid-cols-12 gap-6 items-center">
          <!-- Total Balance (Left / Main) -->
          <div class="md:col-span-6 space-y-3.5">
            <div class="text-[11px] font-mono font-bold uppercase tracking-widest text-slate-400 flex items-center space-x-2">
              <span>Total Saldo Bersih</span>
              <span class="text-[9px] font-semibold bg-emerald-500/10 text-emerald-400 px-2 py-0.5 rounded-full border border-emerald-500/20">
                Net Cash Flow
              </span>
            </div>
            <div class="text-3xl sm:text-4xl lg:text-5xl font-black font-mono text-white tracking-tight">
              {{ formatIDR(monthlyReport ? monthlyReport.net_cash_flow_idr : 0) }}
            </div>

            <!-- Action Pill Buttons (Send / Request / Anggaran) with Doppelrand Double-Bezel -->
            <div class="flex items-center space-x-2.5 pt-1">
              <!-- Send (Catat Pengeluaran) -->
              <button
                @click="openAddExpense"
                class="bg-white hover:bg-slate-100 text-slate-950 font-bold px-4 py-2 rounded-full text-xs shadow-md transition-all active:scale-95 flex items-center space-x-2"
                title="Catat Pengeluaran Baru"
              >
                <span>Catat Pengeluaran</span>
                <span class="w-5 h-5 rounded-full bg-slate-950/10 flex items-center justify-center text-[11px] font-bold">↗</span>
              </button>
              
              <!-- Request (Catat Pemasukan) -->
              <button
                @click="openAddIncome"
                class="bg-white/[0.08] hover:bg-white/[0.14] text-white font-semibold px-4 py-2 rounded-full text-xs border border-white/[0.12] transition-all active:scale-95 flex items-center space-x-2 shadow-inner"
                title="Catat Pemasukan Baru"
              >
                <span>Pemasukan</span>
                <span class="w-5 h-5 rounded-full bg-white/10 flex items-center justify-center text-[11px] font-bold">↙</span>
              </button>

              <!-- Anggaran Button -->
              <button
                @click="showBudgetModal = true"
                class="bg-white/[0.04] hover:bg-white/[0.08] text-slate-300 font-medium px-3.5 py-2 rounded-full text-xs border border-white/[0.08] transition-all active:scale-95"
                title="Atur Limit Anggaran"
              >
                <span>Limit</span>
              </button>
            </div>
          </div>

          <!-- Cards Stack (Clicking this expands the full vertical deck) -->
          <div class="md:col-span-6 flex justify-end">
            <CardsStack
              :active-method="activeMethod"
              :method-spend-map="methodSpendMap"
              :is-expanded="isDeckExpanded"
              @update:active-method="activeMethod = $event"
              @toggle-expand="isDeckExpanded = !isDeckExpanded"
            />
          </div>
        </div>

        <!-- ======================================================== -->
        <!-- EXPANDABLE CARDS DECK (Smooth Dark Curtain Overlay)     -->
        <!-- ======================================================== -->
        <CardsExpandedView
          :is-open="isDeckExpanded"
          :active-method="activeMethod"
          :method-spend-map="methodSpendMap"
          :total-balance="monthlyReport ? monthlyReport.net_cash_flow_idr : 0"
          @select-method="onMethodSelectedFromDeck"
          @close="isDeckExpanded = false"
        />

        <!-- Spending Widget & Category Filter Pills (When deck is not expanded) -->
        <div
          v-if="!isDeckExpanded"
          class="bg-white/[0.04] border border-white/[0.08] p-4 sm:p-5 rounded-2xl flex flex-col sm:flex-row sm:items-center justify-between gap-3 shadow-lg backdrop-blur-md transition-all"
        >
          <div>
            <div class="text-[10px] font-mono font-bold uppercase tracking-widest text-slate-400 flex items-center space-x-2">
              <span>{{ activeCardLabel }}</span>
              <span v-if="selectedCategoryFilter !== null" class="text-emerald-400 text-[10px]">
                (Kategori Terfilter)
              </span>
            </div>
            <div class="text-xl sm:text-2xl font-black font-mono text-white tracking-tight mt-0.5">
              {{ formatIDR(currentSpendingAmount) }}
            </div>
          </div>

          <!-- Interactive Category Filter Chips -->
          <div class="flex items-center space-x-2 overflow-x-auto py-1">
            <span class="text-[11px] font-mono text-slate-400 mr-1 hidden sm:inline">Kategori:</span>
            <button
              v-for="(cat, idx) in categories.slice(0, 5)"
              :key="cat.id"
              @click="toggleCategoryFilter(cat.id)"
              :class="[
                'w-9 h-9 rounded-full flex items-center justify-center font-bold text-xs text-white shadow-sm transition-all duration-200 active:scale-90 border',
                selectedCategoryFilter === cat.id
                  ? 'ring-2 ring-white scale-110 shadow-lg border-white'
                  : 'hover:scale-105 opacity-80 hover:opacity-100 border-white/10',
                idx === 0 ? 'bg-rose-600' :
                idx === 1 ? 'bg-amber-600' :
                idx === 2 ? 'bg-cyan-600' :
                idx === 3 ? 'bg-purple-600' :
                'bg-emerald-600'
              ]"
              :title="`Filter: ${cat.name}`"
            >
              {{ getCategoryInitial(cat.name) }}
            </button>

            <!-- Reset category filter button -->
            <button
              v-if="selectedCategoryFilter !== null"
              @click="selectedCategoryFilter = null"
              class="text-[10px] font-bold text-slate-300 hover:text-white bg-white/10 px-3 py-1.5 rounded-full border border-white/15 ml-1 transition-all"
            >
              ✕ Reset
            </button>
          </div>
        </div>
      </section>

      <!-- ======================================================== -->
      <!-- BOTTOM WHITE SHEET SECTION (Smooth Transition on Expand) -->
      <!-- ======================================================== -->
      <section
        :class="[
          'flex-1 bg-white rounded-t-[32px] sm:rounded-t-[40px] px-4 sm:px-8 pt-6 pb-20 text-slate-800 shadow-2xl flex flex-col transition-all duration-500 ease-[cubic-bezier(0.32,0.72,0,1)]',
          isDeckExpanded ? 'translate-y-8 opacity-25 pointer-events-none' : 'translate-y-0 opacity-100'
        ]"
      >
        <!-- Sheet Header Bar -->
        <div class="flex items-center justify-between pb-4 border-b border-slate-100 mb-4">
          <!-- Left Hamburger Toggle -->
          <button
            @click="activeTab = activeTab === 'dashboard' ? 'transactions' : 'dashboard'"
            class="text-slate-500 hover:text-slate-900 p-2 rounded-xl hover:bg-slate-100 transition-colors flex items-center space-x-1.5 text-xs font-semibold"
            :title="activeTab === 'dashboard' ? 'Buka Tabel Transaksi Lengkap' : 'Buka Dashboard Ringkasan'"
          >
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.2" d="M4 6h16M4 12h10M4 18h16"/>
            </svg>
            <span class="hidden sm:inline font-mono">{{ activeTab === 'dashboard' ? 'Tabel Transaksi' : 'Ringkasan' }}</span>
          </button>

          <!-- Center Title with Context Badge -->
          <div class="text-center">
            <h2 class="text-base sm:text-lg font-extrabold text-slate-900 tracking-tight">
              {{
                activeTab === 'budgets' ? 'Manajemen Anggaran' :
                activeTab === 'categories' ? 'Struktur Kategori' :
                'Transactions'
              }}
            </h2>
            <div class="flex items-center justify-center space-x-2 text-xs font-semibold mt-0.5">
              <span v-if="activeMethod !== 'all'" class="text-emerald-700 bg-emerald-50 px-2 py-0.5 rounded-full border border-emerald-200">
                Kartu: {{ activeMethod.toUpperCase() }}
              </span>
              <span v-if="selectedCategoryFilter !== null" class="text-cyan-700 bg-cyan-50 px-2 py-0.5 rounded-full border border-cyan-200">
                Kategori Terfilter
              </span>
              <span class="text-slate-400 font-mono text-[11px]">
                ({{ displayedTransactions.length }} baris)
              </span>
            </div>
          </div>

          <!-- Right Search Toggle Button -->
          <button
            @click="toggleSearch"
            :class="[
              'p-2 rounded-xl transition-all duration-200 flex items-center space-x-1',
              showSearch ? 'bg-slate-900 text-white' : 'text-slate-500 hover:text-slate-900 hover:bg-slate-100'
            ]"
            title="Cari Transaksi"
          >
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"/>
            </svg>
          </button>
        </div>

        <!-- Live Search Input Bar -->
        <div v-if="showSearch" class="mb-4 animate-in fade-in slide-in-from-top-2 duration-150">
          <div class="relative">
            <input
              ref="searchInputRef"
              v-model="searchQuery"
              type="text"
              placeholder="Cari transaksi berdasarkan catatan atau kategori..."
              class="w-full pl-10 pr-10 py-2.5 border border-slate-300 rounded-2xl text-sm font-medium focus:ring-2 focus:ring-emerald-500 focus:border-emerald-500 shadow-sm"
            />
            <span class="absolute left-3.5 top-3 text-slate-400 text-sm">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"/>
              </svg>
            </span>
            <button
              v-if="searchQuery"
              @click="searchQuery = ''"
              class="absolute right-3.5 top-2.5 text-slate-400 hover:text-slate-600 font-bold text-sm"
            >
              ✕
            </button>
          </div>
        </div>

        <!-- TAB CONTENT VIEW AREA -->

        <!-- 1. Budgets View Tab -->
        <div v-if="activeTab === 'budgets'" class="flex-1">
          <BudgetsView
            :categories="categories"
            :current-month="currentMonth"
            @open-budget-modal="showBudgetModal = true"
          />
        </div>

        <!-- 2. Categories View Tab -->
        <div v-else-if="activeTab === 'categories'" class="flex-1">
          <CategoriesView
            :categories="categories"
            @refresh="loadCategories"
            @open-add-modal="showCategoryModal = true"
          />
        </div>

        <!-- 3. Default: Transactions List (Matching Image, Full Width & Responsive) -->
        <div v-else class="flex-1 space-y-1 divide-y divide-slate-100">
          <div v-if="loading" class="py-16 text-center text-slate-400 text-sm">
            Memuat transaksi...
          </div>

          <div v-else-if="displayedTransactions.length === 0" class="py-16 text-center text-slate-400 text-sm space-y-3">
            <p class="font-medium">Tidak ada transaksi yang cocok dengan filter ini.</p>
            <div class="flex items-center justify-center space-x-2">
              <button
                v-if="activeMethod !== 'all' || selectedCategoryFilter !== null || searchQuery"
                @click="activeMethod = 'all'; selectedCategoryFilter = null; searchQuery = ''"
                class="text-xs font-bold text-emerald-700 bg-emerald-50 border border-emerald-200 px-3 py-1.5 rounded-full hover:bg-emerald-100 transition-colors"
              >
                Reset Semua Filter
              </button>
            </div>
          </div>

          <!-- Transaction List Row Item -->
          <div
            v-else
            v-for="tx in displayedTransactions"
            :key="tx.id"
            class="py-3 sm:py-3.5 flex items-center justify-between hover:bg-slate-50/80 rounded-2xl px-2 sm:px-3 transition-colors group"
          >
            <!-- Left: Avatar Icon with Method Status Dot & Details -->
            <div class="flex items-center space-x-3 sm:space-x-4">
              <div class="relative">
                <div
                  class="w-11 h-11 rounded-full flex items-center justify-center font-extrabold text-sm shadow-sm"
                  :class="tx.type === 'income' ? 'bg-emerald-100 text-emerald-800' : 'bg-slate-900 text-white'"
                >
                  {{ getCategoryInitial(tx.category?.name || 'T') }}
                </div>
                <!-- Status dot matching payment method -->
                <span
                  class="absolute -bottom-0.5 -right-0.5 w-3.5 h-3.5 rounded-full border-2 border-white"
                  :class="[
                    extractPaymentMethod(tx.note, tx.id) === 'qris' ? 'bg-emerald-400' :
                    extractPaymentMethod(tx.note, tx.id) === 'cash' ? 'bg-amber-400' :
                    extractPaymentMethod(tx.note, tx.id) === 'atm' ? 'bg-cyan-400' :
                    'bg-purple-400'
                  ]"
                ></span>
              </div>

              <!-- Title, Date & Method Badge -->
              <div>
                <div class="text-sm font-bold text-slate-900 flex items-center space-x-2">
                  <span>{{ cleanNote(tx.note) || tx.category?.name || 'Transaksi' }}</span>
                </div>
                <div class="flex items-center space-x-2 text-xs text-slate-400 mt-0.5">
                  <span class="font-mono text-[11px]">{{ tx.occurred_on }}</span>
                  <span>•</span>
                  <span
                    class="text-[9px] font-mono font-bold uppercase px-2 py-0.5 rounded-full border"
                    :class="getMethodBadgeColor(extractPaymentMethod(tx.note, tx.id))"
                  >
                    {{ extractPaymentMethod(tx.note, tx.id).toUpperCase() }}
                  </span>
                  <span v-if="tx.category" class="hidden sm:inline text-slate-500 font-medium">
                    ({{ tx.category.name }})
                  </span>
                </div>
              </div>
            </div>

            <!-- Right: Amount (Green for Income, Bold Black for Expense) -->
            <div class="text-right flex items-center space-x-3">
              <span
                class="font-mono font-black text-sm sm:text-base tracking-tight"
                :class="tx.type === 'income' ? 'text-emerald-600' : 'text-slate-900'"
              >
                {{ tx.type === 'income' ? '+ ' : '- ' }}{{ formatIDR(tx.amount_idr) }}
              </span>

              <!-- Delete Action Button -->
              <button
                @click="handleDeleteTx(tx.id)"
                class="opacity-0 group-hover:opacity-100 text-slate-400 hover:text-rose-600 p-1.5 rounded-lg hover:bg-rose-50 text-xs transition-all"
                title="Hapus transaksi"
              >
                ✕
              </button>
            </div>
          </div>
        </div>
      </section>
    </div>

    <!-- Floating Bottom Dock Navigation -->
    <BottomDock
      :active-tab="activeTab"
      @update:active-tab="activeTab = $event"
      @open-add-modal="openAddExpense"
    />

    <!-- Global Toast Alert -->
    <div
      v-if="toastMessage"
      class="fixed bottom-24 z-50 bg-slate-900 text-white px-5 py-3 rounded-full shadow-2xl border border-white/10 text-xs sm:text-sm font-bold flex items-center space-x-2 animate-in slide-in-from-bottom-5 duration-200"
    >
      <span>{{ toastMessage }}</span>
    </div>

    <!-- Modals -->
    <AddTransactionModal
      :is-open="showAddTxModal"
      :categories="categories"
      @close="showAddTxModal = false"
      @created="onTransactionCreated"
    />

    <UpsertBudgetModal
      :is-open="showBudgetModal"
      :categories="categories"
      :initial-month="currentMonth"
      @close="showBudgetModal = false"
      @saved="onBudgetSaved"
    />

    <AddCategoryModal
      :is-open="showCategoryModal"
      @close="showCategoryModal = false"
      @created="onCategoryCreated"
    />
  </div>
</template>
