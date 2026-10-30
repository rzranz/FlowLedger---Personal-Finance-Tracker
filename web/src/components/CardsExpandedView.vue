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

// Tactile selection feedback state
const selectingCardId = ref<PaymentMethod | null>(null)
const pressingActionKey = ref<string | null>(null)

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

function triggerButtonPulse(key: string) {
  pressingActionKey.value = key
  setTimeout(() => {
    if (pressingActionKey.value === key) {
      pressingActionKey.value = null
    }
  }, 240)
}

function handleSelectCard(method: PaymentMethod) {
  if (selectingCardId.value !== null) return
  selectingCardId.value = method
  // Give tactile visual confirmation pulse before smoothly folding back
  setTimeout(() => {
    emit('select-method', method)
    setTimeout(() => {
      selectingCardId.value = null
    }, 300)
  }, 320)
}

function toggleFreeze(card: CardMethodInfo, event: Event) {
  event.stopPropagation()
  triggerButtonPulse(`freeze-${card.id}`)
  frozenCards.value[card.id] = !frozenCards.value[card.id]
  const isNowFrozen = frozenCards.value[card.id]
  notify(
    isNowFrozen
      ? `${card.name} berhasil dibekukan sementara`
      : `${card.name} telah diaktifkan kembali`
  )
}

function openCardDetails(card: CardMethodInfo, event: Event) {
  event.stopPropagation()
  triggerButtonPulse(`details-${card.id}`)
  selectedCardForDetails.value = card
  showDetailsModal.value = true
}

function openCardSettings(card: CardMethodInfo, event: Event) {
  event.stopPropagation()
  triggerButtonPulse(`settings-${card.id}`)
  selectedCardForSettings.value = card
  showSettingsModal.value = true
}

function copyCardNumber() {
  if (!selectedCardForDetails.value) return
  const fullNumber = '4532 8901 ' + selectedCardForDetails.value.cardNumber.replace('•••• ', '') + ' 9482'
  navigator.clipboard.writeText(fullNumber.replace(/\s/g, ''))
  isCopied.value = true
  notify('Nomor kartu berhasil disalin ke clipboard')
  setTimeout(() => {
    isCopied.value = false
  }, 2000)
}
</script>

<template>
  <!-- Smooth CSS Grid Curtain Wrapper (No abrupt max-height jump) -->
  <div
    class="cards-curtain-grid select-none"
    :class="isOpen ? 'is-open' : 'is-closed'"
  >
    <div class="overflow-hidden">
      <div
        class="pt-2 pb-6 transition-all duration-500"
        :class="isOpen ? 'opacity-100 translate-y-0' : 'opacity-0 -translate-y-4 pointer-events-none'"
      >
        <!-- Screen Header: "Cards" + "Order a card" Pill Button -->
        <div class="flex items-center justify-between pb-6 pt-1">
          <div class="flex items-center space-x-3">
            <h1 class="text-2xl sm:text-3xl font-black text-white tracking-tight">
              Cards
            </h1>
            <span class="text-[11px] font-mono font-bold text-slate-300 bg-white/[0.06] border border-white/10 px-2.5 py-0.5 rounded-full">
              {{ PAYMENT_METHODS.length }} Active
            </span>
          </div>

          <div class="flex items-center space-x-2.5">
            <!-- Order a Card Button -->
            <button
              @click="emit('order-card')"
              class="tactile-btn bg-white hover:bg-slate-100 text-slate-950 font-bold px-4 sm:px-5 py-2 rounded-full text-xs sm:text-sm shadow-[0_8px_20px_rgba(255,255,255,0.14)] flex items-center space-x-2"
            >
              <span>Order a card</span>
              <svg class="w-3.5 h-3.5 stroke-[2.5]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" d="M12 4v16m8-8H4"/>
              </svg>
            </button>

            <!-- Close / Return back button -->
            <button
              @click="emit('close')"
              class="tactile-btn w-9 h-9 rounded-full bg-white/[0.08] hover:bg-white/[0.14] text-slate-300 hover:text-white border border-white/[0.12] flex items-center justify-center"
              title="Kembali ke Ringkasan"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.2" d="M6 18L18 6M6 6l12 12"/>
              </svg>
            </button>
          </div>
        </div>

        <!-- Active Filter Reset Pill if currently filtered -->
        <div
          v-if="activeMethod !== 'all'"
          class="mb-5 flex items-center justify-between bg-white/[0.05] border border-white/10 rounded-2xl px-4 py-3 backdrop-blur-md"
        >
          <div class="flex items-center space-x-2.5 text-xs">
            <span class="w-2 h-2 rounded-full bg-emerald-400 animate-ping"></span>
            <span class="text-slate-300">Menampilkan transaksi untuk kartu:</span>
            <span class="font-bold text-white uppercase font-mono">{{ activeMethod }}</span>
          </div>
          <button
            @click="handleSelectCard('all')"
            class="tactile-btn text-xs font-bold text-emerald-300 hover:text-white bg-emerald-500/15 hover:bg-emerald-500/25 px-3.5 py-1.5 rounded-full border border-emerald-500/30 transition-colors"
          >
            Semua Kartu
          </button>
        </div>

        <!-- Vertically Stacked Cards ("Berjajar ke Bawah" with Staggered Deal Choreography) -->
        <div class="space-y-6 sm:space-y-7 pt-1 pb-6 px-0.5" style="perspective: 1200px;">
          <div
            v-for="(card, index) in PAYMENT_METHODS"
            :key="`${card.id}-${isOpen ? 'open' : 'closed'}`"
            @click="handleSelectCard(card.id)"
            :style="{
              boxShadow: selectingCardId === card.id
                ? `0 24px 52px -4px ${card.glowColor}`
                : `0 18px 38px -8px ${card.glowColor}`,
              animationDelay: isOpen ? `${index * 95}ms` : '0ms'
            }"
            :class="[
              'card-vertical-item relative w-full rounded-[28px] sm:rounded-[32px] p-5 sm:p-7 cursor-pointer overflow-hidden group',
              `bg-gradient-to-br ${card.gradient}`,
              isOpen ? 'animate-card-deal' : '',
              selectingCardId === card.id ? 'animate-card-select ring-4 ring-white z-20' : '',
              selectingCardId !== null && selectingCardId !== card.id ? 'opacity-50 scale-[0.97]' : '',
              activeMethod === card.id && selectingCardId === null
                ? 'ring-[3px] ring-white/90 shadow-2xl'
                : 'hover:-translate-y-1 hover:scale-[1.01] active:scale-[0.985]'
            ]"
          >
            <!-- Background Decorative Geometric Arcs (Matching Reference Image) -->
            <div class="absolute -right-14 -top-16 w-64 h-64 rounded-full border-[34px] border-black/[0.06] pointer-events-none transition-transform duration-500 group-hover:scale-105"></div>
            <div class="absolute -right-6 -top-8 w-44 h-44 rounded-full border-[22px] border-black/[0.04] pointer-events-none"></div>
            <div class="absolute left-1/3 -bottom-24 w-64 h-64 rounded-full bg-black/[0.04] pointer-events-none"></div>

            <!-- Specular Highlight Sheen -->
            <div class="absolute inset-0 bg-gradient-to-tr from-transparent via-white/[0.08] to-white/[0.22] pointer-events-none"></div>

            <!-- FROZEN STATE OVERLAY (Vector SVG Icon, No Emoji) -->
            <div
              v-if="frozenCards[card.id]"
              class="absolute inset-0 bg-slate-950/65 backdrop-blur-md z-30 flex items-center justify-center space-x-2.5 text-white font-extrabold text-xs sm:text-sm tracking-widest uppercase animate-in fade-in duration-200"
            >
              <svg class="w-5 h-5 text-sky-400 animate-spin" style="animation-duration: 8s;" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 3v18M3 12h18M6.343 6.343l11.314 11.314M6.343 17.657L17.657 6.343"/>
              </svg>
              <span>Kartu Dibekukan Sementara</span>
            </div>

            <!-- CARD TOP ROW: "Digital card" + Masked Number -->
            <div class="flex items-center justify-between relative z-10 mb-7 sm:mb-9">
              <div class="flex items-center space-x-2.5">
                <span class="font-bold text-xs sm:text-sm tracking-tight" :class="card.textColor">
                  Digital card
                </span>
                <span
                  v-if="selectingCardId === card.id"
                  class="text-[10px] font-mono font-extrabold uppercase px-2.5 py-0.5 rounded-full bg-white text-slate-950 shadow-sm"
                >
                  Memilih Kartu...
                </span>
                <span
                  v-else-if="activeMethod === card.id"
                  class="text-[10px] font-mono font-black uppercase px-2.5 py-0.5 rounded-full bg-black/20 text-white border border-white/20 backdrop-blur-sm"
                >
                  Filter Aktif
                </span>
              </div>

              <div class="font-mono font-bold text-xs sm:text-sm tracking-wider" :class="card.textColor">
                {{ card.cardNumber }}
              </div>
            </div>

            <!-- CARD MIDDLE ROW: Payment Channel & Total Spending -->
            <div class="relative z-10 mb-6 flex items-end justify-between">
              <div>
                <div class="text-[10px] font-bold uppercase tracking-widest font-mono opacity-75" :class="card.textColor">
                  {{ card.name }}
                </div>
                <div class="text-xl sm:text-2xl font-black font-mono tracking-tight mt-0.5" :class="card.textColor">
                  {{ formatIDR(methodSpendMap[card.id] || 0) }}
                  <span class="text-xs font-semibold opacity-75 ml-1">terpakai</span>
                </div>
              </div>

              <!-- Subtle tap prompt on hover -->
              <div
                class="hidden sm:flex items-center space-x-1 text-[11px] font-bold opacity-0 group-hover:opacity-90 transition-all duration-200 translate-x-2 group-hover:translate-x-0"
                :class="card.textColor"
              >
                <span>Pilih Filter</span>
                <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M9 5l7 7-7 7"/>
                </svg>
              </div>
            </div>

            <!-- CARD BOTTOM ROW: 3 Frosted Circular Action Buttons + Dual Overlapping Circles -->
            <div class="flex items-center justify-between relative z-10 pt-1">
              <!-- 3 Frosted Glass Action Buttons (Freeze, Card 123, Settings) -->
              <div class="flex items-center space-x-2.5">
                <!-- 1. Freeze Button (Snowflake Vector) -->
                <button
                  type="button"
                  @click="toggleFreeze(card, $event)"
                  :class="[
                    'tactile-btn w-10 h-10 sm:w-11 sm:h-11 rounded-full flex items-center justify-center backdrop-blur-xl border shadow-sm',
                    pressingActionKey === `freeze-${card.id}` ? 'scale-90 ring-2 ring-white' : '',
                    frozenCards[card.id]
                      ? 'bg-slate-950 text-sky-400 border-sky-400/50 shadow-lg'
                      : 'bg-black/20 hover:bg-black/30 text-white border-white/25'
                  ]"
                  :title="frozenCards[card.id] ? 'Aktifkan Kartu' : 'Bekukan Kartu'"
                >
                  <svg class="w-4 h-4 stroke-[2.2]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M12 3v18M3 12h18M6.343 6.343l11.314 11.314M6.343 17.657L17.657 6.343M12 7l2-2m-4 0l2 2m0 10l2 2m-4 0l2-2"/>
                  </svg>
                </button>

                <!-- 2. Card Details Button (Card 123 Vector) -->
                <button
                  type="button"
                  @click="openCardDetails(card, $event)"
                  :class="[
                    'tactile-btn w-10 h-10 sm:w-11 sm:h-11 rounded-full flex items-center justify-center backdrop-blur-xl bg-black/20 hover:bg-black/30 text-white border border-white/25 shadow-sm',
                    pressingActionKey === `details-${card.id}` ? 'scale-90 ring-2 ring-white' : ''
                  ]"
                  title="Lihat Nomor Kartu & CVV"
                >
                  <svg class="w-4 h-4 stroke-[2.1]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <rect x="2" y="5" width="20" height="14" rx="2.5"/>
                    <path stroke-linecap="round" stroke-linejoin="round" d="M2 10h20M6 14.5h4"/>
                  </svg>
                </button>

                <!-- 3. Settings Button (Gear Vector) -->
                <button
                  type="button"
                  @click="openCardSettings(card, $event)"
                  :class="[
                    'tactile-btn w-10 h-10 sm:w-11 sm:h-11 rounded-full flex items-center justify-center backdrop-blur-xl bg-black/20 hover:bg-black/30 text-white border border-white/25 shadow-sm',
                    pressingActionKey === `settings-${card.id}` ? 'scale-90 ring-2 ring-white' : ''
                  ]"
                  title="Pengaturan Limit & Keamanan Kartu"
                >
                  <svg class="w-4 h-4 stroke-[2.1]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"/>
                    <circle cx="12" cy="12" r="3"/>
                  </svg>
                </button>
              </div>

              <!-- Mastercard-style Dual Overlapping Circles -->
              <div class="flex items-center -space-x-3.5 pr-1">
                <div class="w-9 h-9 rounded-full bg-black/25 backdrop-blur-sm border border-white/10"></div>
                <div class="w-9 h-9 rounded-full bg-black/35 backdrop-blur-sm border border-white/10"></div>
              </div>
            </div>
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
          <div class="flex items-center space-x-2.5">
            <span class="w-2.5 h-2.5 rounded-full" :style="{ backgroundColor: selectedCardForDetails?.color }"></span>
            <span class="font-extrabold text-base tracking-tight text-white">{{ selectedCardForDetails?.name }}</span>
          </div>
          <button @click="showDetailsModal = false" class="text-slate-400 hover:text-white p-1 rounded-lg hover:bg-white/10">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.2" d="M6 18L18 6M6 6l12 12"/>
            </svg>
          </button>
        </div>

        <div class="py-5 space-y-4">
          <div>
            <div class="text-[10px] font-mono uppercase tracking-wider text-slate-400">Nomor Kartu Digital</div>
            <div class="flex items-center justify-between mt-1.5 bg-white/[0.05] p-3.5 rounded-xl border border-white/10">
              <span class="font-mono text-base font-bold tracking-widest text-emerald-400">
                4532 8901 {{ selectedCardForDetails?.cardNumber.replace('•••• ', '') }} 9482
              </span>
              <button
                type="button"
                @click="copyCardNumber"
                class="tactile-btn text-xs bg-white/10 hover:bg-white/20 text-white px-3 py-1.5 rounded-lg transition-colors font-semibold"
              >
                {{ isCopied ? 'Tersalin' : 'Salin' }}
              </button>
            </div>
          </div>

          <div class="grid grid-cols-2 gap-3">
            <div class="bg-white/[0.05] p-3.5 rounded-xl border border-white/10">
              <div class="text-[10px] font-mono uppercase tracking-wider text-slate-400">Masa Berlaku</div>
              <div class="font-mono text-sm font-bold text-white mt-0.5">08 / 29</div>
            </div>
            <div class="bg-white/[0.05] p-3.5 rounded-xl border border-white/10">
              <div class="text-[10px] font-mono uppercase tracking-wider text-slate-400">CVV / CVC</div>
              <div class="font-mono text-sm font-bold text-white mt-0.5">842</div>
            </div>
          </div>

          <div class="bg-white/[0.05] p-3.5 rounded-xl border border-white/10">
            <div class="text-[10px] font-mono uppercase tracking-wider text-slate-400">Pemegang Kartu</div>
            <div class="font-mono text-sm font-bold text-white mt-0.5">Randi Zakaria</div>
          </div>
        </div>

        <div class="pt-2">
          <v-btn
            block
            color="primary"
            class="rounded-xl font-bold py-2.5 text-slate-950"
            @click="showDetailsModal = false"
          >
            Tutup Rincian
          </v-btn>
        </div>
      </v-card>
    </v-dialog>

    <!-- 2. Card Settings Dialog (Vuetify VDialog + Controls) -->
    <v-dialog v-model="showSettingsModal" max-width="440">
      <v-card class="bg-[#0f131c] text-white rounded-3xl border border-white/10 p-6 shadow-2xl">
        <div class="flex items-center justify-between pb-4 border-b border-white/10">
          <div class="font-extrabold text-base tracking-tight text-white">
            Pengaturan {{ selectedCardForSettings?.shortName }}
          </div>
          <button @click="showSettingsModal = false" class="text-slate-400 hover:text-white p-1 rounded-lg hover:bg-white/10">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.2" d="M6 18L18 6M6 6l12 12"/>
            </svg>
          </button>
        </div>

        <div class="py-4 space-y-4">
          <div>
            <div class="flex justify-between text-xs font-mono mb-2">
              <span class="text-slate-400">Limit Transaksi Harian:</span>
              <span class="text-emerald-400 font-bold">{{ formatIDR(dailyLimit) }}</span>
            </div>
            <v-slider
              v-model="dailyLimit"
              :min="500000"
              :max="25000000"
              :step="500000"
              color="primary"
              track-color="rgba(255,255,255,0.12)"
              hide-details
            ></v-slider>
          </div>

          <div class="flex items-center justify-between py-2 border-t border-white/10">
            <div>
              <div class="text-xs font-bold text-white">Transaksi Online &amp; QRIS</div>
              <div class="text-[11px] text-slate-400">Izinkan pembayaran digital instan</div>
            </div>
            <v-switch v-model="onlineTransactions" color="primary" hide-details></v-switch>
          </div>

          <div class="flex items-center justify-between py-2 border-t border-white/10">
            <div>
              <div class="text-xs font-bold text-white">Contactless / Tap NFC</div>
              <div class="text-[11px] text-slate-400">Otorisasi mesin EDC tanpa PIN di bawah Rp 1 Juta</div>
            </div>
            <v-switch v-model="nfcEnabled" color="primary" hide-details></v-switch>
          </div>
        </div>

        <div class="pt-2">
          <v-btn
            block
            color="primary"
            class="rounded-xl font-bold py-2.5 text-slate-950"
            @click="showSettingsModal = false; notify('Pengaturan kartu berhasil diperbarui')"
          >
            Simpan Pengaturan
          </v-btn>
        </div>
      </v-card>
    </v-dialog>

    <!-- 3. Vuetify VSnackbar for tactile notifications -->
    <v-snackbar v-model="snackbar" :timeout="2800" location="top" color="#141923" rounded="pill">
      <div class="flex items-center space-x-2 text-xs font-bold text-white">
        <span class="w-2 h-2 rounded-full bg-emerald-400"></span>
        <span>{{ snackbarText }}</span>
      </div>
    </v-snackbar>
  </div>
</template>

<style scoped>
.cards-curtain-grid {
  display: grid;
  transition:
    grid-template-rows 0.68s cubic-bezier(0.22, 1, 0.36, 1),
    opacity 0.45s cubic-bezier(0.22, 1, 0.36, 1);
}

.cards-curtain-grid.is-open {
  grid-template-rows: 1fr;
  opacity: 1;
}

.cards-curtain-grid.is-closed {
  grid-template-rows: 0fr;
  opacity: 0;
  pointer-events: none;
}

.card-vertical-item {
  transition:
    transform 0.42s cubic-bezier(0.34, 1.56, 0.64, 1),
    box-shadow 0.35s cubic-bezier(0.22, 1, 0.36, 1),
    opacity 0.3s ease;
  will-change: transform, opacity;
}
</style>
