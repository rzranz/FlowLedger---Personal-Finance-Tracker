<script setup lang="ts">
import { ref } from 'vue'
import { PAYMENT_METHODS, type PaymentMethod, formatIDR } from '../types'

defineProps<{
  isOpen: boolean
  activeMethod: PaymentMethod
  methodSpendMap: Record<PaymentMethod, number>
  totalBalance: number
}>()

const emit = defineEmits<{
  (e: 'select-method', method: PaymentMethod): void
  (e: 'close'): void
}>()

const selectedPending = ref<PaymentMethod | null>(null)

function handleCardClick(method: PaymentMethod) {
  selectedPending.value = method
  // Smooth tactile delay before folding back down
  setTimeout(() => {
    emit('select-method', method)
    selectedPending.value = null
  }, 220)
}
</script>

<template>
  <div
    class="card-curtain-wrapper overflow-hidden transition-all duration-500 ease-[cubic-bezier(0.32,0.72,0,1)]"
    :class="[
      isOpen ? 'max-h-[1400px] opacity-100 py-6 sm:py-8' : 'max-h-0 opacity-0 py-0 pointer-events-none'
    ]"
  >
    <div class="space-y-6">
      <!-- Expanded Drawer Header -->
      <div class="flex items-center justify-between border-b border-white/[0.08] pb-4">
        <div>
          <div class="text-[11px] font-bold uppercase tracking-widest text-slate-400 font-mono">
            Dompet &amp; Metode Pembayaran
          </div>
          <h3 class="text-lg sm:text-xl font-extrabold text-white tracking-tight mt-0.5">
            Pilih Kartu untuk Memfilter Pengeluaran
          </h3>
        </div>

        <div class="flex items-center space-x-2">
          <!-- Reset to all button -->
          <button
            @click="handleCardClick('all')"
            :class="[
              'px-3 py-1.5 rounded-xl text-xs font-semibold transition-all border',
              activeMethod === 'all'
                ? 'bg-white text-slate-950 border-white'
                : 'bg-white/[0.05] text-slate-300 hover:text-white border-white/[0.1]'
            ]"
          >
            Semua Kartu
          </button>

          <!-- Close / Kembali button -->
          <button
            @click="emit('close')"
            class="w-8 h-8 rounded-xl bg-white/[0.06] hover:bg-white/[0.12] text-slate-300 hover:text-white border border-white/[0.1] flex items-center justify-center transition-all active:scale-90"
            title="Tutup dompet"
          >
            ✕
          </button>
        </div>
      </div>

      <!-- Vertical Cascading Cards Stack ("Berjajar ke Bawah") -->
      <div class="grid grid-cols-1 md:grid-cols-2 gap-4 sm:gap-5 pt-2">
        <div
          v-for="(card, index) in PAYMENT_METHODS"
          :key="card.id"
          @click="handleCardClick(card.id)"
          :style="{
            animationDelay: `${index * 60}ms`
          }"
          :class="[
            'group relative rounded-3xl p-5 sm:p-6 cursor-pointer border select-none transition-all duration-300',
            `bg-gradient-to-br ${card.gradient}`,
            activeMethod === card.id || selectedPending === card.id
              ? 'ring-4 ring-white/50 shadow-2xl scale-[1.02] border-white/40'
              : 'border-white/15 opacity-90 hover:opacity-100 hover:scale-[1.01] hover:shadow-xl'
          ]"
        >
          <!-- Specular Light Effect -->
          <div class="absolute inset-0 rounded-3xl bg-gradient-to-tr from-transparent via-white/[0.07] to-white/[0.16] pointer-events-none"></div>

          <!-- Card Top: Chip & Status -->
          <div class="flex items-center justify-between relative z-10 mb-6">
            <div class="flex items-center space-x-3">
              <!-- Golden EMV Chip -->
              <div class="w-10 h-7 rounded-lg bg-gradient-to-br from-amber-200 to-amber-400 border border-amber-600/60 p-0.5 shadow-md flex items-center justify-center">
                <div class="w-full h-full border border-amber-800/40 rounded grid grid-cols-2 gap-0.5 opacity-60">
                  <div class="border-r border-amber-800/40"></div>
                  <div></div>
                </div>
              </div>

              <!-- Contactless Wave -->
              <span class="text-slate-900/60 text-xs font-bold tracking-tighter">
                <span>)</span><span>)</span><span>)</span>
              </span>
            </div>

            <!-- Active / Select Status Indicator -->
            <div
              :class="[
                'text-[10px] font-extrabold uppercase tracking-wider px-2.5 py-1 rounded-full border backdrop-blur-sm font-mono',
                activeMethod === card.id
                  ? 'bg-black text-white border-white/20 shadow-md'
                  : 'bg-black/20 text-slate-900 border-black/10 group-hover:bg-black/30 group-hover:text-white'
              ]"
            >
              {{ activeMethod === card.id ? '✓ Kartu Aktif' : 'Pilih Kartu' }}
            </div>
          </div>

          <!-- Card Center: Spend Info -->
          <div class="relative z-10 mb-6 space-y-1">
            <div class="text-[10px] font-extrabold uppercase tracking-widest text-slate-900/65 font-mono">
              Total Pengeluaran Bulan Ini
            </div>
            <div class="text-2xl sm:text-3xl font-black font-mono tracking-tight text-slate-950">
              {{ formatIDR(methodSpendMap[card.id] || 0) }}
            </div>
          </div>

          <!-- Card Bottom: Cardholder Name, Masked Number & Brand -->
          <div class="flex items-end justify-between relative z-10 pt-2 border-t border-slate-900/10">
            <div>
              <div class="text-[9px] uppercase tracking-wider font-mono text-slate-900/60 font-semibold">
                Pemegang Kartu
              </div>
              <div class="text-xs font-black tracking-wide text-slate-950 uppercase font-mono">
                Randi Zakaria
              </div>
            </div>

            <div class="text-right">
              <div class="text-xs font-black text-slate-950 font-mono tracking-widest">
                •••• {{ 4242 + index * 123 }}
              </div>
              <div class="text-[10px] font-black uppercase tracking-wider text-slate-900/75">
                {{ card.name }}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.card-curtain-wrapper {
  transform-origin: top center;
  will-change: max-height, opacity;
}
</style>
