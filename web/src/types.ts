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
  cardNumber: string
  icon: string
  color: string
  glowColor: string
  gradient: string
  cardBg: string
  border: string
  textColor: string
  badgeBg: string
  isFrozen?: boolean
}

export const PAYMENT_METHODS: CardMethodInfo[] = [
  {
    id: 'qris',
    name: 'QRIS Digital Card',
    shortName: 'QRIS',
    cardNumber: '•••• 7642',
    icon: '📱',
    color: '#FACC15',
    glowColor: 'rgba(250, 204, 21, 0.45)',
    gradient: 'from-[#FDE047] via-[#FACC15] to-[#EAB308]',
    cardBg: 'bg-[#FACC15]',
    border: 'border-yellow-200/40',
    textColor: 'text-slate-950',
    badgeBg: 'bg-yellow-500/20 text-yellow-300 border-yellow-500/30',
  },
  {
    id: 'cash',
    name: 'Cash / Tunai Wallet',
    shortName: 'Cash',
    cardNumber: '•••• 5123',
    icon: '💵',
    color: '#2DD4BF',
    glowColor: 'rgba(45, 212, 191, 0.45)',
    gradient: 'from-[#5EEAD4] via-[#2DD4BF] to-[#0D9488]',
    cardBg: 'bg-[#2DD4BF]',
    border: 'border-teal-200/40',
    textColor: 'text-slate-950',
    badgeBg: 'bg-teal-500/20 text-teal-300 border-teal-500/30',
  },
  {
    id: 'atm',
    name: 'Debit / ATM Card',
    shortName: 'ATM',
    cardNumber: '•••• 3413',
    icon: '💳',
    color: '#38BDF8',
    glowColor: 'rgba(56, 189, 248, 0.45)',
    gradient: 'from-[#38BDF8] via-[#0EA5E9] to-[#0284C7]',
    cardBg: 'bg-[#0EA5E9]',
    border: 'border-sky-200/40',
    textColor: 'text-white',
    badgeBg: 'bg-sky-500/20 text-sky-300 border-sky-500/30',
  },
  {
    id: 'transfer',
    name: 'Bank Transfer Vault',
    shortName: 'Transfer',
    cardNumber: '•••• 9021',
    icon: '🏦',
    color: '#A855F7',
    glowColor: 'rgba(168, 85, 247, 0.45)',
    gradient: 'from-[#C084FC] via-[#A855F7] to-[#7E22CE]',
    cardBg: 'bg-[#A855F7]',
    border: 'border-purple-200/40',
    textColor: 'text-white',
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


