<script setup lang="ts">
import type { Todo } from "~/types/api"

const api = useAdminApi()
const todos = ref<Todo[]>([])
const title = ref("")
const description = ref("")
const loading = ref(false)
const saving = ref(false)
const error = ref("")
const success = ref("")

async function load() {
  loading.value = true
  error.value = ""
  try {
    todos.value = await api.listTodos()
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    loading.value = false
  }
}

async function createTodo() {
  if (!title.value.trim()) {
    return
  }

  saving.value = true
  error.value = ""
  success.value = ""
  try {
    await api.createTodo({
      description: description.value.trim(),
      title: title.value.trim()
    })
    title.value = ""
    description.value = ""
    success.value = "Todo 已创建。"
    await load()
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    saving.value = false
  }
}

async function toggleTodo(todo: Todo) {
  try {
    await api.updateTodo(todo.id, { completed: !todo.completed })
    await load()
  } catch (err) {
    error.value = errorMessage(err)
  }
}

async function removeTodo(id: number) {
  try {
    await api.deleteTodo(id)
    success.value = `Todo #${id} 已删除。`
    await load()
  } catch (err) {
    error.value = errorMessage(err)
  }
}

onMounted(load)

useHead({
  title: "Demo Todo - Aoi Admin"
})
</script>

<template>
  <div class="page-grid">
    <PageHeader title="Demo Todo" icon="list-checks" description="使用公开 Demo Todo API 验证静态管理台与 Go Result 契约。">
      <template #actions>
        <AoiButton appearance="soft" icon="refresh-cw" :loading="loading" @click="load">刷新</AoiButton>
      </template>
    </PageHeader>

    <AoiStatusMessage tone="danger" :message="error" />
    <AoiStatusMessage tone="success" :message="success" />

    <section class="two-column-grid">
      <article class="admin-card">
        <div class="admin-card__header">
          <h2>Todo 列表</h2>
          <span class="badge">{{ todos.length }} 条</span>
        </div>
        <div class="data-table-wrap">
          <table class="data-table">
            <thead>
              <tr>
                <th>ID</th>
                <th>标题</th>
                <th>描述</th>
                <th>状态</th>
                <th>更新时间</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="todo in todos" :key="todo.id">
                <td>{{ todo.id }}</td>
                <td>{{ todo.title }}</td>
                <td>{{ todo.description || "-" }}</td>
                <td><span class="badge" :class="todo.completed ? 'badge--success' : 'badge--warning'">{{ todo.completed ? "完成" : "待办" }}</span></td>
                <td>{{ formatDateTime(todo.updatedAt) }}</td>
                <td>
                  <div class="action-row">
                    <AoiButton appearance="soft" icon="check" @click="toggleTodo(todo)">切换</AoiButton>
                    <AoiButton appearance="soft" intent="danger" icon="trash-2" @click="removeTodo(todo.id)">删除</AoiButton>
                  </div>
                </td>
              </tr>
              <tr v-if="!todos.length">
                <td colspan="6" class="muted">暂无 Todo。</td>
              </tr>
            </tbody>
          </table>
        </div>
      </article>

      <article class="admin-card">
        <div class="admin-card__header">
          <h2>新建 Todo</h2>
        </div>
        <form class="admin-card__body form-grid" @submit.prevent="createTodo">
          <AoiTextField v-model="title" label="标题" icon="type" />
          <AoiTextField v-model="description" label="描述" type="textarea" icon="file-text" />
          <AoiButton type="submit" icon="plus" :loading="saving" :disabled="!title">创建 Todo</AoiButton>
        </form>
      </article>
    </section>
  </div>
</template>
