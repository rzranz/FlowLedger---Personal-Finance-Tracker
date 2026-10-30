<script setup lang="ts">
import { ref } from 'vue'
import { api } from '../api/client'
import type { Category, ProblemDetails } from '../types'

defineProps<{
  categories: Category[]
}>()

const emit = defineEmits<{
  (e: 'refresh'): void
  (e: 'open-add-modal'): void
}>()

const editingId = ref<number | null>(null)
const editName = ref<string>('')
const actionError = ref<string | null>(null)

function startEdit(cat: Category) {
  editingId.value = cat.id
  editName.value = cat.name
  actionError.value = null
}

function cancelEdit() {
  editingId.value = null
  editName.value = ''
}

async function saveEdit(id: number) {
  if (!editName.value.trim()) return
  try {
    await api.updateCategory(id, editName.value.trim())
    editingId.value = null
    emit('refresh')
  } catch (err: any) {
    const prob = err as ProblemDetails
    actionError.value = prob.detail || 'Gagal mengubah nama kategori'
  }
}

async function handleDelete(cat: Category) {
  if (!confirm(`Hapus kategori "${cat.name}"?`)) return
  actionError.value = null
  try {
    await api.deleteCategory(cat.id)
    emit('refresh')
  } catch (err: any) {
    const prob = err as ProblemDetails
    actionError.value = prob.detail || 'Kategori tidak dapat dihapus karena masih digunakan oleh transaksi atau anggaran.'
  }
}
</script>

<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 bg-white p-4 sm:p-6 rounded-2xl border border-slate-200 shadow-sm">
      <div>
        <h1 class="text-xl sm:text-2xl font-extrabold text-slate-900 tracking-tight">Kategori Finansial</h1>
        <p class="text-xs sm:text-sm text-slate-500">Struktur klasifikasi pemasukan dan pengeluaran dengan referential integrity</p>
      </div>

      <button
        @click="emit('open-add-modal')"
        class="px-4 py-2 bg-emerald-600 hover:bg-emerald-700 text-white rounded-xl text-sm font-semibold shadow-sm transition-all"
      >
        + Tambah Kategori
      </button>
    </div>

    <!-- Action Error Notice -->
    <div v-if="actionError" class="bg-rose-50 border border-rose-200 text-rose-700 p-4 rounded-xl text-sm flex items-center justify-between">
      <span>{{ actionError }}</span>
      <button @click="actionError = null" class="text-rose-500 hover:text-rose-700 font-bold ml-4">✕</button>
    </div>

    <!-- Categories Grid (Expense & Income) -->
    <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
      <!-- Pengeluaran (Expense) -->
      <div class="bg-white rounded-2xl border border-slate-200 shadow-sm p-6">
        <div class="flex items-center space-x-2 pb-4 border-b border-slate-100 mb-4">
          <span class="w-3 h-3 rounded-full bg-rose-500"></span>
          <h2 class="text-base font-bold text-slate-900">Kategori Pengeluaran (Expense)</h2>
        </div>

        <ul class="divide-y divide-slate-100">
          <li
            v-for="cat in categories.filter(c => c.type === 'expense')"
            :key="cat.id"
            class="py-3 flex items-center justify-between"
          >
            <div v-if="editingId === cat.id" class="flex items-center space-x-2 flex-1 mr-3">
              <input
                v-model="editName"
                type="text"
                class="px-3 py-1.5 border border-slate-300 rounded-lg text-sm flex-1 focus:ring-2 focus:ring-emerald-500"
              />
              <button
                @click="saveEdit(cat.id)"
                class="text-xs bg-emerald-600 text-white px-2.5 py-1.5 rounded-lg font-semibold"
              >
                Simpan
              </button>
              <button
                @click="cancelEdit"
                class="text-xs bg-slate-100 text-slate-600 px-2 py-1.5 rounded-lg"
              >
                Batal
              </button>
            </div>
            <div v-else class="flex items-center space-x-2">
              <span class="text-sm font-medium text-slate-800">{{ cat.name }}</span>
            </div>

            <div v-if="editingId !== cat.id" class="flex items-center space-x-2 text-xs">
              <button
                @click="startEdit(cat)"
                class="text-slate-500 hover:text-slate-800 font-medium px-2 py-1 rounded hover:bg-slate-50"
              >
                Edit
              </button>
              <button
                @click="handleDelete(cat)"
                class="text-rose-500 hover:text-rose-700 font-medium px-2 py-1 rounded hover:bg-rose-50"
              >
                Hapus
              </button>
            </div>
          </li>
        </ul>
      </div>

      <!-- Pemasukan (Income) -->
      <div class="bg-white rounded-2xl border border-slate-200 shadow-sm p-6">
        <div class="flex items-center space-x-2 pb-4 border-b border-slate-100 mb-4">
          <span class="w-3 h-3 rounded-full bg-emerald-500"></span>
          <h2 class="text-base font-bold text-slate-900">Kategori Pemasukan (Income)</h2>
        </div>

        <ul class="divide-y divide-slate-100">
          <li
            v-for="cat in categories.filter(c => c.type === 'income')"
            :key="cat.id"
            class="py-3 flex items-center justify-between"
          >
            <div v-if="editingId === cat.id" class="flex items-center space-x-2 flex-1 mr-3">
              <input
                v-model="editName"
                type="text"
                class="px-3 py-1.5 border border-slate-300 rounded-lg text-sm flex-1 focus:ring-2 focus:ring-emerald-500"
              />
              <button
                @click="saveEdit(cat.id)"
                class="text-xs bg-emerald-600 text-white px-2.5 py-1.5 rounded-lg font-semibold"
              >
                Simpan
              </button>
              <button
                @click="cancelEdit"
                class="text-xs bg-slate-100 text-slate-600 px-2 py-1.5 rounded-lg"
              >
                Batal
              </button>
            </div>
            <div v-else class="flex items-center space-x-2">
              <span class="text-sm font-medium text-slate-800">{{ cat.name }}</span>
            </div>

            <div v-if="editingId !== cat.id" class="flex items-center space-x-2 text-xs">
              <button
                @click="startEdit(cat)"
                class="text-slate-500 hover:text-slate-800 font-medium px-2 py-1 rounded hover:bg-slate-50"
              >
                Edit
              </button>
              <button
                @click="handleDelete(cat)"
                class="text-rose-500 hover:text-rose-700 font-medium px-2 py-1 rounded hover:bg-rose-50"
              >
                Hapus
              </button>
            </div>
          </li>
        </ul>
      </div>
    </div>
  </div>
</template>
