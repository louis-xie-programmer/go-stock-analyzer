<template>
  <div class="strategy-list-view">
    <h2>动态策略列表</h2>
    <el-button type="primary" @click="openAddDialog">新增策略</el-button>
    <el-table :data="strategies" style="width: 100%; margin-top: 20px;">
      <el-table-column prop="id" label="ID" width="60"/>
      <el-table-column prop="name" label="名称"/>
      <el-table-column prop="description" label="描述"/>
      <el-table-column prop="author" label="作者"/>
      <el-table-column prop="created_at" label="创建时间" :formatter="formatDate"/>
      <el-table-column prop="updated_at" label="更新时间" :formatter="formatDate"/>
      <el-table-column label="操作" width="220">
        <template #default="scope">
          <el-button size="small" @click="openEditDialog(scope.row)">编辑</el-button>
          <el-button size="small" type="danger" @click="deleteStrategy(scope.row.id)">删除</el-button>
          <el-button size="small" type="success" @click="openTestDialog(scope.row)">测试</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 新增/编辑弹窗 -->
    <el-dialog :title="editMode ? '编辑策略' : '新增策略'" v-model="dialogVisible">
      <el-form :model="form">
        <el-form-item label="名称"><el-input v-model="form.name"/></el-form-item>
        <el-form-item label="描述"><el-input v-model="form.description"/></el-form-item>
        <el-form-item label="作者"><el-input v-model="form.author"/></el-form-item>
        <el-form-item label="代码"><el-input type="textarea" v-model="form.code" rows="6"/></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveStrategy">保存</el-button>
      </template>
    </el-dialog>

    <!-- 测试弹窗 -->
    <el-dialog title="策略测试" v-model="testDialogVisible">
      <el-form :model="testForm">
        <el-form-item label="目标">
          <el-select v-model="testForm.target" placeholder="请选择">
            <el-option label="自选股" value="watchlist"/>
            <el-option label="全部" value="all"/>
            <el-option label="上证主板" value="board:上证主板"/>
            <el-option label="深证主板" value="board:深证主板"/>
            <el-option label="创业板" value="board:创业板"/>
            <el-option label="科创板" value="board:科创板"/>
          </el-select>
        </el-form-item>
        <el-form-item label="回测天数"><el-input v-model.number="testForm.days" type="number"/></el-form-item>
      </el-form>
      <div v-if="testResult">
        <h4>测试结果：</h4>
        <pre>{{ testResult }}</pre>
      </div>
      <template #footer>
        <el-button @click="testDialogVisible = false">关闭</el-button>
        <el-button type="primary" @click="runTest">运行</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import axios from 'axios'
import { ElMessageBox, ElMessage } from 'element-plus'

const strategies = ref([])
const dialogVisible = ref(false)
const editMode = ref(false)
const form = ref({ id: null, name: '', description: '', author: '', code: '' })

const testDialogVisible = ref(false)
const testForm = ref({ id: null, target: 'watchlist', days: 120 })
const testResult = ref('')

function formatDate(row, column, cellValue) {
  if (!cellValue) return ''
  return typeof cellValue === 'string' ? cellValue.replace('T', ' ').slice(0, 19) : cellValue
}

function fetchStrategies() {
  axios.get('/api/strategy').then(res => {
    strategies.value = res.data.list || []
  })
}

function openAddDialog() {
  editMode.value = false
  form.value = { id: null, name: '', description: '', author: '', code: '' }
  dialogVisible.value = true
}

function openEditDialog(row) {
  editMode.value = true
  form.value = { ...row }
  dialogVisible.value = true
}

function saveStrategy() {
  if (!form.value.name || !form.value.code) {
    ElMessage.error('名称和代码不能为空')
    return
  }
  if (editMode.value) {
    axios.put(`/api/strategy/${form.value.id}`, form.value).then(() => {
      ElMessage.success('修改成功')
      dialogVisible.value = false
      fetchStrategies()
    })
  } else {
    axios.post('/api/strategy', form.value).then(() => {
      ElMessage.success('新增成功')
      dialogVisible.value = false
      fetchStrategies()
    })
  }
}

function deleteStrategy(id) {
  ElMessageBox.confirm('确定要删除该策略吗？', '提示', { type: 'warning' })
    .then(() => {
      axios.delete(`/api/strategy/${id}`).then(() => {
        ElMessage.success('删除成功')
        fetchStrategies()
      })
    })
}

function openTestDialog(row) {
  testForm.value = { id: row.id, target: 'watchlist', days: 120 }
  testResult.value = ''
  testDialogVisible.value = true
}

function runTest() {
  const body = {
    id: testForm.value.id,
    target: testForm.value.target,
    days: testForm.value.days
  }
  axios.post('/api/strategy/run', body).then(res => {
    testResult.value = JSON.stringify(res.data, null, 2)
  })
}

onMounted(fetchStrategies)
</script>

<style scoped>
.strategy-list-view {
  padding: 24px;
}
</style>