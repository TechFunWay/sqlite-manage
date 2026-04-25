import axios from 'axios'

const api = axios.create({
  baseURL: '/api',
  timeout: 30000
})

// Request interceptor to add auth token
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => Promise.reject(error)
)

// Response interceptor to handle auth errors
api.interceptors.response.use(
  (response) => response,
  (error) => {
    // Don't redirect on login/register/check endpoints
    const skipRedirectUrls = ['/auth/login', '/auth/register', '/auth/check']
    const requestUrl = error.config?.url || ''
    
    if (error.response?.status === 401 && !skipRedirectUrls.some(url => requestUrl.includes(url))) {
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

export const systemApi = {
  getVersion: () => api.get('/version'),
  getUpgradeStatus: () => api.get('/upgrade-status')
}

export const authApi = {
  check: () => api.get('/auth/check'),
  login: (data) => api.post('/auth/login', data),
  register: (data) => api.post('/auth/register', data),
  changePassword: (data) => api.post('/auth/change-password', data),
  logout: () => api.post('/auth/logout')
}

export const fileApi = {
  browse: (path) => api.get('/files/browse', { params: { path } }),
  getShares: () => api.get('/files/shares')
}

export const recentApi = {
  get: () => api.get('/recent-databases'),
  add: (data) => api.post('/recent-databases', data),
  clear: () => api.delete('/recent-databases')
}

export const databaseApi = {
  open: (path) => api.post('/database/open', { path }),
  create: (name) => api.post('/database/create', { name }),
  upload: (file) => {
    const formData = new FormData()
    formData.append('file', file)
    return api.post('/database/upload', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
  },
  getInfo: () => api.get('/database/info'),
  getAll: () => api.get('/databases'),
  activate: (id) => api.put(`/databases/${id}/activate`),
  close: (id) => api.delete(`/databases/${id}`),
  getTables: () => api.get('/tables')
}

export const tableApi = {
  getSchema: (name) => api.get(`/tables/${name}/schema`),
  create: (data) => api.post('/tables', data),
  drop: (name) => api.delete(`/tables/${name}`),
  rename: (oldName, newName) => api.put('/tables/rename', { oldName, newName }),
  addColumn: (tableName, column) => api.post(`/tables/${tableName}/columns`, column),
  dropColumn: (tableName, columnName) => api.delete(`/tables/${tableName}/columns/${columnName}`),
  getIndexes: (name) => api.get(`/tables/${name}/indexes`),
  createIndex: (tableName, data) => api.post(`/tables/${tableName}/indexes`, data),
  dropIndex: (name) => api.delete(`/tables/${name}/indexes/${name}`),
  getPrimaryKey: (name) => api.get(`/tables/${name}/primarykey`)
}

export const dataApi = {
  getData: (name, page = 1, pageSize = 100, where = '') => 
    api.get(`/tables/${name}/data`, { params: { page, pageSize, where } }),
  insert: (name, data) => api.post(`/tables/${name}/data`, { data }),
  update: (name, primaryKey, pkValue, data) => 
    api.put(`/tables/${name}/data`, { primaryKey, pkValue, data }),
  delete: (name, primaryKey, pkValue) => 
    api.delete(`/tables/${name}/data`, { data: { primaryKey, pkValue } }),
  importJSON: (name, data) => api.post(`/tables/${name}/import?format=json`, data, {
    headers: { 'Content-Type': 'application/json' }
  }),
  importCSV: (name, file) => {
    const formData = new FormData()
    formData.append('file', file)
    return api.post(`/tables/${name}/import?format=csv`, formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
  },
  exportJSON: (name) => {
    return api.get(`/tables/${name}/export?format=json`, { responseType: 'blob' })
  },
  exportCSV: (name) => {
    return api.get(`/tables/${name}/export?format=csv`, { responseType: 'blob' })
  }
}

export const downloadApi = {
  downloadDatabase: () => api.get('/database/download', { responseType: 'blob' })
}

export const queryApi = {
  execute: (sql) => api.post('/query', { sql })
}

export default api
