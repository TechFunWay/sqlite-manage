import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { databaseApi, tableApi, dataApi, downloadApi } from '../api'
import { useToastStore } from './toast'

export const useDatabaseStore = defineStore('database', () => {
  const toast = useToastStore()
  
  const isConnected = ref(false)
  const databases = ref([])
  const databaseInfo = ref(null)
  const tables = ref([])
  const currentTable = ref(null)
  const currentSchema = ref(null)
  const currentData = ref(null)
  const loading = ref(false)
  const currentPage = ref(1)
  const pageSize = ref(100)
  const totalRows = ref(0)
  const primaryKey = ref(null)
  const activeTab = ref('select') // 'select', 'data', 'schema', 'info'
  const queryWhere = ref('') // 查询条件 SQL
  const queryConditions = ref([]) // 查询条件结构化数据
  const queryLogic = ref('AND') // 查询条件逻辑 AND/OR

  const hasDatabase = computed(() => databases.value.length > 0)
  const hasQueryCondition = computed(() => queryConditions.value.length > 0)

  function showDatabaseSelector() {
    activeTab.value = 'select'
  }

  async function openDatabase(path) {
    loading.value = true
    try {
      const response = await databaseApi.open(path)
      const db = response.data
      await loadDatabases()
      await activateDatabase(db.id)
      toast.success('Database opened successfully')
      return true
    } catch (error) {
      toast.error(error.response?.data?.error || 'Failed to open database')
      return false
    } finally {
      loading.value = false
    }
  }

  async function createDatabase(name) {
    loading.value = true
    try {
      const response = await databaseApi.create(name)
      const db = response.data
      await loadDatabases()
      await activateDatabase(db.id)
      toast.success('数据库创建成功')
      return true
    } catch (error) {
      toast.error(error.response?.data?.error || '创建数据库失败')
      return false
    } finally {
      loading.value = false
    }
  }

  async function uploadDatabase(file) {
    loading.value = true
    try {
      const response = await databaseApi.upload(file)
      const db = response.data
      await loadDatabases()
      await activateDatabase(db.id)
      toast.success('Database uploaded successfully')
      return true
    } catch (error) {
      toast.error(error.response?.data?.error || 'Failed to upload database')
      return false
    } finally {
      loading.value = false
    }
  }

  async function loadDatabases() {
    try {
      const response = await databaseApi.getAll()
      databases.value = response.data
      
      // Find active database and load its tables
      const activeDb = databases.value.find(db => db.active)
      if (activeDb) {
        databaseInfo.value = activeDb
        isConnected.value = true
        tables.value = activeDb.tables || []
      } else if (databases.value.length > 0) {
        // If no active database, activate the first one
        await activateDatabase(databases.value[0].id)
      }
    } catch (error) {
      console.error('Failed to load databases', error)
    }
  }

  async function activateDatabase(id) {
    loading.value = true
    try {
      await databaseApi.activate(id)
      databases.value.forEach(db => {
        db.active = db.id === id
      })
      const db = databases.value.find(d => d.id === id)
      if (db) {
        databaseInfo.value = db
        isConnected.value = true
        // Load tables for this database
        await loadTables()
        // Save tables to database object
        db.tables = [...tables.value]
      }
    } catch (error) {
      toast.error('Failed to activate database')
    } finally {
      loading.value = false
    }
  }

  async function closeDatabase(id) {
    try {
      await databaseApi.close(id)
      databases.value = databases.value.filter(db => db.id !== id)
      
      if (databases.value.length === 0) {
        isConnected.value = false
        databaseInfo.value = null
        tables.value = []
        currentTable.value = null
        currentSchema.value = null
        currentData.value = null
      } else if (databaseInfo.value?.id === id) {
        const active = databases.value.find(db => db.active)
        if (active) {
          databaseInfo.value = active
          tables.value = active.tables || []
        } else {
          await activateDatabase(databases.value[0].id)
        }
      }
      toast.success('Database closed')
    } catch (error) {
      toast.error(error.response?.data?.error || 'Failed to close database')
    }
  }

  async function loadTables() {
    try {
      const response = await databaseApi.getTables()
      tables.value = response.data
      if (databaseInfo.value) {
        databaseInfo.value.tables = response.data
      }
    } catch (error) {
      toast.error('Failed to load tables')
    }
  }

  async function selectTable(tableName) {
    loading.value = true
    currentTable.value = tableName
    currentPage.value = 1
    // 切换表时清空查询条件
    queryWhere.value = ''
    queryConditions.value = []
    queryLogic.value = 'AND'
    activeTab.value = 'data' // Auto switch to data tab
    try {
      const [schemaRes, dataRes, pkRes] = await Promise.all([
        tableApi.getSchema(tableName),
        dataApi.getData(tableName, 1, pageSize.value),
        tableApi.getPrimaryKey(tableName)
      ])
      currentSchema.value = schemaRes.data
      currentData.value = dataRes.data.data
      totalRows.value = dataRes.data.total
      primaryKey.value = pkRes.data.primaryKey
    } catch (error) {
      toast.error('Failed to load table data')
    } finally {
      loading.value = false
    }
  }

  async function loadPage(page) {
    if (!currentTable.value) return
    loading.value = true
    try {
      const response = await dataApi.getData(currentTable.value, page, pageSize.value, queryWhere.value)
      currentData.value = response.data.data
      currentPage.value = page
      totalRows.value = response.data.total
    } catch (error) {
      toast.error('Failed to load page')
    } finally {
      loading.value = false
    }
  }

  async function executeQuery() {
    if (!currentTable.value) return
    loading.value = true
    currentPage.value = 1
    try {
      const response = await dataApi.getData(currentTable.value, 1, pageSize.value, queryWhere.value)
      currentData.value = response.data.data
      totalRows.value = response.data.total
      toast.success(`查询完成，共 ${response.data.total} 条记录`)
    } catch (error) {
      toast.error(error.response?.data?.error || '查询失败')
    } finally {
      loading.value = false
    }
  }

  function clearQuery() {
    queryWhere.value = ''
    queryConditions.value = []
    queryLogic.value = 'AND'
    if (currentTable.value) {
      loadPage(1)
    }
  }

  async function insertRow(data) {
    try {
      await dataApi.insert(currentTable.value, data)
      await loadPage(currentPage.value)
      toast.success('Row inserted successfully')
      return true
    } catch (error) {
      toast.error(error.response?.data?.error || 'Failed to insert row')
      return false
    }
  }

  async function updateRow(primaryKey, pkValue, data) {
    try {
      await dataApi.update(currentTable.value, primaryKey, pkValue, data)
      await loadPage(currentPage.value)
      toast.success('Row updated successfully')
      return true
    } catch (error) {
      toast.error(error.response?.data?.error || 'Failed to update row')
      return false
    }
  }

  async function deleteRow(primaryKey, pkValue) {
    try {
      await dataApi.delete(currentTable.value, primaryKey, pkValue)
      await loadPage(currentPage.value)
      toast.success('Row deleted successfully')
      return true
    } catch (error) {
      toast.error(error.response?.data?.error || 'Failed to delete row')
      return false
    }
  }

  async function createTable(name, columns) {
    try {
      const formattedColumns = columns.map(col => ({
        name: col.name,
        type: col.type,
        comment: col.comment || null,
        nullable: col.nullable,
        primaryKey: col.primaryKey,
        defaultValue: col.defaultValue
      }))
      await tableApi.create({ name, columns: formattedColumns })
      await loadTables()
      toast.success('Table created successfully')
      return true
    } catch (error) {
      toast.error(error.response?.data?.error || 'Failed to create table')
      return false
    }
  }

  async function dropTable(name) {
    try {
      await tableApi.drop(name)
      await loadTables()
      if (currentTable.value === name) {
        currentTable.value = null
        currentSchema.value = null
        currentData.value = null
      }
      toast.success('Table dropped successfully')
      return true
    } catch (error) {
      toast.error(error.response?.data?.error || 'Failed to drop table')
      return false
    }
  }

  async function addColumn(column) {
    try {
      await tableApi.addColumn(currentTable.value, column)
      await selectTable(currentTable.value)
      toast.success('Column added successfully')
      return true
    } catch (error) {
      toast.error(error.response?.data?.error || 'Failed to add column')
      return false
    }
  }

  async function dropColumn(columnName) {
    try {
      await tableApi.dropColumn(currentTable.value, columnName)
      await selectTable(currentTable.value)
      toast.success('Column dropped successfully')
      return true
    } catch (error) {
      toast.error(error.response?.data?.error || 'Failed to drop column')
      return false
    }
  }

  async function createIndex(name, columns, unique) {
    try {
      await tableApi.createIndex(currentTable.value, { name, columns, unique })
      await selectTable(currentTable.value)
      toast.success('Index created successfully')
      return true
    } catch (error) {
      toast.error(error.response?.data?.error || 'Failed to create index')
      return false
    }
  }

  async function dropIndex(name) {
    try {
      await tableApi.dropIndex(name)
      await selectTable(currentTable.value)
      toast.success('Index dropped successfully')
      return true
    } catch (error) {
      toast.error(error.response?.data?.error || 'Failed to drop index')
      return false
    }
  }

  async function importTableData(name, format, data) {
    try {
      let response
      if (format === 'json') {
        response = await dataApi.importJSON(name, data)
      } else if (format === 'csv') {
        response = await dataApi.importCSV(name, data)
      }
      await selectTable(name)
      toast.success(response.data?.message || '导入成功')
      return response.data
    } catch (error) {
      toast.error(error.response?.data?.error || '导入失败')
      return null
    }
  }

  async function exportTableData(name, format) {
    try {
      let response
      if (format === 'json') {
        response = await dataApi.exportJSON(name)
      } else if (format === 'csv') {
        response = await dataApi.exportCSV(name)
      }
      
      if (response) {
        const blob = new Blob([response.data])
        const url = URL.createObjectURL(blob)
        const a = document.createElement('a')
        a.href = url
        a.download = `${name}.${format}`
        a.click()
        URL.revokeObjectURL(url)
        toast.success('导出成功')
        return true
      }
      return false
    } catch (error) {
      toast.error(error.response?.data?.error || '导出失败')
      return false
    }
  }

  async function downloadDatabase() {
    try {
      const response = await downloadApi.downloadDatabase()
      const blob = new Blob([response.data])
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      const contentDisposition = response.headers['content-disposition']
      let filename = 'database.db'
      if (contentDisposition) {
        const match = contentDisposition.match(/filename="?([^"]+)"?/)
        if (match) filename = match[1]
      }
      a.download = filename
      a.click()
      URL.revokeObjectURL(url)
      toast.success('数据库下载成功')
      return true
    } catch (error) {
      toast.error(error.response?.data?.error || '数据库下载失败')
      return false
    }
  }

  return {
    isConnected,
    databases,
    databaseInfo,
    tables,
    currentTable,
    currentSchema,
    currentData,
    loading,
    currentPage,
    pageSize,
    totalRows,
    primaryKey,
    activeTab,
    queryWhere,
    queryConditions,
    queryLogic,
    hasQueryCondition,
    hasDatabase,
    showDatabaseSelector,
    openDatabase,
    createDatabase,
    uploadDatabase,
    loadDatabases,
    activateDatabase,
    closeDatabase,
    loadTables,
    selectTable,
    loadPage,
    executeQuery,
    clearQuery,
    insertRow,
    updateRow,
    deleteRow,
    createTable,
    dropTable,
    addColumn,
    dropColumn,
    createIndex,
    dropIndex,
    importTableData,
    exportTableData,
    downloadDatabase
  }
})
