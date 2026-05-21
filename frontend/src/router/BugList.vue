<template>
  <div class="bug-list-page">
    <div class="page-header">
      <h2>缺陷管理</h2>
      <div class="header-actions">
        <el-button type="primary" icon="Plus" @click="handleCreateDefect">提缺陷</el-button>
      </div>
    </div>

    <!-- 过滤栏 -->
    <div class="filter-bar">
      <div class="filter-left">
        <el-input
          v-model="searchKey"
          placeholder="搜索缺陷标题或ID..."
          prefix-icon="Search"
          clearable
          class="search-input"
        />
        <el-select v-model="filterStatus" placeholder="状态" clearable class="filter-select">
          <el-option label="待确认" value="pending_confirm" />
          <el-option label="待修复" value="pending_fix" />
          <el-option label="修复中" value="fixing" />
          <el-option label="待回归" value="pending_regression" />
          <el-option label="回归通过" value="regression_passed" />
          <el-option label="已关闭" value="closed" />
        </el-select>
        <el-select v-model="filterSeverity" placeholder="严重程度" clearable class="filter-select">
          <el-option label="致命 (Critical)" value="critical" />
          <el-option label="严重 (High)" value="high" />
          <el-option label="一般 (Medium)" value="medium" />
          <el-option label="轻微 (Low)" value="low" />
        </el-select>
        <el-checkbox v-model="filterBlocking" class="blocking-checkbox">仅看阻塞验收</el-checkbox>
      </div>
    </div>

    <!-- 缺陷列表 -->
    <el-table :data="filteredBugs" style="width: 100%" border hover>
      <el-table-column prop="id" label="缺陷编号" width="100" />
      
      <el-table-column label="标题" min-width="250" show-overflow-tooltip>
        <template #default="{ row }">
          <div class="title-cell" @click="openBugDetail(row)">
            <span class="bug-title">{{ row.title }}</span>
            <el-tag v-if="row.blockAcceptance" type="danger" size="small" effect="dark" class="block-tag">
              阻塞验收
            </el-tag>
          </div>
        </template>
      </el-table-column>
      
      <el-table-column label="严重程度" width="100">
        <template #default="{ row }">
          <span :class="['severity-dot', row.severity]"></span>
          {{ getSeverityText(row.severity) }}
        </template>
      </el-table-column>

      <el-table-column label="状态" width="120">
        <template #default="{ row }">
          <el-tag :type="getStatusType(row.status)">{{ getStatusText(row.status) }}</el-tag>
        </template>
      </el-table-column>

      <el-table-column prop="assignee" label="当前处理人" width="100">
        <template #default="{ row }">
          <div class="assignee-cell">
            <el-avatar :size="20" class="mini-avatar">{{ row.assignee.charAt(0) }}</el-avatar>
            <span>{{ row.assignee }}</span>
          </div>
        </template>
      </el-table-column>
      
      <el-table-column prop="reporter" label="提出人" width="80" />
      <el-table-column prop="createTime" label="创建时间" width="150" />

      <el-table-column label="操作" width="100" fixed="right">
        <template #default="{ row }">
          <el-button type="primary" link size="small" @click="openBugDetail(row)">
            处理
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 缺陷详情与流转抽屉 -->
    <el-drawer
      v-model="drawerVisible"
      :title="`${currentBug?.id} 详情及处理`"
      size="600px"
      destroy-on-close
    >
      <template v-if="currentBug">
        <!-- 当前状态高亮 -->
        <div class="bug-status-banner">
          <div class="status-left">
            <span class="label">当前状态：</span>
            <el-tag :type="getStatusType(currentBug.status)" size="large" effect="dark">
              {{ getStatusText(currentBug.status) }}
            </el-tag>
          </div>
          <div class="status-right" v-if="currentBug.blockAcceptance">
            <el-icon color="#F56C6C" :size="16"><WarnTriangleFilled /></el-icon>
            <span class="block-text">此缺陷阻塞业务验收，请优先处理！</span>
          </div>
        </div>

        <!-- 基础信息 -->
        <div class="drawer-section">
          <h3 class="bug-detail-title">{{ currentBug.title }}</h3>
          <el-descriptions :column="2" border size="small" class="mt-3">
            <el-descriptions-item label="严重程度">
              <span :class="['severity-dot', currentBug.severity]"></span>
              {{ getSeverityText(currentBug.severity) }}
            </el-descriptions-item>
            <el-descriptions-item label="优先级">{{ currentBug.priority }}</el-descriptions-item>
            <el-descriptions-item label="处理人">{{ currentBug.assignee }}</el-descriptions-item>
            <el-descriptions-item label="提出人">{{ currentBug.reporter }}</el-descriptions-item>
            <el-descriptions-item label="关联用例" :span="2">
              <el-link type="primary" v-if="currentBug.linkedCase">{{ currentBug.linkedCase }}</el-link>
              <span v-else>-</span>
            </el-descriptions-item>
          </el-descriptions>
        </div>

        <!-- 描述与复现步骤 -->
        <div class="drawer-section">
          <h4>复现步骤与证据</h4>
          <div class="info-box">
            <p><strong>描述：</strong><br/>{{ currentBug.description }}</p>
            <el-divider border-style="dashed" />
            <p><strong>实际结果：</strong><br/>{{ currentBug.actualResult }}</p>
            <p><strong>预期结果：</strong><br/>{{ currentBug.expectedResult }}</p>
          </div>
        </div>

        <!-- 状态流转操作区 -->
        <div class="drawer-footer">
          <el-divider content-position="left">状态流转</el-divider>
          <div class="action-buttons">
            <template v-if="availableActions.length > 0">
              <el-button 
                v-for="action in availableActions" 
                :key="action.status" 
                :type="action.type" 
                :plain="action.plain"
                @click="updateBugStatus(action.status, action.label)"
              >
                {{ action.label }}
              </el-button>
            </template>
            <el-alert 
              v-else 
              title="该缺陷已完结或处于非活跃状态，无可用操作。" 
              type="info" 
              show-icon 
              :closable="false" 
            />
          </div>
        </div>
      </template>
    </el-drawer>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue';
import { useRoute } from 'vue-router';
import { ElMessage, ElMessageBox } from 'element-plus';
import { Search, Plus, WarnTriangleFilled } from '@element-plus/icons-vue';
import { api } from '../api/client';

const route = useRoute();
// 过滤状态
const searchKey = ref('');
const filterStatus = ref('');
const filterSeverity = ref('');
const filterBlocking = ref(false);

const drawerVisible = ref(false);
const currentBug = ref(null);

const bugs = ref([]);

const loadBugs = async () => {
  try {
    const data = await api.listDefects(route.params.id);
    bugs.value = (data || []).map(defect => ({
      id: defect.id,
      title: defect.title,
      status: defect.status,
      severity: 'high',
      priority: '高',
      assignee: defect.assigned_to || 'PM',
      reporter: defect.created_by || 'AI',
      blockAcceptance: false,
      linkedCase: defect.test_run_id || '',
      createTime: defect.created_at ? new Date(defect.created_at).toLocaleString() : '-',
      description: defect.description,
      actualResult: defect.description,
      expectedResult: '-'
    }));
  } catch (error) {
    ElMessage.error(error.message || '缺陷列表加载失败');
  }
};

// 创建缺陷
const handleCreateDefect = () => {
  ElMessageBox.prompt('请输入缺陷标题', '创建缺陷', {
    confirmButtonText: '创建',
    cancelButtonText: '取消',
    inputPlaceholder: '例如：登录按钮点击无响应',
    inputValidator: (value) => {
      if (!value || !value.trim()) {
        return '缺陷标题不能为空';
      }
      return true;
    }
  }).then(async ({ value }) => {
    try {
      await api.createDefect(route.params.id, {
        title: value.trim(),
        description: '通过前端创建的缺陷'
        // 不传 created_by，让后端处理
      });
      await loadBugs();
      ElMessage.success('缺陷创建成功');
    } catch (error) {
      ElMessage.error(error.message || '缺陷创建失败');
    }
  }).catch(() => {
    // 用户取消
  });
};

onMounted(loadBugs);

// 计算过滤后的列表
const filteredBugs = computed(() => {
  return bugs.value.filter(bug => {
    const matchKey = bug.title.includes(searchKey.value) || bug.id.includes(searchKey.value);
    const matchStatus = filterStatus.value ? bug.status === filterStatus.value : true;
    const matchSeverity = filterSeverity.value ? bug.severity === filterSeverity.value : true;
    const matchBlock = filterBlocking.value ? bug.blockAcceptance === true : true;
    return matchKey && matchStatus && matchSeverity && matchBlock;
  });
});

// 辅助格式化
const getSeverityText = (s) => ({ 'critical': '致命', 'high': '严重', 'medium': '一般', 'low': '轻微' }[s] || s);
const getStatusText = (s) => {
  const map = {
    'pending_confirm': '待确认', 'pending_fix': '待修复', 'fixing': '修复中',
    'pending_regression': '待回归', 'regression_passed': '回归通过', 'closed': '已关闭', 'rejected': '拒绝/非缺陷'
  };
  return map[s] || s;
};
const getStatusType = (s) => {
  const map = {
    'pending_confirm': 'warning', 'pending_fix': 'danger', 'fixing': 'primary',
    'pending_regression': 'warning', 'regression_passed': 'success', 'closed': 'info', 'rejected': 'info'
  };
  return map[s] || 'info';
};

// 状态机：根据当前状态计算可以流转的下一个状态
const availableActions = computed(() => {
  if (!currentBug.value) return [];
  switch(currentBug.value.status) {
    case 'pending_confirm': return [{label: '确认缺陷', status: 'pending_fix', type: 'primary'}, {label: '非缺陷拒绝', status: 'rejected', type: 'danger', plain: true}];
    case 'pending_fix': return [{label: '开始修复', status: 'fixing', type: 'primary'}];
    case 'fixing': return [{label: '修复完成 (提测)', status: 'pending_regression', type: 'success'}];
    case 'pending_regression': return [{label: '回归通过', status: 'regression_passed', type: 'success'}, {label: '回归不通过', status: 'pending_fix', type: 'danger'}];
    case 'regression_passed': return [{label: '关闭缺陷', status: 'closed', type: 'info'}];
    default: return [];
  }
});

// 交互事件
const openBugDetail = (row) => {
  currentBug.value = { ...row };
  drawerVisible.value = true;
};

const updateBugStatus = (newStatus, actionLabel) => {
  ElMessageBox.confirm(`确定要执行【${actionLabel}】操作吗？`, '提示', { type: 'warning' }).then(() => {
    api.updateDefectStatus(route.params.id, currentBug.value.id, newStatus).then(async () => {
      await loadBugs();
      currentBug.value.status = newStatus;
      ElMessage.success(`操作成功！缺陷状态已流转至 [${getStatusText(newStatus)}]`);
    }).catch(error => {
      ElMessage.error(error.message || '缺陷状态更新失败');
    });
  }).catch(() => {});
};
</script>

<style scoped>
.bug-list-page { background: #fff; padding: 20px; border-radius: 8px; box-shadow: 0 1px 4px rgba(0,0,0,0.05); min-height: calc(100vh - 110px); }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.page-header h2 { margin: 0; font-size: 18px; color: #303133; }
.filter-bar { display: flex; margin-bottom: 16px; }
.filter-left { display: flex; gap: 12px; align-items: center; }
.search-input { width: 220px; }
.filter-select { width: 130px; }
.blocking-checkbox { margin-left: 8px; }

.title-cell { display: flex; align-items: center; gap: 8px; cursor: pointer; }
.title-cell:hover .bug-title { color: #409EFF; text-decoration: underline; }
.bug-title { font-weight: 500; color: #303133; }
.block-tag { transform: scale(0.9); }
.severity-dot { display: inline-block; width: 8px; height: 8px; border-radius: 50%; margin-right: 6px; }
.severity-dot.critical { background-color: #722ed1; }
.severity-dot.high { background-color: #F56C6C; }
.severity-dot.medium { background-color: #E6A23C; }
.severity-dot.low { background-color: #909399; }
.assignee-cell { display: flex; align-items: center; gap: 6px; }
.mini-avatar { background: #409EFF; font-size: 12px; }

.bug-status-banner { display: flex; justify-content: space-between; align-items: center; background: #f0f9eb; padding: 12px 16px; border-radius: 4px; margin-bottom: 20px; border-left: 4px solid #67C23A; }
.bug-status-banner:has(.el-icon) { background: #fef0f0; border-left-color: #F56C6C; }
.status-left { display: flex; align-items: center; }
.status-left .label { font-size: 14px; color: #606266; margin-right: 8px; }
.status-right { display: flex; align-items: center; gap: 6px; color: #F56C6C; font-weight: 500; font-size: 13px; }
.drawer-section { margin-bottom: 24px; }
.drawer-section h4 { margin: 0 0 12px 0; font-size: 15px; color: #303133; }
.bug-detail-title { margin: 0 0 16px 0; font-size: 18px; color: #1f2f3d; }
.info-box { background: #f8f9fa; padding: 16px; border-radius: 4px; font-size: 14px; color: #606266; line-height: 1.6; border: 1px solid #ebeef5; }
.drawer-footer { margin-top: 32px; }
.action-buttons { display: flex; gap: 12px; }
</style>
