<template>
  <div class="test-acceptance-page">
    <div class="page-header">
      <h2>测试与验收</h2>
      <div class="header-actions">
        <el-button type="primary" icon="VideoPlay" @click="triggerAITest" :loading="isAiTesting">
          触发 AI 自动化测试
        </el-button>
      </div>
    </div>

    <el-tabs v-model="activeTab" class="custom-tabs">
      <!-- Tab 1: 测试用例与 AI 复核 -->
      <el-tab-pane label="测试用例与复核" name="testcases">
        <div class="filter-bar">
          <el-radio-group v-model="reviewFilter" size="small">
            <el-radio-button label="all">全部</el-radio-button>
            <el-radio-button label="pending">待复核 (AI已执行)</el-radio-button>
            <el-radio-button label="done">已复核</el-radio-button>
          </el-radio-group>
          <el-input v-model="searchKey" placeholder="搜索用例..." prefix-icon="Search" class="search-input" clearable />
        </div>

        <el-table :data="filteredTestCases" style="width: 100%" border>
          <el-table-column prop="id" label="用例编号" width="100" />
          <el-table-column prop="title" label="用例标题" min-width="200" show-overflow-tooltip />
          <el-table-column prop="feature" label="关联功能点" min-width="150" show-overflow-tooltip />
          <el-table-column prop="priority" label="优先级" width="80">
            <template #default="{ row }">
              <el-tag :type="getPriorityType(row.priority)" size="small">{{ row.priority }}</el-tag>
            </template>
          </el-table-column>
          
          <!-- AI 执行结果列 -->
          <el-table-column label="AI 执行结果" width="140">
            <template #default="{ row }">
              <div class="ai-result-cell">
                <el-icon v-if="row.aiResult === 'pass'" color="#67C23A"><CircleCheckFilled /></el-icon>
                <el-icon v-else-if="row.aiResult === 'fail'" color="#F56C6C"><CircleCloseFilled /></el-icon>
                <el-icon v-else color="#909399"><InfoFilled /></el-icon>
                <span :class="`text-${row.aiResult}`">{{ getAiResultText(row.aiResult) }}</span>
              </div>
            </template>
          </el-table-column>

          <!-- 人工复核状态列 -->
          <el-table-column label="人工复核状态" width="120">
            <template #default="{ row }">
              <el-tag :type="getReviewStatusType(row.reviewStatus)" effect="light">
                {{ getReviewStatusText(row.reviewStatus) }}
              </el-tag>
            </template>
          </el-table-column>

          <el-table-column label="操作" width="120" fixed="right">
            <template #default="{ row }">
              <el-button 
                v-if="row.reviewStatus === 'pending' && row.aiResult !== 'none'" 
                type="primary" link size="small" 
                @click="openReviewDrawer(row)"
              >
                去复核
              </el-button>
              <el-button v-else type="info" link size="small" @click="openReviewDrawer(row)">
                查看详情
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <!-- Tab 2: 验收检查项 -->
      <el-tab-pane label="验收检查项" name="acceptance">
        <el-empty description="验收检查项模块开发中..." />
      </el-tab-pane>
    </el-tabs>

    <!-- 人工复核抽屉 -->
    <el-drawer
      v-model="drawerVisible"
      title="测试结果复核"
      size="600px"
      destroy-on-close
    >
      <template v-if="currentCase">
        <div class="drawer-section">
          <h4>用例信息</h4>
          <el-descriptions :column="1" border size="small">
            <el-descriptions-item label="编号">{{ currentCase.id }}</el-descriptions-item>
            <el-descriptions-item label="标题">{{ currentCase.title }}</el-descriptions-item>
            <el-descriptions-item label="前置条件">{{ currentCase.precondition }}</el-descriptions-item>
            <el-descriptions-item label="操作步骤">
              <div style="white-space: pre-wrap;">{{ currentCase.steps }}</div>
            </el-descriptions-item>
            <el-descriptions-item label="预期结果">{{ currentCase.expected }}</el-descriptions-item>
          </el-descriptions>
        </div>

        <div class="drawer-section ai-evidence-section">
          <h4>
            AI 执行证据 
            <el-tag size="small" :type="currentCase.aiResult === 'pass' ? 'success' : 'danger'">
              判定: {{ getAiResultText(currentCase.aiResult) }}
            </el-tag>
          </h4>
          <div class="evidence-box">
            <p class="actual-result"><strong>实际结果：</strong>{{ currentCase.actual }}</p>
            <div class="screenshot-placeholder" v-if="currentCase.hasScreenshot">
              <el-icon :size="40" color="#c0c4cc"><Picture /></el-icon>
              <span>AI 截图证据 .png</span>
            </div>
            <div class="log-placeholder" v-if="currentCase.log">
              <code>{{ currentCase.log }}</code>
            </div>
          </div>
        </div>

        <!-- 抽屉底部操作区 -->
        <div class="drawer-footer" v-if="currentCase.reviewStatus === 'pending'">
          <el-divider content-position="left">人工复核结论</el-divider>
          <div class="action-buttons">
            <el-button type="success" @click="submitReview('passed')">确认通过</el-button>
            <el-button type="danger" @click="submitReview('failed')">不通过并建缺陷</el-button>
            <el-button type="warning" plain @click="submitReview('manual_retest')">需人工复测</el-button>
          </div>
        </div>
        <div class="drawer-footer" v-else>
          <el-alert 
            :title="`该用例已复核，结论为：${getReviewStatusText(currentCase.reviewStatus)}`" 
            :type="getReviewStatusType(currentCase.reviewStatus)" 
            show-icon 
            :closable="false" 
          />
        </div>
      </template>
    </el-drawer>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { VideoPlay, Search, CircleCheckFilled, CircleCloseFilled, InfoFilled, Picture } from '@element-plus/icons-vue';

const activeTab = ref('testcases');
const isAiTesting = ref(false);
const searchKey = ref('');
const reviewFilter = ref('pending');

const drawerVisible = ref(false);
const currentCase = ref(null);

// 模拟测试用例数据 (包含 AI 执行结果和复核状态)
const testCases = ref([
  {
    id: 'TC-001', title: '正确输入账号密码，成功登录', feature: '账号密码登录', priority: '高',
    precondition: '系统已存在测试账号 testuser / 123456',
    steps: '1. 打开登录页\n2. 输入账号 testuser\n3. 输入密码 123456\n4. 点击登录按钮',
    expected: '页面跳转至工作台，右上角显示用户信息',
    aiResult: 'pass', actual: '跳转成功，API /api/user/info 返回 200，并渲染了用户信息。', 
    hasScreenshot: true, log: null, reviewStatus: 'pending'
  },
  {
    id: 'TC-002', title: '密码错误时，提示错误信息', feature: '账号密码登录', priority: '高',
    precondition: '系统已存在测试账号 testuser',
    steps: '1. 打开登录页\n2. 输入账号 testuser\n3. 输入错误密码 111111\n4. 点击登录按钮',
    expected: '登录失败，表单提示"账号或密码错误"',
    aiResult: 'fail', actual: '登录失败，但页面没有出现弹窗提示，API 返回了 401 错误。', 
    hasScreenshot: true, log: '[Error] /api/login 401 Unauthorized. UI Notification not found in DOM.', 
    reviewStatus: 'pending'
  },
  {
    id: 'TC-003', title: '列表查询按负责人筛选', feature: '项目列表与搜索', priority: '中',
    precondition: '存在多个不同负责人的项目数据',
    steps: '1. 进入项目列表页\n2. 负责人下拉选择"张三"\n3. 点击查询',
    expected: '列表仅展示负责人为张三的项目',
    aiResult: 'pass', actual: '筛选成功，列表返回 2 条张三的项目数据。', 
    hasScreenshot: false, log: null, reviewStatus: 'passed'
  },
  {
    id: 'TC-004', title: '边界测试：项目名称超长截断', feature: '新建项目', priority: '低',
    precondition: '进入新建项目页面',
    steps: '1. 输入超过 50 个字符的项目名称\n2. 点击保存',
    expected: '输入框应限制输入或提示超出长度',
    aiResult: 'none', actual: '', 
    hasScreenshot: false, log: null, reviewStatus: 'unreviewed'
  }
]);

// 计算属性：过滤用例
const filteredTestCases = computed(() => {
  return testCases.value.filter(tc => {
    // 搜索过滤
    const matchSearch = tc.title.includes(searchKey.value) || tc.id.includes(searchKey.value);
    // 状态过滤
    let matchFilter = true;
    if (reviewFilter.value === 'pending') matchFilter = tc.reviewStatus === 'pending' && tc.aiResult !== 'none';
    if (reviewFilter.value === 'done') matchFilter = tc.reviewStatus === 'passed' || tc.reviewStatus === 'failed' || tc.reviewStatus === 'manual_retest';
    
    return matchSearch && matchFilter;
  });
});

// 辅助格式化函数
const getPriorityType = (p) => ({ '高': 'danger', '中': 'warning', '低': 'info' }[p] || 'info');

const getAiResultText = (status) => {
  const map = { 'pass': 'AI测试通过', 'fail': 'AI发现异常', 'none': '未执行' };
  return map[status] || status;
};

const getReviewStatusText = (status) => {
  const map = {
    'pending': '待复核',
    'passed': '已复核通过',
    'failed': '复核不通过',
    'manual_retest': '需人工复测',
    'unreviewed': '未开始'
  };
  return map[status] || status;
};

const getReviewStatusType = (status) => {
  const map = {
    'pending': 'warning',
    'passed': 'success',
    'failed': 'danger',
    'manual_retest': 'info',
    'unreviewed': 'info'
  };
  return map[status] || 'info';
};

// 交互逻辑
const triggerAITest = () => {
  isAiTesting.value = true;
  setTimeout(() => {
    isAiTesting.value = false;
    ElMessage.success('AI 测试执行完毕，发现 1 处异常，请进行人工复核。');
  }, 2000);
};

const openReviewDrawer = (row) => {
  currentCase.value = { ...row };
  drawerVisible.value = true;
};

const submitReview = (action) => {
  if (action === 'failed') {
    ElMessageBox.confirm(
      '确认将该用例标记为不通过，并为其创建一个缺陷吗？',
      '建缺陷提示',
      { confirmButtonText: '确定并创建', cancelButtonText: '取消', type: 'warning' }
    ).then(() => {
      finalizeReview(action, '已生成草稿缺陷 BUG-101，用例状态已更新。');
    }).catch(() => {});
  } else {
    finalizeReview(action, '复核成功！');
  }
};

const finalizeReview = (action, msg) => {
  // 更新本地数据状态模拟后端请求
  const index = testCases.value.findIndex(tc => tc.id === currentCase.value.id);
  if (index !== -1) {
    testCases.value[index].reviewStatus = action;
  }
  ElMessage.success(msg);
  drawerVisible.value = false;
};
</script>

<style scoped>
.test-acceptance-page { background: #fff; padding: 20px; border-radius: 8px; box-shadow: 0 1px 4px rgba(0,0,0,0.05); min-height: calc(100vh - 110px); }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.page-header h2 { margin: 0; font-size: 18px; color: #303133; }
.custom-tabs :deep(.el-tabs__content) { padding-top: 16px; }
.filter-bar { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.search-input { width: 260px; }

.ai-result-cell { display: flex; align-items: center; gap: 6px; font-weight: 500; }
.text-pass { color: #67C23A; }
.text-fail { color: #F56C6C; }
.text-none { color: #909399; }

.drawer-section { margin-bottom: 24px; }
.drawer-section h4 { margin: 0 0 12px 0; font-size: 15px; color: #303133; display: flex; align-items: center; justify-content: space-between; }
.evidence-box { background: #f8f9fa; border: 1px solid #ebeef5; border-radius: 4px; padding: 16px; }
.actual-result { margin: 0 0 12px 0; font-size: 14px; color: #606266; }
.screenshot-placeholder { border: 1px dashed #dcdfe6; background: #fff; border-radius: 4px; height: 120px; display: flex; flex-direction: column; align-items: center; justify-content: center; color: #909399; font-size: 12px; margin-bottom: 12px; }
.screenshot-placeholder span { margin-top: 8px; }
.log-placeholder { background: #282c34; color: #abb2bf; padding: 12px; border-radius: 4px; font-size: 12px; overflow-x: auto; }
.drawer-footer { margin-top: 32px; }
.action-buttons { display: flex; gap: 12px; justify-content: center; }
</style>