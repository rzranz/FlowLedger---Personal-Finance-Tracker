export type CategoryType = 'income' | 'expense'
export type TransactionType = 'income' | 'expense'
export type BudgetStatus = 'SAFE' | 'WARNING' | 'EXCEEDED'

export interface Category {
  id: number
  name: string
  type: CategoryType
  created_at: string
}

export interface Transaction {
  id: number
  category_id: number
  category?: Category
  type: TransactionType
  amount_idr: number
  occurred_on: string
  note: string
  created_at: string
  updated_at: string
}

export interface Budget {
  id: number
  category_id: number
  month: string
  limit_idr: number
  updated_at: string
}

export interface BudgetReportItem {
  budget: Budget
  category_name: string
  spent_idr: number
  remaining_idr: number
  utilization_pct: number
  status: BudgetStatus
}

export interface CategorySpendItem {
  category_id: number
  category_name: string
  amount_idr: number
  percentage: number
}

export interface MoMDelta {
  prev_income_idr: number
  prev_expense_idr: number
  income_delta_pct: number
  expense_delta_pct: number
}

export interface MonthlyReport {
  month: string
  total_income_idr: number
  total_expense_idr: number
  net_cash_flow_idr: number
  savings_rate_pct: number
  mom_comparison: MoMDelta
  category_breakdown: CategorySpendItem[]
  budget_watchlist: BudgetReportItem[]
}

export interface ProblemDetails {
  type: string
  title: string
  status: number
  detail: string
  instance: string
  validation_errors?: Array<{ field: string; reason: string }>
}

export function formatIDR(amount: number): string {
  const isNegative = amount < 0
  const abs = Math.abs(amount)
  const formatted = new Intl.NumberFormat('id-ID').format(abs)
  return isNegative ? `-Rp ${formatted}` : `Rp ${formatted}`
}

export function getCurrentMonth(): string {
  const now = new Date()
  const year = now.getFullYear()
  const month = String(now.getMonth() + 1).padStart(2, '0')
  return `${year}-${month}`
}

export type PaymentMethod = 'all' | 'qris' | 'cash' | 'atm' | 'transfer'

export interface CardMethodInfo {
  id: PaymentMethod
  name: string
  shortName: string
  icon: string
  color: string
  gradient: string
  border: string
  text: string
  badgeBg: string
}

export const PAYMENT_METHODS: CardMethodInfo[] = [
  {
    id: 'qris',
    name: 'QRIS / E-Wallet',
    shortName: 'QRIS',
    icon: '📱',
    color: '#10b981',
    gradient: 'from-emerald-400 via-teal-400 to-cyan-500',
    border: 'border-emerald-300',
    text: 'text-emerald-400',
    badgeBg: 'bg-emerald-500/20 text-emerald-300 border-emerald-500/30',
  },
  {
    id: 'cash',
    name: 'Uang Tunai',
    shortName: 'Cash',
    icon: '💵',
    color: '#f59e0b',
    gradient: 'from-amber-400 via-yellow-400 to-orange-500',
    border: 'border-amber-300',
    text: 'text-amber-400',
    badgeBg: 'bg-amber-500/20 text-amber-300 border-amber-500/30',
  },
  {
    id: 'atm',
    name: 'Kartu Debit / ATM',
    shortName: 'ATM',
    icon: '💳',
    color: '#06b6d4',
    gradient: 'from-cyan-400 via-sky-500 to-blue-600',
    border: 'border-cyan-300',
    text: 'text-cyan-400',
    badgeBg: 'bg-cyan-500/20 text-cyan-300 border-cyan-500/30',
  },
  {
    id: 'transfer',
    name: 'Transfer Bank',
    shortName: 'Transfer',
    icon: '🏦',
    color: '#8b5cf6',
    gradient: 'from-violet-400 via-purple-500 to-indigo-600',
    border: 'border-purple-300',
    text: 'text-purple-400',
    badgeBg: 'bg-purple-500/20 text-purple-300 border-purple-500/30',
  },
]

export function extractPaymentMethod(note: string, txId: number): PaymentMethod {
  const lower = (note || '').toLowerCase()
  if (lower.includes('[qris]') || lower.includes('qris')) return 'qris'
  if (lower.includes('[cash]') || lower.includes('[tunai]') || lower.includes('cash') || lower.includes('tunai')) return 'cash'
  if (lower.includes('[atm]') || lower.includes('[debit]') || lower.includes('atm') || lower.includes('debit') || lower.includes('kartu')) return 'atm'
  if (lower.includes('[transfer]') || lower.includes('transfer') || lower.includes('tf')) return 'transfer'
  
  // Deterministic fallback based on ID so sample data naturally maps to methods
  const methods: PaymentMethod[] = ['qris', 'cash', 'atm', 'transfer']
  return methods[txId % methods.length]
}

export function cleanNote(note: string): string {
  return (note || '')
    .replace(/\[(QRIS|CASH|TUNAI|ATM|DEBIT|TRANSFER)\]/gi, '')
    .trim()
}


