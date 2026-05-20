<template>
  <div class="kanban-page">
    <!-- 顶部过滤区 -->
    <div class="filter-bar">
      <div class="filter-left">
        <el-input
          v-model="searchKey"
          placeholder="搜索任务标题或ID"
          prefix-icon="Search"
          clearable
          class="search-input"
        />
        <el-select v-model="filterAssignee" placeholder="负责人" clearable class="filter-select">
          <el-option label="张三 (前端)" value="张三" />
          <el-option label="李四 (后端)" value="李四" />
          <el-option label="王五 (测试)" value="王五" />
        </el-select>
        <el-select v-model="filterPriority" placeholder="优先级" clearable class="filter-select">
          <el-option label="高" value="high" />
          <el-option label="中" value="medium" />
          <el-option label="低" value="low" />
        </el-select>
      </div>
      <div class="filter-right">
        <el-button type="primary" icon="Plus">新建任务</el-button>
      </div>
    </div>

    <!-- 看板主体区 -->
    <div class="kanban-board">
      <div
        v-for="col in columns"
        :key="col.status"
        class="kanban-column"
        @dragover.prevent
        @drop="onDrop($event, col.status)"
      >
        <div class="column-header">
          <span class="column-title">{{ col.label }}</span>
          <el-tag size="small" type="info" class="task-count">
            {{ getTasksByStatus(col.status).length }}
          </el-tag>
        </div>

        <div class="column-body">
          <div
            v-for="task in getTasksByStatus(col.status)"
            :key="task.id"
            class="task-card"
            draggable="true"
            @dragstart="onDragStart($event, task)"
            @dragend="onDragEnd"
          >
            <div class="task-header">
              <span class="task-id">{{ task.id }}</span>
              <el-tag :type="getPriorityType(task.priority)" size="small" effect="plain">
                {{ getPriorityLabel(task.priority) }}
              </el-tag>
            </div>
            <div class="task-title">{{ task.title }}</div>
            
            <div class="task-footer">
              <div class="linked-feature" :title="task.feature">
                <el-icon><Link /></el-icon> {{ task.feature }}
              </div>
              <el-avatar :size="24" class="assignee-avatar" :title="task.assignee">
                {{ task.assignee.charAt(0) }}
              </el-avatar>
            </div>
          </div>
          
          <!-- 占位符，当列为空时显示，方便拖拽放入 -->
          <div v-if="getTasksByStatus(col.status).length === 0" class="empty-placeholder">
            拖拽至此
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue';
import { ElMessage } from 'element-plus';
import { Search, Plus, Link } from '@element-plus/icons-vue';

// 看板列定义 (根据 PRD 5.5)
const columns = [
  { label: '待开发', status: 'todo' },
  { label: '开发中', status: 'dev_doing' },
  { label: '待测试', status: 'test_todo' },
  { label: '测试中', status: 'test_doing' },
  { label: '待验收', status: 'acceptance_todo' },
  { label: '已验收', status: 'accepted' },
  { label: '已上线', status: 'online' }
];

// 模拟任务数据
const tasks = ref([
  { id: 'TASK-101', title: '前端登录页面及表单校验', status: 'todo', priority: 'high', assignee: '张三', feature: '账号密码登录' },
  { id: 'TASK-102', title: '登录接口开发及鉴权', status: 'dev_doing', priority: 'high', assignee: '李四', feature: '账号密码登录' },
  { id: 'TASK-103', title: '列表查询接口集成', status: 'test_todo', priority: 'medium', assignee: '李四', feature: '项目列表与搜索' },
  { id: 'TASK-104', title: '修改密码弹窗样式优化', status: 'test_doing', priority: 'low', assignee: '张三', feature: '找回密码' },
  { id: 'TASK-105', title: '项目创建表单提交接口', status: 'acceptance_todo', priority: 'high', assignee: '李四', feature: '新建项目' }
]);

// 过滤状态
const searchKey = ref('');
const filterAssignee = ref('');
const filterPriority = ref('');

// 获取过滤后的任务列表
const filteredTasks = computed(() => {
  return tasks.value.filter(task => {
    const matchKey = task.title.includes(searchKey.value) || task.id.includes(searchKey.value);
    const matchAssignee = filterAssignee.value ? task.assignee === filterAssignee.value : true;
    const matchPriority = filterPriority.value ? task.priority === filterPriority.value : true;
    return matchKey && matchAssignee && matchPriority;
  });
});

// 按状态分组获取任务
const getTasksByStatus = (status) => {
  return filteredTasks.value.filter(task => task.status === status);
};

// 拖拽逻辑
let draggedTask = null;

const onDragStart = (event, task) => {
  draggedTask = task;
  // 设置拖拽时的视觉效果
  event.dataTransfer.effectAllowed = 'move';
  setTimeout(() => {
    event.target.classList.add('dragging');
  }, 0);
};

const onDragEnd = (event) => {
  event.target.classList.remove('dragging');
  draggedTask = null;
};

const onDrop = (event, targetStatus) => {
  if (draggedTask && draggedTask.status !== targetStatus) {
    const oldStatus = columns.find(c => c.status === draggedTask.status)?.label;
    const newStatus = columns.find(c => c.status === targetStatus)?.label;
    
    // 更新状态
    draggedTask.status = targetStatus;
    
    // 在实际项目中，这里需要调用后端 API 保存状态
    ElMessage.success(`任务 ${draggedTask.id} 已从 [${oldStatus}] 移至 [${newStatus}]`);
  }
};

// 辅助格式化函数
const getPriorityType = (p) => {
  const map = { high: 'danger', medium: 'warning', low: 'info' };
  return map[p] || 'info';
};

const getPriorityLabel = (p) => {
  const map = { high: '高', medium: '中', low: '低' };
  return map[p] || '未知';
};
</script>

<style scoped>
.kanban-page {
  display: flex;
  flex-direction: column;
  height: calc(100vh - 110px);
  background: #fff;
  padding: 20px;
  border-radius: 8px;
  box-shadow: 0 1px 4px rgba(0,0,0,0.05);
}

.filter-bar {
  display: flex;
  justify-content: space-between;
  margin-bottom: 20px;
}
.filter-left { display: flex; gap: 12px; }
.search-input { width: 240px; }
.filter-select { width: 120px; }

.kanban-board {
  flex: 1;
  display: flex;
  gap: 16px;
  overflow-x: auto;
  padding-bottom: 8px; /* 给滚动条留出空间 */
}

.kanban-column {
  flex: 0 0 280px;
  background-color: #f4f5f7;
  border-radius: 6px;
  display: flex;
  flex-direction: column;
  max-height: 100%;
}

.column-header {
  padding: 12px 16px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 2px solid #e1e4e8;
}
.column-title { font-weight: 600; color: #172b4d; font-size: 14px; }

.column-body {
  flex: 1;
  padding: 12px;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.task-card { background: #fff; border-radius: 4px; padding: 12px; box-shadow: 0 1px 2px rgba(9, 30, 66, 0.25); cursor: grab; transition: transform 0.1s; border-left: 3px solid transparent; }
.task-card:hover { box-shadow: 0 2px 6px rgba(9, 30, 66, 0.15); }
.task-card.dragging { opacity: 0.5; border: 1px dashed #0052cc; }
.task-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px; }
.task-id { font-size: 12px; color: #5e6c84; font-weight: 600; }
.task-title { font-size: 14px; color: #172b4d; margin-bottom: 12px; line-height: 1.4; word-break: break-all; }
.task-footer { display: flex; justify-content: space-between; align-items: center; }
.linked-feature { font-size: 12px; color: #5e6c84; display: flex; align-items: center; gap: 4px; max-width: 180px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.assignee-avatar { background: #0052cc; color: #fff; font-size: 12px; }

.empty-placeholder { text-align: center; color: #a5adba; font-size: 12px; padding: 20px 0; border: 2px dashed #dfe1e6; border-radius: 4px; margin-top: 4px; }
</style>