<script setup lang="ts">
import { computed } from 'vue'
import { PAYMENT_METHODS, type PaymentMethod, formatIDR } from '../types'

const props = defineProps<{
  activeMethod: PaymentMethod
  methodSpendMap: Record<PaymentMethod, number>
  isExpanded: boolean
}>()

const emit = defineEmits<{
  (e: 'update:activeMethod', method: PaymentMethod): void
  (e: 'toggle-expand'): void
}>()

const activeCardInfo = computed(() => {
  return PAYMENT_METHODS.find((m) => m.id === props.activeMethod)
})
</script>

<template>
  <div class="relative select-none flex flex-col items-end">
    <!-- Active Indicator Pill (when filtered) -->
    <div
      v-if="activeMethod !== 'all' && activeCardInfo"
      class="mb-2 flex items-center space-x-2 bg-white/[0.06] backdrop-blur-md border border-white/[0.12] px-3 py-1 rounded-full text-xs animate-fadeIn"
    >
      <span class="w-1.5 h-1.5 rounded-full" :style="{ backgroundColor: activeCardInfo.color }"></span>
      <span class="text-slate-200 text-[11px] font-medium">
        Kartu: <span class="font-semibold text-white">{{ activeCardInfo.name }}</span>
      </span>
      <button
        @click.stop="emit('update:activeMethod', 'all')"
        class="text-slate-400 hover:text-white ml-1 text-[11px] font-bold px-1 hover:bg-white/10 rounded transition-colors"
      >
        ✕
      </button>
    </div>

    <!-- Interactive Compact Deck (Clicking expands to full vertical list) -->
    <div
      @click="emit('toggle-expand')"
      class="group relative h-32 sm:h-36 w-48 sm:w-60 flex items-center justify-end cursor-pointer pr-1"
      title="Klik untuk membuka & memilih kartu pembayaran"
    >
      <!-- "Tap to open cards" subtle hint -->
      <div class="absolute -bottom-5 right-1 text-[10px] uppercase tracking-wider font-semibold text-slate-500 group-hover:text-slate-300 transition-colors flex items-center space-x-1">
        <span>Buka Dompet</span>
        <svg class="w-3 h-3 transition-transform group-hover:translate-x-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"/>
        </svg>
      </div>

      <!-- The Stacked Cards (Apple Wallet Peek Style) -->
      <div class="relative w-44 sm:w-52 h-28 sm:h-32 flex items-center justify-end">
        <div
          v-for="(card, index) in PAYMENT_METHODS"
          :key="card.id"
          :style="{
            zIndex: activeMethod === card.id ? 20 : index + 5,
            transform: activeMethod === card.id
              ? 'translateX(-32px) translateY(-6px) scale(1.04) rotate(-1deg)'
              : `translateX(-${(PAYMENT_METHODS.length - 1 - index) * 22}px) translateY(${index * 1.5}px) scale(${0.92 + index * 0.025}) rotate(${(index - 1.5) * 2}deg)`,
          }"
          :class="[
            'card-deck-item absolute right-0 top-0 w-40 sm:w-48 h-26 sm:h-30 rounded-2xl p-3.5 flex flex-col justify-between border border-white/[0.16] shadow-xl transition-all duration-300',
            `bg-gradient-to-br ${card.gradient}`,
            activeMethod === card.id
              ? 'ring-2 ring-white/60 shadow-[0_15px_30px_rgba(0,0,0,0.6)]'
              : 'group-hover:translate-x-[-8px] opacity-90 group-hover:opacity-100'
          ]"
        >
          <!-- Specular Highlight Overlay -->
          <div class="absolute inset-0 rounded-2xl bg-gradient-to-tr from-transparent via-white/[0.08] to-white/[0.18] pointer-events-none"></div>

          <!-- Card Header: Gold Chip & Method Label -->
          <div class="flex items-center justify-between relative z-10">
            <!-- Simulated EMV Chip -->
            <div class="w-7 h-5 rounded-md bg-gradient-to-br from-amber-200 to-amber-400 border border-amber-600/50 p-0.5 shadow-sm flex items-center justify-center">
              <div class="w-full h-full border border-amber-800/30 rounded-[2px] grid grid-cols-2 gap-0.5 opacity-60">
                <div class="border-r border-amber-800/30"></div>
                <div></div>
              </div>
            </div>

            <!-- Card Badge -->
            <span class="text-[9px] font-black uppercase tracking-wider bg-black/30 backdrop-blur-sm text-white px-2 py-0.5 rounded-full border border-white/10 font-mono">
              {{ card.shortName }}
            </span>
          </div>

          <!-- Total Spend -->
          <div class="relative z-10">
            <div class="text-[8px] font-bold uppercase tracking-wider text-slate-900/60 font-mono">
              Pengeluaran
            </div>
            <div class="text-xs sm:text-sm font-black font-mono tracking-tight text-slate-950 truncate">
              {{ formatIDR(methodSpendMap[card.id] || 0) }}
            </div>
          </div>

          <!-- Card Footer -->
          <div class="flex items-center justify-between text-[9px] font-mono text-slate-900/75 font-semibold relative z-10">
            <span>•••• {{ 2040 + index * 128 }}</span>
            <span class="text-[11px] font-bold">)))</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.card-deck-item {
  transform-origin: center right;
  transition: transform 0.4s cubic-bezier(0.16, 1, 0.3, 1), box-shadow 0.3s ease, opacity 0.2s ease;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(-4px); }
  to { opacity: 1; transform: translateY(0); }
}

.animate-fadeIn {
  animation: fadeIn 0.25s cubic-bezier(0.16, 1, 0.3, 1) forwards;
}
</style>
