<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import { api } from '../api/client'
import type { Category, Transaction } from '../types'
import { formatIDR } from '../types'

const props = defineProps<{
  categories: Category[]
  currentMonth: string
}>()

const emit = defineEmits<{
  (e: 'open-add-modal'): void
}>()

const transactions = ref<Transaction[]>([])
const loading = ref(false)
const total = ref(0)
const totalPages = ref(1)

// Filters
const page = ref(1)
const limit = ref(15)
const selectedMonth = ref(props.currentMonth)
const selectedType = ref<string>('')
const selectedCategoryId = ref<number | ''>('')
const searchQuery = ref<string>('')

let searchDebounceTimer: any = null

async function loadTransactions() {
  loading.value = true
  try {
    const res = await api.getTransactions({
      month: selectedMonth.value || undefined,
      type: selectedType.value || undefined,
      category_id: selectedCategoryId.value ? Number(selectedCategoryId.value) : undefined,
      q: searchQuery.value.trim() || undefined,
      page: page.value,
      limit: limit.value,
    })
    transactions.value = res.items || []
    total.value = res.pagination.total
    totalPages.value = res.pagination.total_pages || 1
  } catch (err) {
    console.error('Failed to load transactions:', err)
  } finally {
    loading.value = false
  }
}

watch([selectedMonth, selectedType, selectedCategoryId], () => {
  page.value = 1
  loadTransactions()
})

function onSearchInput() {
  clearTimeout(searchDebounceTimer)
  searchDebounceTimer = setTimeout(() => {
    page.value = 1
    loadTransactions()
  }, 300)
}

function changePage(newPage: number) {
  if (newPage >= 1 && newPage <= totalPages.value) {
    page.value = newPage
    loadTransactions()
  }
}

async function handleDelete(id: number) {
  if (!confirm('Apakah Anda yakin ingin menghapus transaksi ini?')) return
  try {
    await api.deleteTransaction(id)
    loadTransactions()
  } catch (err: any) {
    alert(err?.detail || 'Gagal menghapus transaksi')
  }
}

function handleExportCSV() {
  const url = api.getExportCsvUrl({
    month: selectedMonth.value || undefined,
    type: selectedType.value || undefined,
    category_id: selectedCategoryId.value ? Number(selectedCategoryId.value) : undefined,
    q: searchQuery.value.trim() || undefined,
  })
  window.open(url, '_blank')
}

onMounted(() => {
  loadTransactions()
})
</script>

<template>
  <div class="space-y-6">
    <!-- Header & Action Bar -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 bg-white p-4 sm:p-6 rounded-2xl border border-slate-200 shadow-sm">
      <div>
        <h1 class="text-xl sm:text-2xl font-extrabold text-slate-900 tracking-tight">Eksplorasi Transaksi</h1>
        <p class="text-xs sm:text-sm text-slate-500">Pencarian terindeks dan filter transaksi finansial deterministik</p>
      </div>

      <div class="flex items-center space-x-3">
        <button
          @click="handleExportCSV"
          class="inline-flex items-center space-x-1.5 px-3.5 py-2 border border-slate-300 rounded-xl text-sm font-semibold text-slate-700 bg-white hover:bg-slate-50 shadow-sm transition-colors"
          title="Download data sebagai CSV (Streaming RFC 4180)"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"/>
          </svg>
          <span>Export CSV</span>
        </button>
        <button
          @click="emit('open-add-modal')"
          class="px-4 py-2 bg-emerald-600 hover:bg-emerald-700 text-white rounded-xl text-sm font-semibold shadow-sm transition-all"
        >
          + Catat Transaksi
        </button>
      </div>
    </div>

    <!-- Filters Bar -->
    <div class="bg-white p-4 rounded-2xl border border-slate-200 shadow-sm space-y-3">
      <div class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-4 gap-3">
        <!-- Search Notes -->
        <div>
          <label class="block text-[11px] font-bold uppercase tracking-wider text-slate-500 mb-1">Cari Catatan</label>
          <input
            v-model="searchQuery"
            @input="onSearchInput"
            type="text"
            placeholder="Ketik catatan transaksi..."
            class="w-full px-3 py-2 border border-slate-300 rounded-xl text-sm focus:ring-2 focus:ring-emerald-500"
          />
        </div>

        <!-- Month Filter -->
        <div>
          <label class="block text-[11px] font-bold uppercase tracking-wider text-slate-500 mb-1">Bulan</label>
          <input
            v-model="selectedMonth"
            type="month"
            class="w-full px-3 py-2 border border-slate-300 rounded-xl text-sm focus:ring-2 focus:ring-emerald-500"
          />
        </div>

        <!-- Type Filter -->
        <div>
          <label class="block text-[11px] font-bold uppercase tracking-wider text-slate-500 mb-1">Tipe</label>
          <select
            v-model="selectedType"
            class="w-full px-3 py-2 border border-slate-300 rounded-xl text-sm focus:ring-2 focus:ring-emerald-500"
          >
            <option value="">Semua Tipe</option>
            <option value="income">Pemasukan (Income)</option>
            <option value="expense">Pengeluaran (Expense)</option>
          </select>
        </div>

        <!-- Category Filter -->
        <div>
          <label class="block text-[11px] font-bold uppercase tracking-wider text-slate-500 mb-1">Kategori</label>
          <select
            v-model="selectedCategoryId"
            class="w-full px-3 py-2 border border-slate-300 rounded-xl text-sm focus:ring-2 focus:ring-emerald-500"
          >
            <option value="">Semua Kategori</option>
            <option v-for="cat in categories" :key="cat.id" :value="cat.id">
              {{ cat.name }} ({{ cat.type }})
            </option>
          </select>
        </div>
      </div>
    </div>

    <!-- Transactions Data Table -->
    <div class="bg-white rounded-2xl border border-slate-200 shadow-sm overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full text-left text-sm">
          <thead class="bg-slate-50 text-slate-500 text-xs uppercase tracking-wider font-bold border-b border-slate-200">
            <tr>
              <th class="py-3 px-4">Tanggal</th>
              <th class="py-3 px-4">Tipe</th>
              <th class="py-3 px-4">Kategori</th>
              <th class="py-3 px-4">Catatan</th>
              <th class="py-3 px-4 text-right">Nominal (IDR)</th>
              <th class="py-3 px-4 text-center">Aksi</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100">
            <tr v-if="loading">
              <td colspan="6" class="text-center py-12 text-slate-400">
                Memuat data transaksi...
              </td>
            </tr>
            <tr v-else-if="transactions.length === 0">
              <td colspan="6" class="text-center py-12 text-slate-400">
                Tidak ada transaksi yang cocok dengan filter.
              </td>
            </tr>
            <tr
              v-else
              v-for="tx in transactions"
              :key="tx.id"
              class="hover:bg-slate-50/70 transition-colors"
            >
              <td class="py-3 px-4 font-mono text-xs text-slate-600 whitespace-nowrap">
                {{ tx.occurred_on }}
              </td>
              <td class="py-3 px-4 whitespace-nowrap">
                <span
                  class="text-[11px] font-bold uppercase px-2 py-0.5 rounded-full"
                  :class="tx.type === 'income' ? 'bg-emerald-100 text-emerald-800' : 'bg-rose-100 text-rose-800'"
                >
                  {{ tx.type === 'income' ? 'Pemasukan' : 'Pengeluaran' }}
                </span>
              </td>
              <td class="py-3 px-4 font-semibold text-slate-800 whitespace-nowrap">
                {{ tx.category?.name || '-' }}
              </td>
              <td class="py-3 px-4 text-slate-600 max-w-xs truncate">
                {{ tx.note || '-' }}
              </td>
              <td
                class="py-3 px-4 text-right font-mono font-bold whitespace-nowrap"
                :class="tx.type === 'income' ? 'text-emerald-700' : 'text-slate-900'"
              >
                {{ tx.type === 'income' ? '+' : '-' }}{{ formatIDR(tx.amount_idr) }}
              </td>
              <td class="py-3 px-4 text-center whitespace-nowrap">
                <button
                  @click="handleDelete(tx.id)"
                  class="text-xs text-rose-500 hover:text-rose-700 font-semibold px-2 py-1 rounded hover:bg-rose-50 transition-colors"
                >
                  Hapus
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Pagination Footer -->
      <div class="px-4 py-3 border-t border-slate-100 flex items-center justify-between text-xs text-slate-500 bg-slate-50/50">
        <div>
          Menampilkan total <b>{{ total }}</b> transaksi (Halaman {{ page }} dari {{ totalPages }})
        </div>
        <div class="flex items-center space-x-2">
          <button
            @click="changePage(page - 1)"
            :disabled="page <= 1"
            class="px-3 py-1.5 border border-slate-200 rounded-lg bg-white disabled:opacity-40 hover:bg-slate-50 font-medium"
          >
            ← Sebelumnya
          </button>
          <button
            @click="changePage(page + 1)"
            :disabled="page >= totalPages"
            class="px-3 py-1.5 border border-slate-200 rounded-lg bg-white disabled:opacity-40 hover:bg-slate-50 font-medium"
          >
            Berikutnya →
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
