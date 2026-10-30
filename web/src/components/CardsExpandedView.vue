<script setup lang="ts">
import { ref } from 'vue'
import { PAYMENT_METHODS, type PaymentMethod, type CardMethodInfo, formatIDR } from '../types'

defineProps<{
  isOpen: boolean
  activeMethod: PaymentMethod
  methodSpendMap: Record<PaymentMethod, number>
  totalBalance: number
}>()

const emit = defineEmits<{
  (e: 'select-method', method: PaymentMethod): void
  (e: 'close'): void
  (e: 'order-card'): void
}>()

// Card details modal state
const showDetailsModal = ref(false)
const selectedCardForDetails = ref<CardMethodInfo | null>(null)
const isCopied = ref(false)

// Card settings modal state
const showSettingsModal = ref(false)
const selectedCardForSettings = ref<CardMethodInfo | null>(null)
const dailyLimit = ref(5000000)
const onlineTransactions = ref(true)
const nfcEnabled = ref(true)

// Frozen status tracker
const frozenCards = ref<Record<string, boolean>>({})

// Snackbar notification
const snackbar = ref(false)
const snackbarText = ref('')

function notify(text: string) {
  snackbarText.value = text
  snackbar.value = true
}

function handleSelectCard(method: PaymentMethod) {
  emit('select-method', method)
}

function toggleFreeze(card: CardMethodInfo, event: Event) {
  event.stopPropagation()
  frozenCards.value[card.id] = !frozenCards.value[card.id]
  const isNowFrozen = frozenCards.value[card.id]
  notify(
    isNowFrozen
      ? `❄️ ${card.name} berhasil dibekukan sementara`
      : `✓ ${card.name} telah diaktifkan kembali`
  )
}

function openCardDetails(card: CardMethodInfo, event: Event) {
  event.stopPropagation()
  selectedCardForDetails.value = card
  showDetailsModal.value = true
}

function openCardSettings(card: CardMethodInfo, event: Event) {
  event.stopPropagation()
  selectedCardForSettings.value = card
  showSettingsModal.value = true
}

function copyCardNumber() {
  if (!selectedCardForDetails.value) return
  const fullNumber = '4532 8901 ' + selectedCardForDetails.value.cardNumber.replace('•••• ', '') + ' 9482'
  navigator.clipboard.writeText(fullNumber.replace(/\s/g, ''))
  isCopied.value = true
  notify('✓ Nomor kartu berhasil disalin ke clipboard')
  setTimeout(() => {
    isCopied.value = false
  }, 2000)
}
</script>

<template>
  <div
    class="cards-screen-container transition-all duration-500 ease-[cubic-bezier(0.32,0.72,0,1)] select-none"
    :class="[
      isOpen ? 'opacity-100 translate-y-0 max-h-[2200px] py-2' : 'opacity-0 -translate-y-4 max-h-0 pointer-events-none overflow-hidden'
    ]"
  >
    <!-- Screen Header: "Cards" + "Order a card" Pill Button (Matching Reference Exactly) -->
    <div class="flex items-center justify-between pb-6 pt-2">
      <div class="flex items-center space-x-3">
        <h1 class="text-2xl sm:text-3xl font-black text-white tracking-tight">
          Cards
        </h1>
        <span class="text-xs font-mono font-bold text-slate-400 bg-white/[0.06] border border-white/10 px-2.5 py-0.5 rounded-full">
          {{ PAYMENT_METHODS.length }} Cards
        </span>
      </div>

      <div class="flex items-center space-x-2.5">
        <!-- Order a Card Button -->
        <button
          @click="emit('order-card')"
          class="bg-white hover:bg-slate-100 text-slate-950 font-bold px-4 sm:px-5 py-2 rounded-full text-xs sm:text-sm shadow-lg transition-all duration-200 hover:scale-105 active:scale-95 flex items-center space-x-1.5"
        >
          <span>Order a card</span>
          <span class="text-xs">＋</span>
        </button>

        <!-- Close / Return back button -->
        <button
          @click="emit('close')"
          class="w-9 h-9 rounded-full bg-white/[0.08] hover:bg-white/[0.14] text-slate-300 hover:text-white border border-white/[0.12] flex items-center justify-center transition-all active:scale-90"
          title="Kembali ke Ringkasan"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M6 18L18 6M6 6l12 12"/>
          </svg>
        </button>
      </div>
    </div>

    <!-- Active Filter Reset Pill if currently filtered -->
    <div v-if="activeMethod !== 'all'" class="mb-4 flex items-center justify-between bg-white/[0.05] border border-white/10 rounded-2xl px-4 py-2.5">
      <div class="flex items-center space-x-2 text-xs">
        <span class="w-2 h-2 rounded-full bg-emerald-400 animate-pulse"></span>
        <span class="text-slate-300">Filter sedang aktif untuk kartu:</span>
        <span class="font-bold text-white uppercase">{{ activeMethod }}</span>
      </div>
      <button
        @click="handleSelectCard('all')"
        class="text-xs font-bold text-emerald-400 hover:text-emerald-300 bg-emerald-500/10 px-3 py-1 rounded-full border border-emerald-500/20 transition-colors"
      >
        Tampilkan Semua
      </button>
    </div>

    <!-- Vertically Stacked Cards ("Berjajar ke Bawah") -->
    <div class="space-y-5 sm:space-y-6 pt-1 pb-4">
      <div
        v-for="(card, index) in PAYMENT_METHODS"
        :key="card.id"
        @click="handleSelectCard(card.id)"
        :style="{
          boxShadow: `0 18px 38px -6px ${card.glowColor}`,
          animationDelay: `${index * 80}ms`
        }"
        :class="[
          'relative w-full rounded-[28px] sm:rounded-[32px] p-5 sm:p-7 cursor-pointer overflow-hidden transition-all duration-300 group',
          `bg-gradient-to-br ${card.gradient}`,
          activeMethod === card.id
            ? 'ring-4 ring-white shadow-2xl scale-[1.01]'
            : 'hover:scale-[1.01] hover:brightness-105 active:scale-[0.99]'
        ]"
      >
        <!-- Background Decorative Geometric Arcs (Concentric Ripple) -->
        <div class="absolute -right-16 -top-16 w-64 h-64 rounded-full border-[32px] border-black/[0.06] pointer-events-none"></div>
        <div class="absolute -right-8 -top-8 w-48 h-48 rounded-full border-[24px] border-black/[0.05] pointer-events-none"></div>
        <div class="absolute -left-12 -bottom-12 w-48 h-48 rounded-full bg-white/[0.06] pointer-events-none blur-xl"></div>

        <!-- Specular Highlight Sheen -->
        <div class="absolute inset-0 bg-gradient-to-tr from-transparent via-white/[0.07] to-white/[0.18] pointer-events-none"></div>

        <!-- FROZEN STATE OVERLAY -->
        <div
          v-if="frozenCards[card.id]"
          class="absolute inset-0 bg-slate-950/60 backdrop-blur-md z-30 flex items-center justify-center space-x-2 text-white font-extrabold text-sm tracking-widest uppercase"
        >
          <span class="text-xl">❄️</span>
          <span>Kartu Dibekukan</span>
        </div>

        <!-- CARD TOP ROW: "Digital card" + Card Masked Number -->
        <div class="flex items-center justify-between relative z-10 mb-8 sm:mb-10">
          <div class="flex items-center space-x-2">
            <span class="font-bold text-xs sm:text-sm tracking-tight" :class="card.textColor">
              Digital card
            </span>
            <span
              v-if="activeMethod === card.id"
              class="text-[10px] font-mono font-black uppercase px-2 py-0.5 rounded-full bg-black/20 text-black border border-black/10 backdrop-blur-sm"
            >
              Aktif
            </span>
          </div>

          <div class="font-mono font-bold text-xs sm:text-sm tracking-wider" :class="card.textColor">
            {{ card.cardNumber }}
          </div>
        </div>

        <!-- CARD MIDDLE ROW: Category / Spend Context -->
        <div class="relative z-10 mb-6">
          <div class="text-[10px] font-bold uppercase tracking-wider font-mono opacity-70" :class="card.textColor">
            {{ card.name }}
          </div>
          <div class="text-xl sm:text-2xl font-black font-mono tracking-tight" :class="card.textColor">
            {{ formatIDR(methodSpendMap[card.id] || 0) }}
            <span class="text-xs font-semibold opacity-70 ml-1">terpakai</span>
          </div>
        </div>

        <!-- CARD BOTTOM ROW: 3 Frosted Action Buttons + Mastercard Logo -->
        <div class="flex items-center justify-between relative z-10 pt-2">
          <!-- 3 Frosted Glass Action Buttons -->
          <div class="flex items-center space-x-2.5">
            <!-- 1. Freeze Button (Snowflake) -->
            <button
              @click="toggleFreeze(card, $event)"
              :class="[
                'w-10 h-10 rounded-full flex items-center justify-center backdrop-blur-xl border transition-all duration-200 active:scale-90 shadow-sm',
                frozenCards[card.id]
                  ? 'bg-blue-600 text-white border-blue-400'
                  : 'bg-black/15 hover:bg-black/25 text-slate-900 border-white/20'
              ]"
              :title="frozenCards[card.id] ? 'Aktifkan Kartu' : 'Bekukan Kartu'"
            >
              <svg class="w-4 h-4 stroke-[2.2]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" d="M12 3v18M3 12h18M6.343 6.343l11.314 11.314M6.343 17.657L17.657 6.343M12 8l2-2m-4 0l2 2m0 8l2 2m-4 0l2-2m-4-6l-2 2m0-4l2 2m8 0l2-2m0 4l-2-2"/>
              </svg>
            </button>

            <!-- 2. Card Details Button (123) -->
            <button
              @click="openCardDetails(card, $event)"
              class="w-10 h-10 rounded-full flex items-center justify-center backdrop-blur-xl bg-black/15 hover:bg-black/25 text-slate-900 border border-white/20 transition-all duration-200 active:scale-90 shadow-sm"
              title="Lihat Detail Kartu (CVV & Exp)"
            >
              <svg class="w-4 h-4 stroke-[2.2]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <rect x="2" y="5" width="20" height="14" rx="2"/>
                <path stroke-linecap="round" stroke-linejoin="round" d="M2 10h20M6 14h3"/>
              </svg>
            </button>

            <!-- 3. Settings Button (Gear) -->
            <button
              @click="openCardSettings(card, $event)"
              class="w-10 h-10 rounded-full flex items-center justify-center backdrop-blur-xl bg-black/15 hover:bg-black/25 text-slate-900 border border-white/20 transition-all duration-200 active:scale-90 shadow-sm"
              title="Pengaturan Kartu"
            >
              <svg class="w-4 h-4 stroke-[2.2]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"/>
                <path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/>
              </svg>
            </button>
          </div>

          <!-- Mastercard Overlapping Circles Logo -->
          <div class="flex items-center -space-x-3.5 pr-1">
            <div class="w-9 h-9 rounded-full bg-black/20 backdrop-blur-sm border border-black/10"></div>
            <div class="w-9 h-9 rounded-full bg-black/30 backdrop-blur-sm border border-black/10"></div>
          </div>
        </div>
      </div>
    </div>

    <!-- ======================================================== -->
    <!-- VUETIFY MODALS INTEGRATION                               -->
    <!-- ======================================================== -->

    <!-- 1. Card Details Dialog (Vuetify VDialog) -->
    <v-dialog v-model="showDetailsModal" max-width="420">
      <v-card class="bg-[#0f131c] text-white rounded-3xl border border-white/10 p-6 shadow-2xl">
        <div class="flex items-center justify-between pb-4 border-b border-white/10">
          <div class="flex items-center space-x-2">
            <span class="w-2.5 h-2.5 rounded-full" :style="{ backgroundColor: selectedCardForDetails?.color }"></span>
            <span class="font-extrabold text-base tracking-tight">{{ selectedCardForDetails?.name }}</span>
          </div>
          <button @click="showDetailsModal = false" class="text-slate-400 hover:text-white p-1">✕</button>
        </div>

        <div class="py-5 space-y-4">
          <div>
            <div class="text-[10px] font-mono uppercase tracking-wider text-slate-400">Nomor Kartu Lengkap</div>
            <div class="flex items-center justify-between mt-1 bg-white/[0.05] p-3 rounded-xl border border-white/10">
              <span class="font-mono text-base font-bold tracking-widest text-emerald-400">
                4532 8901 {{ selectedCardForDetails?.cardNumber.replace('•••• ', '') }} 9482
              </span>
              <button
                @click="copyCardNumber"
                class="text-xs bg-white/10 hover:bg-white/20 text-white px-2.5 py-1 rounded-lg transition-colors font-medium"
              >
                {{ isCopied ? 'Tersalin!' : 'Salin' }}
              </button>
            </div>
          </div>

          <div class="grid grid-cols-2 gap-3">
            <div class="bg-white/[0.05] p-3 rounded-xl border border-white/10">
              <div class="text-[10px] font-mono uppercase tracking-wider text-slate-400">Kedaluwarsa</div>
              <div class="font-mono text-sm font-bold text-white mt-0.5">08 / 29</div>
            </div>
            <div class="bg-white/[0.05] p-3 rounded-xl border border-white/10">
              <div class="text-[10px] font-mono uppercase tracking-wider text-slate-400">CVV / CVC</div>
              <div class="font-mono text-sm font-bold text-white mt-0.5">842</div>
            </div>
          </div>

          <div class="bg-white/[0.05] p-3 rounded-xl border border-white/10">
            <div class="text-[10px] font-mono uppercase tracking-wider text-slate-400">Pemegang Kartu</div>
            <div class="font-mono text-sm font-bold text-white mt-0.5">Randi Zakaria</div>
          </div>
        </div>

        <div class="pt-2">
          <v-btn
            block
            color="primary"
            class="rounded-xl font-bold py-2"
            @click="showDetailsModal = false"
          >
            Selesai
          </v-btn>
        </div>
      </v-card>
    </v-dialog>

    <!-- 2. Card Settings Dialog (Vuetify VDialog + Controls) -->
    <v-dialog v-model="showSettingsModal" max-width="440">
      <v-card class="bg-[#0f131c] text-white rounded-3xl border border-white/10 p-6 shadow-2xl">
        <div class="flex items-center justify-between pb-4 border-b border-white/10">
          <div class="font-extrabold text-base tracking-tight">
            Pengaturan {{ selectedCardForSettings?.shortName }}
          </div>
          <button @click="showSettingsModal = false" class="text-slate-400 hover:text-white p-1">✕</button>
        </div>

        <div class="py-4 space-y-4">
          <div>
            <div class="flex justify-between text-xs font-mono mb-2">
              <span class="text-slate-400">Batas Transaksi Harian:</span>
              <span class="text-emerald-400 font-bold">{{ formatIDR(dailyLimit) }}</span>
            </div>
            <v-slider
              v-model="dailyLimit"
              :min="500000"
              :max="25000000"
              :step="500000"
              color="primary"
              track-color="rgba(255,255,255,0.1)"
              hide-details
            ></v-slider>
          </div>

          <div class="flex items-center justify-between py-2 border-t border-white/10">
            <div>
              <div class="text-xs font-bold text-white">Transaksi Online / E-Commerce</div>
              <div class="text-[11px] text-slate-400">Izinkan pembayaran di merchant web</div>
            </div>
            <v-switch v-model="onlineTransactions" color="primary" hide-details></v-switch>
          </div>

          <div class="flex items-center justify-between py-2 border-t border-white/10">
            <div>
              <div class="text-xs font-bold text-white">Contactless / NFC</div>
              <div class="text-[11px] text-slate-400">Tap to pay di mesin EDC fisik</div>
            </div>
            <v-switch v-model="nfcEnabled" color="primary" hide-details></v-switch>
          </div>
        </div>

        <div class="pt-2">
          <v-btn
            block
            color="primary"
            class="rounded-xl font-bold py-2"
            @click="showSettingsModal = false; notify('✓ Pengaturan kartu berhasil diperbarui')"
          >
            Simpan Pengaturan
          </v-btn>
        </div>
      </v-card>
    </v-dialog>

    <!-- 3. Vuetify VSnackbar for tactile notifications -->
    <v-snackbar v-model="snackbar" :timeout="3000" location="top" color="#0f131c" rounded="pill">
      <span class="text-xs font-bold text-white">{{ snackbarText }}</span>
    </v-snackbar>
  </div>
</template>

<style scoped>
.cards-screen-container {
  will-change: transform, opacity, max-height;
}
</style>
