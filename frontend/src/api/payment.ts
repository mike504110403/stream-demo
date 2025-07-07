import request from '@/utils/request'
import type { 
  Payment, 
  CreatePaymentRequest, 
  ProcessPaymentRequest,
  RefundPaymentRequest 
} from '@/types'

// 獲取支付列表
export const getPayments = (params?: { offset?: number; limit?: number }) => {
  return request.get<Payment[]>('/payments', { params })
}

// 創建支付
export const createPayment = (data: CreatePaymentRequest) => {
  return request.post<Payment>('/payments', data)
}

// 獲取單個支付
export const getPayment = (id: number) => {
  return request.get<Payment>(`/payments/${id}`)
}

// 獲取用戶的支付記錄
export const getUserPayments = (userId: number) => {
  return request.get<Payment[]>(`/users/${userId}/payments`)
}

// 處理支付
export const processPayment = (id: number, data: ProcessPaymentRequest) => {
  return request.post<Payment>(`/payments/${id}/process`, data)
}

// 退款
export const refundPayment = (id: number, data: RefundPaymentRequest) => {
  return request.post<Payment>(`/payments/${id}/refund`, data)
} 