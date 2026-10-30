<script setup lang="ts">
import { ref, computed } from 'vue'
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

const isTouchPressing = ref(false)
const hoveredCardIdx = ref<number | null>(null)

const activeCardInfo = computed(() => {
  return PAYMENT_METHODS.find((m) => m.id === props.activeMethod)
})

function handleDeckClick() {
  if (isTouchPressing.value) return
  isTouchPressing.value = true
  // Small tactile anticipation touch before curtain expands
  setTimeout(() => {
    isTouchPressing.value = false
    emit('toggle-expand')
  }, 150)
}
</script>

<template>
  <div class="relative select-none flex flex-col items-end">
    <!-- Active Method Filter Indicator Pill -->
    <div
      v-if="activeMethod !== 'all' && activeCardInfo"
      class="mb-2.5 flex items-center space-x-2 bg-white/[0.08] backdrop-blur-md border border-white/[0.14] pl-3 pr-2 py-1 rounded-full text-xs shadow-lg transition-all duration-300"
    >
      <span class="w-2 h-2 rounded-full animate-pulse" :style="{ backgroundColor: activeCardInfo.color }"></span>
      <span class="text-slate-300 text-[11px] font-medium">
        Kartu Aktif: <span class="font-bold text-white">{{ activeCardInfo.shortName }}</span>
        <span class="text-slate-400 font-mono ml-1">({{ activeCardInfo.cardNumber }})</span>
      </span>
      <button
        @click.stop="emit('update:activeMethod', 'all')"
        class="tactile-btn text-slate-400 hover:text-white ml-1 p-1 rounded-full hover:bg-white/15 transition-colors"
        title="Tampilkan semua transaksi"
      >
        <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M6 18L18 6M6 6l12 12"/>
        </svg>
      </button>
    </div>

    <!-- Peek Cards Deck (With tactile anticipation press & hover fan-out) -->
    <div
      @click="handleDeckClick"
      @mouseleave="hoveredCardIdx = null"
      :class="[
        'group relative h-32 sm:h-36 w-48 sm:w-60 flex items-center justify-end cursor-pointer pr-1',
        isTouchPressing ? 'scale-[0.93]' : 'hover:scale-[1.02]'
      ]"
      style="transition: transform 0.32s cubic-bezier(0.34, 1.56, 0.64, 1); perspective: 900px;"
      title="Klik untuk membuka dan memilih kartu pembayaran"
    >
      <!-- Ambient Touch Ripple Glow on Press -->
      <div
        class="absolute inset-0 rounded-3xl pointer-events-none transition-opacity duration-300"
        :class="isTouchPressing ? 'opacity-100' : 'opacity-0 group-hover:opacity-60'"
        style="background: radial-gradient(circle at 75% 50%, rgba(250, 204, 21, 0.18), rgba(45, 212, 191, 0.12), transparent 70%);"
      ></div>

      <!-- Subtitle badge below deck -->
      <div class="absolute -bottom-5 right-1 text-[10px] uppercase tracking-wider font-bold text-slate-400 group-hover:text-white transition-all duration-300 flex items-center space-x-1.5">
        <span>Ketuk Dompet Kartu</span>
        <span class="w-4 h-4 rounded-full bg-white/[0.08] group-hover:bg-white/20 flex items-center justify-center transition-transform duration-300 group-hover:translate-x-1">
          <svg class="w-2.5 h-2.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M9 5l7 7-7 7"/>
          </svg>
        </span>
      </div>

      <!-- The Horizontal Peeking Cards Stack (Yellow, Teal, Blue) -->
      <div class="relative w-44 sm:w-56 h-28 sm:h-32 flex items-center justify-end">
        <div
          v-for="(card, index) in PAYMENT_METHODS.slice(0, 3)"
          :key="card.id"
          @mouseenter="hoveredCardIdx = index"
          :style="{
            zIndex: activeMethod === card.id ? 25 : index + 5,
            transform: isTouchPressing
              ? `translateX(-${(2 - index) * 11}px) translateY(${index * 4}px) scale(0.94) rotate(${(index - 1) * 1.5}deg)`
              : activeMethod === card.id
                ? 'translateX(-32px) translateY(-6px) scale(1.05) rotate(-2.5deg)'
                : hoveredCardIdx !== null
                  ? `translateX(-${(2 - index) * 28 + (hoveredCardIdx === index ? 8 : 0)}px) translateY(${hoveredCardIdx === index ? -6 : index * 2}px) scale(${0.94 + index * 0.03 + (hoveredCardIdx === index ? 0.03 : 0)}) rotate(${(index - 1) * 3.5}deg)`
                  : `translateX(-${(2 - index) * 22}px) translateY(${index * 2}px) scale(${0.92 + index * 0.035}) rotate(${(index - 1) * 2.2}deg)`,
            boxShadow: `0 14px 30px -6px ${card.glowColor}`,
          }"
          :class="[
            'card-peek-item absolute right-0 top-1 w-38 sm:w-46 h-24 sm:h-28 rounded-[22px] p-3.5 flex flex-col justify-between border border-white/25 select-none cursor-pointer',
            `bg-gradient-to-br ${card.gradient}`,
            activeMethod === card.id ? 'ring-2 ring-white/90' : ''
          ]"
        >
          <!-- Specular Shimmer -->
          <div class="absolute inset-0 rounded-[22px] bg-gradient-to-tr from-transparent via-white/10 to-white/25 pointer-events-none"></div>

          <!-- Top Row: Card Type & Masked Number -->
          <div class="flex items-center justify-between relative z-10">
            <span class="text-[9px] font-bold tracking-tight" :class="card.textColor">Digital card</span>
            <span class="text-[9px] font-mono font-bold opacity-85" :class="card.textColor">{{ card.cardNumber }}</span>
          </div>

          <!-- Middle: Spending preview -->
          <div class="relative z-10 my-auto">
            <div class="text-[8px] font-bold uppercase tracking-wider font-mono opacity-70" :class="card.textColor">
              {{ card.shortName }}
            </div>
            <div class="text-xs sm:text-sm font-black font-mono tracking-tight truncate" :class="card.textColor">
              {{ formatIDR(methodSpendMap[card.id] || 0) }}
            </div>
          </div>

          <!-- Bottom: Overlapping Circles -->
          <div class="flex items-center justify-between relative z-10">
            <span class="w-1.5 h-1.5 rounded-full bg-current opacity-40" :class="card.textColor"></span>
            <div class="flex -space-x-2">
              <span class="w-4 h-4 rounded-full bg-slate-950/25 backdrop-blur-sm"></span>
              <span class="w-4 h-4 rounded-full bg-slate-950/35 backdrop-blur-sm"></span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.card-peek-item {
  width: 9.75rem;
  height: 6.25rem;
  transform-origin: bottom right;
  transition:
    transform 0.48s cubic-bezier(0.34, 1.56, 0.64, 1),
    box-shadow 0.35s cubic-bezier(0.22, 1, 0.36, 1);
  will-change: transform;
}

@media (min-width: 640px) {
  .card-peek-item {
    width: 11.5rem;
    height: 7rem;
  }
}
</style>
