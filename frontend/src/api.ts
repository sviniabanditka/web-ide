import axios from 'axios'

export const api = axios.create({
  withCredentials: true
})

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('session_token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})
