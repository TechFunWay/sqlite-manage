import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authApi } from '../api'
import { useToastStore } from './toast'
import CryptoJS from 'crypto-js'

export const useAuthStore = defineStore('auth', () => {
  const toast = useToastStore()
  
  const token = ref(localStorage.getItem('token') || '')
  const user = ref(JSON.parse(localStorage.getItem('user') || 'null'))
  const hasAdmin = ref(false)
  const loading = ref(false)

  const isLoggedIn = computed(() => !!token.value && !!user.value)
  const isAdmin = computed(() => user.value?.username === 'admin')

  // MD5 hash function
  function md5Hash(text) {
    return CryptoJS.MD5(text).toString()
  }

  async function checkAuth() {
    try {
      const response = await authApi.check()
      hasAdmin.value = response.data.hasAdmin
      return response.data
    } catch (error) {
      return { hasAdmin: false }
    }
  }

  async function login(username, password) {
    loading.value = true
    try {
      // Hash password with MD5 before sending
      const hashedPassword = md5Hash(password)
      const response = await authApi.login({ username, password: hashedPassword })
      
      token.value = response.data.token
      user.value = response.data.user
      
      localStorage.setItem('token', response.data.token)
      localStorage.setItem('user', JSON.stringify(response.data.user))
      
      toast.success('登录成功')
      return true
    } catch (error) {
      toast.error(error.response?.data?.error || '登录失败')
      return false
    } finally {
      loading.value = false
    }
  }

  async function register(username, password) {
    loading.value = true
    try {
      // Hash password with MD5 before sending
      const hashedPassword = md5Hash(password)
      const response = await authApi.register({ username, password: hashedPassword })
      
      token.value = response.data.token
      user.value = response.data.user
      
      localStorage.setItem('token', response.data.token)
      localStorage.setItem('user', JSON.stringify(response.data.user))
      
      toast.success('注册成功')
      return true
    } catch (error) {
      toast.error(error.response?.data?.error || '注册失败')
      return false
    } finally {
      loading.value = false
    }
  }

  async function changePassword(oldPassword, newPassword) {
    loading.value = true
    try {
      // Hash passwords with MD5 before sending
      const hashedOldPassword = md5Hash(oldPassword)
      const hashedNewPassword = md5Hash(newPassword)
      
      await authApi.changePassword({ 
        oldPassword: hashedOldPassword, 
        newPassword: hashedNewPassword 
      })
      
      toast.success('密码修改成功')
      return true
    } catch (error) {
      toast.error(error.response?.data?.error || '密码修改失败')
      return false
    } finally {
      loading.value = false
    }
  }

  async function logout() {
    try {
      await authApi.logout()
    } catch (error) {
      // Ignore logout error
    }
    
    token.value = ''
    user.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('user')
    
    toast.success('已退出登录')
  }

  return {
    token,
    user,
    hasAdmin,
    loading,
    isLoggedIn,
    isAdmin,
    checkAuth,
    login,
    register,
    changePassword,
    logout
  }
})
