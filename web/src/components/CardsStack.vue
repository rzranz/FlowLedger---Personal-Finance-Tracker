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
    <!-- Active Method Filter Indicator -->
    <div
      v-if="activeMethod !== 'all' && activeCardInfo"
      class="mb-2 flex items-center space-x-2 bg-white/[0.08] backdrop-blur-md border border-white/[0.14] px-3 py-1 rounded-full text-xs shadow-lg animate-in fade-in zoom-in-95 duration-200"
    >
      <span class="w-2 h-2 rounded-full" :style="{ backgroundColor: activeCardInfo.color }"></span>
      <span class="text-slate-300 text-[11px] font-medium">
        Filter: <span class="font-bold text-white">{{ activeCardInfo.shortName }}</span>
        <span class="text-slate-400 font-mono ml-1">({{ activeCardInfo.cardNumber }})</span>
      </span>
      <button
        @click.stop="emit('update:activeMethod', 'all')"
        class="text-slate-400 hover:text-white ml-1 text-xs font-bold px-1.5 py-0.5 rounded-full hover:bg-white/10 transition-colors"
        title="Tampilkan semua transaksi"
      >
        ✕
      </button>
    </div>

    <!-- Peek Cards Deck (Tap to expand full vertical stack) -->
    <div
      @click="emit('toggle-expand')"
      class="group relative h-28 sm:h-32 w-44 sm:w-56 flex items-center justify-end cursor-pointer pr-1 transition-transform duration-300 hover:scale-[1.03] active:scale-95"
      title="Buka daftar kartu pembayaran"
    >
      <!-- Subtitle badge below deck -->
      <div class="absolute -bottom-5 right-1 text-[10px] uppercase tracking-wider font-bold text-slate-400 group-hover:text-emerald-400 transition-colors flex items-center space-x-1">
        <span>Buka Kartu</span>
        <svg class="w-3.5 h-3.5 transition-transform group-hover:translate-x-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M9 5l7 7-7 7"/>
        </svg>
      </div>

      <!-- The Horizontal Peeking Cards Stack (Yellow, Teal, Blue) -->
      <div class="relative w-40 sm:w-52 h-24 sm:h-28 flex items-center justify-end">
        <div
          v-for="(card, index) in PAYMENT_METHODS.slice(0, 3)"
          :key="card.id"
          :style="{
            zIndex: activeMethod === card.id ? 25 : index + 5,
            transform: activeMethod === card.id
              ? 'translateX(-28px) translateY(-4px) scale(1.05) rotate(-2deg)'
              : `translateX(-${(3 - 1 - index) * 20}px) translateY(${index * 2}px) scale(${0.92 + index * 0.035}) rotate(${(index - 1) * 2.5}deg)`,
            boxShadow: `0 10px 24px -4px ${card.glowColor}`,
          }"
          :class="[
            'card-peek-item absolute right-0 top-0 w-36 sm:w-44 h-22 sm:h-26 rounded-2xl p-3 flex flex-col justify-between border border-white/20 select-none cursor-pointer',
            `bg-gradient-to-br ${card.gradient}`,
            activeMethod === card.id
              ? 'ring-2 ring-white/80'
              : 'group-hover:translate-x-[-6px]'
          ]"
        >
          <!-- Specular Shimmer -->
          <div class="absolute inset-0 rounded-2xl bg-gradient-to-tr from-transparent via-white/10 to-white/20 pointer-events-none"></div>

          <!-- Top Row: Card Type & Masked Number -->
          <div class="flex items-center justify-between relative z-10">
            <span class="text-[9px] font-bold tracking-tight text-slate-900/80">Digital card</span>
            <span class="text-[9px] font-mono font-bold text-slate-900/80">{{ card.cardNumber }}</span>
          </div>

          <!-- Middle: Spending preview -->
          <div class="relative z-10">
            <div class="text-[8px] font-bold uppercase tracking-wider text-slate-900/60 font-mono">
              {{ card.shortName }}
            </div>
            <div class="text-xs sm:text-sm font-black font-mono tracking-tight text-slate-950 truncate">
              {{ formatIDR(methodSpendMap[card.id] || 0) }}
            </div>
          </div>

          <!-- Bottom: Overlapping Circles -->
          <div class="flex items-center justify-end relative z-10">
            <div class="flex -space-x-2">
              <span class="w-4 h-4 rounded-full bg-slate-900/25"></span>
              <span class="w-4 h-4 rounded-full bg-slate-900/25"></span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.card-peek-item {
  transform-origin: center right;
  transition: transform 0.35s cubic-bezier(0.16, 1, 0.3, 1), box-shadow 0.3s ease;
}
</style>
