<script setup lang="ts">
defineProps<{
  activeTab: 'dashboard' | 'transactions' | 'budgets' | 'categories'
}>()

const emit = defineEmits<{
  (e: 'update:activeTab', tab: 'dashboard' | 'transactions' | 'budgets' | 'categories'): void
  (e: 'open-add-modal'): void
}>()
</script>

<template>
  <div class="fixed bottom-6 inset-x-0 z-40 flex items-center justify-center px-4 pointer-events-none">
    <nav class="pointer-events-auto bg-slate-900/90 text-white backdrop-blur-xl px-4 py-2.5 rounded-full shadow-[0_15px_35px_rgba(0,0,0,0.5)] border border-slate-700/60 flex items-center space-x-2 sm:space-x-3 transition-all duration-300">
      <!-- 1. Home / Transactions -->
      <button
        @click="emit('update:activeTab', 'dashboard')"
        :class="[
          'p-2.5 rounded-full transition-all duration-200 flex items-center justify-center active:scale-90',
          activeTab === 'dashboard'
            ? 'bg-white text-slate-950 shadow-md scale-105'
            : 'text-slate-400 hover:text-white hover:bg-slate-800'
        ]"
        title="Dashboard Utama"
      >
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.2" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6"/>
        </svg>
      </button>

      <!-- 2. Transactions List -->
      <button
        @click="emit('update:activeTab', 'transactions')"
        :class="[
          'p-2.5 rounded-full transition-all duration-200 flex items-center justify-center active:scale-90',
          activeTab === 'transactions'
            ? 'bg-white text-slate-950 shadow-md scale-105'
            : 'text-slate-400 hover:text-white hover:bg-slate-800'
        ]"
        title="Daftar Lengkap Transaksi"
      >
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.2" d="M4 6h16M4 12h16M4 18h16"/>
        </svg>
      </button>

      <!-- 3. Vibrant Plus Button for Quick Add (Catat Transaksi) -->
      <button
        @click="emit('open-add-modal')"
        class="bg-gradient-to-tr from-emerald-500 to-teal-400 hover:from-emerald-400 hover:to-teal-300 text-slate-950 font-black p-3 rounded-full shadow-lg shadow-emerald-500/30 transition-all duration-200 hover:scale-110 active:scale-90 mx-1"
        title="Catat Transaksi Cepat (Hotkey: N)"
      >
        <svg class="w-5 h-5 stroke-[3]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" d="M12 4v16m8-8H4"/>
        </svg>
      </button>

      <!-- 4. Budgets / Analytics -->
      <button
        @click="emit('update:activeTab', 'budgets')"
        :class="[
          'p-2.5 rounded-full transition-all duration-200 flex items-center justify-center active:scale-90',
          activeTab === 'budgets'
            ? 'bg-white text-slate-950 shadow-md scale-105'
            : 'text-slate-400 hover:text-white hover:bg-slate-800'
        ]"
        title="Radar Anggaran (Budgets)"
      >
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"/>
        </svg>
      </button>

      <!-- 5. Categories -->
      <button
        @click="emit('update:activeTab', 'categories')"
        :class="[
          'p-2.5 rounded-full transition-all duration-200 flex items-center justify-center active:scale-90',
          activeTab === 'categories'
            ? 'bg-white text-slate-950 shadow-md scale-105'
            : 'text-slate-400 hover:text-white hover:bg-slate-800'
        ]"
        title="Kelola Kategori"
      >
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.2" d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.994 1.994 0 013 12V7a4 4 0 014-4z"/>
        </svg>
      </button>
    </nav>
  </div>
</template>
