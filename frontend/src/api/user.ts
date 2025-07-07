import request from '@/utils/request'
import type { 
  User, 
  LoginRequest, 
  RegisterRequest, 
  UpdateUserRequest 
} from '@/types'

// 用戶註冊
export const register = (data: RegisterRequest) => {
  return request.post<User>('/users/register', data)
}

// 用戶登入
export const login = (data: LoginRequest) => {
  return request.post<{
    token: string
    user: User
    expires_at: string
  }>('/users/login', data)
}

// 獲取用戶資訊
export const getUserInfo = (id: number) => {
  return request.get<User>(`/users/${id}`)
}

// 更新用戶資訊
export const updateUser = (id: number, data: UpdateUserRequest) => {
  return request.put<User>(`/users/${id}`, data)
}

// 刪除用戶
export const deleteUser = (id: number) => {
  return request.delete(`/users/${id}`)
}
