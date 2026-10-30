import type {
  Category,
  CategoryType,
  Transaction,
  BudgetReportItem,
  MonthlyReport,
  ProblemDetails,
} from '../types'

const BASE_URL = ''

async function handleResponse<T>(res: Response): Promise<T> {
  if (!res.ok) {
    let errPayload: ProblemDetails
    try {
      errPayload = await res.json()
    } catch {
      errPayload = {
        type: 'https://flowledger.dev/errors/http-error',
        title: res.statusText || 'Error',
        status: res.status,
        detail: `Request failed with HTTP status ${res.status}`,
        instance: res.url,
      }
    }
    throw errPayload
  }
  if (res.status === 204) {
    return {} as T
  }
  return res.json()
}

export const api = {
  // Probes
  async getHealth() {
    return handleResponse<{ status: string }>(await fetch(`${BASE_URL}/health`))
  },

  async getReady() {
    return handleResponse<{ status: string }>(await fetch(`${BASE_URL}/ready`))
  },

  // Categories
  async getCategories(type?: CategoryType): Promise<{ items: Category[] }> {
    const url = type ? `${BASE_URL}/api/v1/categories?type=${type}` : `${BASE_URL}/api/v1/categories`
    return handleResponse(await fetch(url))
  },

  async createCategory(name: string, type: CategoryType): Promise<Category> {
    return handleResponse(
      await fetch(`${BASE_URL}/api/v1/categories`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name, type }),
      })
    )
  },

  async updateCategory(id: number, name: string): Promise<Category> {
    return handleResponse(
      await fetch(`${BASE_URL}/api/v1/categories/${id}`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name }),
      })
    )
  },

  async deleteCategory(id: number): Promise<void> {
    return handleResponse(
      await fetch(`${BASE_URL}/api/v1/categories/${id}`, {
        method: 'DELETE',
      })
    )
  },

  // Transactions
  async getTransactions(params: {
    month?: string
    type?: string
    category_id?: number
    q?: string
    page?: number
    limit?: number
  }): Promise<{ items: Transaction[]; pagination: { page: number; limit: number; total: number; total_pages: number } }> {
    const sp = new URLSearchParams()
    if (params.month) sp.set('month', params.month)
    if (params.type) sp.set('type', params.type)
    if (params.category_id) sp.set('category_id', String(params.category_id))
    if (params.q) sp.set('q', params.q)
    if (params.page) sp.set('page', String(params.page))
    if (params.limit) sp.set('limit', String(params.limit))

    return handleResponse(await fetch(`${BASE_URL}/api/v1/transactions?${sp.toString()}`))
  },

  async createTransaction(payload: {
    category_id: number
    type: 'income' | 'expense'
    amount_idr: number
    occurred_on: string
    note: string
  }): Promise<Transaction> {
    const idempotencyKey = 'tx-' + crypto.randomUUID()
    return handleResponse(
      await fetch(`${BASE_URL}/api/v1/transactions`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Idempotency-Key': idempotencyKey,
        },
        body: JSON.stringify(payload),
      })
    )
  },

  async updateTransaction(
    id: number,
    payload: {
      category_id?: number
      type?: 'income' | 'expense'
      amount_idr?: number
      occurred_on?: string
      note?: string
    }
  ): Promise<Transaction> {
    return handleResponse(
      await fetch(`${BASE_URL}/api/v1/transactions/${id}`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      })
    )
  },

  async deleteTransaction(id: number): Promise<void> {
    return handleResponse(
      await fetch(`${BASE_URL}/api/v1/transactions/${id}`, {
        method: 'DELETE',
      })
    )
  },

  // Budgets
  async getBudgets(month: string): Promise<{ items: BudgetReportItem[] }> {
    return handleResponse(await fetch(`${BASE_URL}/api/v1/budgets?month=${month}`))
  },

  async upsertBudget(categoryId: number, month: string, limitIdr: number): Promise<any> {
    return handleResponse(
      await fetch(`${BASE_URL}/api/v1/budgets/${categoryId}/${month}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ limit_idr: limitIdr }),
      })
    )
  },

  // Monthly Analytics Report
  async getMonthlyReport(month: string): Promise<MonthlyReport> {
    return handleResponse(await fetch(`${BASE_URL}/api/v1/reports/monthly?month=${month}`))
  },

  // CSV Export URL helper
  getExportCsvUrl(params: { month?: string; type?: string; category_id?: number; q?: string }): string {
    const sp = new URLSearchParams()
    if (params.month) sp.set('month', params.month)
    if (params.type) sp.set('type', params.type)
    if (params.category_id) sp.set('category_id', String(params.category_id))
    if (params.q) sp.set('q', params.q)
    return `${BASE_URL}/api/v1/transactions/export.csv?${sp.toString()}`
  },
}

