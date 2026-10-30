<script setup lang="ts">
defineProps<{
  activeTab: 'dashboard' | 'transactions' | 'budgets' | 'categories'
  isCardsOpen: boolean
}>()

const emit = defineEmits<{
  (e: 'update:activeTab', tab: 'dashboard' | 'transactions' | 'budgets' | 'categories'): void
  (e: 'toggle-cards'): void
  (e: 'open-add-modal'): void
  (e: 'open-community'): void
}>()
</script>

<template>
  <div class="fixed bottom-6 inset-x-0 z-40 flex items-center justify-center px-4 pointer-events-none">
    <nav class="pointer-events-auto bg-[#181a20]/95 text-white backdrop-blur-2xl p-1.5 rounded-full shadow-[0_20px_45px_rgba(0,0,0,0.65)] border border-white/10 flex items-center space-x-1.5 sm:space-x-2 transition-all duration-300">
      <!-- 1. Community Icon (Users) -->
      <button
        @click="emit('open-community')"
        class="p-2.5 sm:p-3 rounded-full text-slate-400 hover:text-white hover:bg-white/10 transition-all duration-200 active:scale-90 flex items-center justify-center"
        title="Komunitas & Split Bill"
      >
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"/>
        </svg>
      </button>

      <!-- 2. Home Icon (Active White Pill when in Home / Overview) -->
      <button
        @click="emit('update:activeTab', 'dashboard')"
        :class="[
          'transition-all duration-300 flex items-center justify-center active:scale-95',
          !isCardsOpen && activeTab === 'dashboard'
            ? 'bg-white text-slate-950 shadow-lg px-4 sm:px-5 py-2 sm:py-2.5 rounded-full font-bold'
            : 'p-2.5 sm:p-3 rounded-full text-slate-400 hover:text-white hover:bg-white/10'
        ]"
        title="Ringkasan Utama"
      >
        <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
          <path d="M10 20v-6h4v6h5v-8h3L12 3 2 12h3v8z"/>
        </svg>
      </button>

      <!-- 3. Quick Action Plus (+) Button -->
      <button
        @click="emit('open-add-modal')"
        class="bg-gradient-to-tr from-emerald-400 to-teal-400 hover:from-emerald-300 hover:to-teal-200 text-slate-950 font-black p-2.5 sm:p-3 rounded-full shadow-lg shadow-emerald-500/25 transition-all duration-200 hover:scale-105 active:scale-90"
        title="Catat Transaksi Cepat (Hotkey: N)"
      >
        <svg class="w-5 h-5 stroke-[2.8]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" d="M12 4v16m8-8H4"/>
        </svg>
      </button>

      <!-- 4. Cards Icon (Active White Pill when in Cards view) -->
      <button
        @click="emit('toggle-cards')"
        :class="[
          'transition-all duration-300 flex items-center justify-center active:scale-95',
          isCardsOpen
            ? 'bg-white text-slate-950 shadow-lg px-4 sm:px-5 py-2 sm:py-2.5 rounded-full font-bold'
            : 'p-2.5 sm:p-3 rounded-full text-slate-400 hover:text-white hover:bg-white/10'
        ]"
        title="Kartu & Dompet Digital"
      >
        <svg class="w-5 h-5 stroke-[2.2]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <rect x="2" y="5" width="20" height="14" rx="3"/>
          <line x1="2" y1="10" x2="22" y2="10"/>
        </svg>
      </button>
    </nav>
  </div>
</template>
