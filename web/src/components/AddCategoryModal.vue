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
    class="fixed inset-0 z-50 overflow-y-auto bg-slate-900/50 backdrop-blur-sm flex items-center justify-center p-4"
    @click.self="emit('close')"
  >
    <div class="bg-white rounded-2xl shadow-xl border border-slate-200 max-w-md w-full overflow-hidden animate-in fade-in zoom-in-95 duration-150">
      <div class="px-6 py-4 border-b border-slate-100 flex items-center justify-between bg-slate-50/50">
        <h3 class="text-base font-bold text-slate-900">Tambah Kategori Baru</h3>
        <button @click="emit('close')" class="text-slate-400 hover:text-slate-600 p-1">✕</button>
      </div>

      <form @submit.prevent="handleSubmit" class="p-6 space-y-4">
        <div v-if="errorMessage" class="bg-rose-50 border border-rose-200 text-rose-700 text-sm px-4 py-3 rounded-xl">
          {{ errorMessage }}
        </div>

        <div>
          <label class="block text-xs font-bold uppercase tracking-wider text-slate-500 mb-1">
            Tipe Kategori
          </label>
          <div class="grid grid-cols-2 gap-2 bg-slate-100 p-1 rounded-xl">
            <button
              type="button"
              @click="catType = 'expense'"
              :class="[
                'py-2 text-sm font-semibold rounded-lg transition-all',
                catType === 'expense'
                  ? 'bg-rose-500 text-white shadow-sm'
                  : 'text-slate-600 hover:text-slate-900'
              ]"
            >
              💸 Pengeluaran
            </button>
            <button
              type="button"
              @click="catType = 'income'"
              :class="[
                'py-2 text-sm font-semibold rounded-lg transition-all',
                catType === 'income'
                  ? 'bg-emerald-600 text-white shadow-sm'
                  : 'text-slate-600 hover:text-slate-900'
              ]"
            >
              💰 Pemasukan
            </button>
          </div>
        </div>

        <div>
          <label class="block text-xs font-bold uppercase tracking-wider text-slate-500 mb-1">
            Nama Kategori
          </label>
          <input
            v-model="name"
            type="text"
            maxlength="100"
            placeholder="Contoh: Belanja Bulanan"
            class="w-full px-3.5 py-2.5 border border-slate-300 rounded-xl text-sm font-medium focus:ring-2 focus:ring-emerald-500"
          />
        </div>

        <div class="pt-4 border-t border-slate-100 flex items-center justify-end space-x-3">
          <button
            type="button"
            @click="emit('close')"
            class="px-4 py-2 text-sm font-semibold text-slate-600 hover:bg-slate-100 rounded-xl"
          >
            Batal
          </button>
          <button
            type="submit"
            :disabled="isSubmitting"
            class="px-5 py-2 text-sm font-semibold text-white bg-emerald-600 hover:bg-emerald-700 disabled:opacity-50 rounded-xl shadow-sm"
          >
            {{ isSubmitting ? 'Menyimpan...' : 'Simpan Kategori' }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>
