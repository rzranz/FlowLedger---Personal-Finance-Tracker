<script setup lang="ts">
import { ref } from 'vue'
import { api } from '../api/client'
import type { CategoryType, ProblemDetails } from '../types'

defineProps<{
  isOpen: boolean
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'created'): void
}>()

const name = ref('')
const catType = ref<CategoryType>('expense')
const isSubmitting = ref(false)
const errorMessage = ref<string | null>(null)

async function handleSubmit() {
  if (!name.value.trim()) {
    errorMessage.value = 'Nama kategori tidak boleh kosong'
    return
  }

  isSubmitting.value = true
  errorMessage.value = null

  try {
    await api.createCategory(name.value.trim(), catType.value)
    name.value = ''
    emit('created')
    emit('close')
  } catch (err: any) {
    const prob = err as ProblemDetails
    errorMessage.value = prob.detail || 'Gagal membuat kategori'
  } finally {
    isSubmitting.value = false
  }
}
</script>

<template>
  <div
    v-if="isOpen"
    class="fixed inset-0 z-50 overflow-y-auto bg-slate-950/70 backdrop-blur-md flex items-center justify-center p-4"
    @click.self="emit('close')"
  >
    <div class="bg-[#0f131c] text-white rounded-3xl shadow-2xl border border-white/10 max-w-md w-full overflow-hidden animate-in fade-in zoom-in-95 duration-200">
      <div class="px-6 py-4 border-b border-white/10 flex items-center justify-between">
        <div class="flex items-center space-x-2.5">
          <span class="w-2.5 h-2.5 rounded-full bg-emerald-400"></span>
          <h3 class="text-base font-extrabold text-white tracking-tight">Tambah Kategori Baru</h3>
        </div>
        <button @click="emit('close')" class="text-slate-400 hover:text-white p-1.5 rounded-lg hover:bg-white/10">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.2" d="M6 18L18 6M6 6l12 12"/>
          </svg>
        </button>
      </div>

      <form @submit.prevent="handleSubmit" class="p-6 space-y-5">
        <div v-if="errorMessage" class="bg-rose-500/15 border border-rose-500/30 text-rose-200 text-xs px-4 py-3 rounded-xl">
          {{ errorMessage }}
        </div>

        <div>
          <label class="block text-[11px] font-mono font-bold uppercase tracking-widest text-slate-400 mb-2">
            Tipe Kategori
          </label>
          <div class="grid grid-cols-2 gap-2 bg-white/[0.05] p-1.5 rounded-2xl border border-white/10">
            <button
              type="button"
              @click="catType = 'expense'"
              :class="[
                'tactile-btn py-2.5 px-3 text-xs font-bold rounded-xl flex items-center justify-center space-x-1.5 transition-all',
                catType === 'expense'
                  ? 'bg-rose-500 text-white shadow-md'
                  : 'text-slate-400 hover:text-white'
              ]"
            >
              <svg class="w-3.5 h-3.5 stroke-[2.4]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" d="M7 17L17 7M17 7H8M17 7V16"/>
              </svg>
              <span>Pengeluaran</span>
            </button>
            <button
              type="button"
              @click="catType = 'income'"
              :class="[
                'tactile-btn py-2.5 px-3 text-xs font-bold rounded-xl flex items-center justify-center space-x-1.5 transition-all',
                catType === 'income'
                  ? 'bg-emerald-500 text-slate-950 shadow-md'
                  : 'text-slate-400 hover:text-white'
              ]"
            >
              <svg class="w-3.5 h-3.5 stroke-[2.4]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" d="M17 7L7 17M7 17H16M7 17V8"/>
              </svg>
              <span>Pemasukan</span>
            </button>
          </div>
        </div>

        <div>
          <label class="block text-[11px] font-mono font-bold uppercase tracking-widest text-slate-400 mb-1.5">
            Nama Kategori
          </label>
          <input
            v-model="name"
            type="text"
            maxlength="100"
            placeholder="Contoh: Belanja Bulanan"
            class="w-full px-4 py-3 bg-[#161c2a] border border-white/15 rounded-xl text-sm font-semibold text-white placeholder:text-slate-500 focus:outline-none focus:border-emerald-400"
          />
        </div>

        <div class="pt-3 border-t border-white/10 flex items-center justify-end space-x-3">
          <button
            type="button"
            @click="emit('close')"
            class="tactile-btn px-4 py-2.5 text-xs font-bold text-slate-300 hover:text-white bg-white/[0.06] hover:bg-white/10 rounded-xl"
          >
            Batal
          </button>
          <button
            type="submit"
            :disabled="isSubmitting"
            class="tactile-btn px-5 py-2.5 text-xs font-black text-slate-950 bg-emerald-400 hover:bg-emerald-300 disabled:opacity-50 rounded-xl shadow-md"
          >
            {{ isSubmitting ? 'Menyimpan...' : 'Simpan Kategori' }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>
