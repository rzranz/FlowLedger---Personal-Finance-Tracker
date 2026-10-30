<script setup lang="ts">
defineProps<{
  activeTab: 'dashboard' | 'transactions' | 'budgets' | 'categories'
  isReady: boolean
}>()

const emit = defineEmits<{
  (e: 'update:activeTab', tab: 'dashboard' | 'transactions' | 'budgets' | 'categories'): void
  (e: 'open-add-modal'): void
}>()

const navItems = [
  { id: 'dashboard', label: 'Dashboard' },
  { id: 'transactions', label: 'Transaksi' },
  { id: 'budgets', label: 'Anggaran' },
  { id: 'categories', label: 'Kategori' },
] as const
</script>

<template>
  <header class="bg-white border-b border-slate-200 sticky top-0 z-30">
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
      <div class="flex items-center justify-between h-16">
        <!-- Brand / Logo -->
        <div class="flex items-center space-x-3">
          <div class="w-9 h-9 rounded-xl bg-gradient-to-tr from-emerald-600 to-teal-400 flex items-center justify-center text-white shadow-sm shadow-emerald-500/20 font-bold text-lg">
            FL
          </div>
          <div>
            <div class="flex items-center space-x-2">
              <span class="font-extrabold text-slate-900 tracking-tight text-lg">FlowLedger</span>
              <span class="text-[10px] font-bold uppercase tracking-wider bg-emerald-50 text-emerald-700 px-1.5 py-0.5 rounded border border-emerald-200">v2.0</span>
            </div>
            <p class="text-xs text-slate-500 hidden sm:block">Personal Finance Engine</p>
          </div>
        </div>

        <!-- Navigation Tabs -->
        <nav class="flex items-center space-x-1 sm:space-x-2">
          <button
            v-for="item in navItems"
            :key="item.id"
            @click="emit('update:activeTab', item.id)"
            :class="[
              'px-3 py-2 rounded-lg text-sm font-medium transition-colors',
              activeTab === item.id
                ? 'bg-slate-900 text-white shadow-sm'
                : 'text-slate-600 hover:text-slate-900 hover:bg-slate-100'
            ]"
          >
            {{ item.label }}
          </button>
        </nav>

        <!-- Actions -->
        <div class="flex items-center space-x-3">
          <!-- Readiness Indicator -->
          <div class="hidden md:flex items-center space-x-1.5 text-xs text-slate-500">
            <span
              class="w-2 h-2 rounded-full"
              :class="isReady ? 'bg-emerald-500 ring-4 ring-emerald-100' : 'bg-amber-500 animate-pulse'"
            />
            <span>{{ isReady ? 'SQLite WAL Ready' : 'Connecting...' }}</span>
          </div>

          <!-- Quick Add Button with 'N' hotkey -->
          <button
            @click="emit('open-add-modal')"
            class="inline-flex items-center space-x-2 bg-emerald-600 hover:bg-emerald-700 text-white px-3.5 py-2 rounded-lg text-sm font-semibold shadow-sm transition-all active:scale-95"
            title="Tekan tombol 'N' di keyboard untuk akses cepat"
          >
            <span>+ Catat Transaksi</span>
            <kbd class="hidden sm:inline-block bg-emerald-800/60 text-emerald-100 text-[11px] font-mono px-1.5 py-0.5 rounded">N</kbd>
          </button>
        </div>
      </div>
    </div>
  </header>
</template>
