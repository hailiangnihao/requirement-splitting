<template>
  <div class="change-workspace">
    <!-- 左侧：变更内容录入区 -->
    <div class="left-panel">
      <div class="panel-header">
        <h3>提交需求变更</h3>
        <el-button link type="primary">查看历史变更记录</el-button>
      </div>
      
      <div class="form-container">
        <el-form :model="changeForm" label-position="top" class="change-form">
          <el-form-item label="变更标题" required>
            <el-input v-model="changeForm.title" placeholder="例如：新增飞书扫码登录功能" />
          </el-form-item>
          
          <el-form-item label="变更原因" required>
            <el-input 
              v-model="changeForm.reason" 
              type="textarea" 
              :rows="3" 
              placeholder="例如：为方便企业内部员工快速登录，提升系统易用性。" 
            />
          </el-form-item>
          
          <el-form-item label="变更具体内容" required class="content-item">
            <el-input 
              v-model="changeForm.content" 
              type="textarea" 
              placeholder="请详细描述需要修改、增加或删除的业务逻辑..." 
            />
          </el-form-item>

          <el-form-item label="期望生效里程碑">
            <el-select v-model="changeForm.milestone" placeholder="请选择" style="width: 100%">
              <el-option label="V1.0 - 基础功能上线 (当前)" value="v1.0" />
              <el-option label="V1.1 - 体验优化版" value="v1.1" />
            </el-select>
          </el-form-item>
        </el-form>
      </div>

      <div class="panel-footer">
        <el-button type="primary" size="large" @click="handleAnalyze" :loading="isAnalyzing" class="analyze-btn">
          <el-icon class="el-icon--left"><MagicStick /></el-icon>
          AI 分析变更影响范围
        </el-button>
      </div>
    </div>

    <!-- 右侧：AI 影响分析结果区 -->
    <div class="right-panel">
      <div class="panel-header">
        <div class="title-with-badge">
          <h3>AI 影响分析</h3>
          <el-tag v-if="hasResult" type="danger" size="small" effect="light">
            影响程度: {{ impactData.level }}
          </el-tag>
        </div>
        <div class="actions" v-if="hasResult">
          <el-button size="small" @click="resetAnalysis">重新分析</el-button>
          <el-button type="success" size="small" @click="confirmChange">
            确认变更并生成补充草稿
          </el-button>
        </div>
      </div>

      <div class="analysis-container" v-loading="isAnalyzing" element-loading-text="AI 正在穿透分析项目链路...">
        <el-empty 
          v-if="!hasResult && !isAnalyzing" 
          description="请在左侧录入变更内容，点击分析后将在此展示受影响的模块、任务和测试用例" 
          :image-size="120"
        />

        <div v-if="hasResult" class="impact-content">
          <!-- 总体影响卡片 -->
          <div class="impact-summary">
            <div class="summary-item">
              <div class="val">{{ impactData.modules.length }}</div>
              <div class="lbl">受影响模块</div>
            </div>
            <div class="summary-item">
              <div class="val text-warning">{{ impactData.tasks.length }}</div>
              <div class="lbl">受影响任务</div>
            </div>
            <div class="summary-item">
              <div class="val text-danger">{{ impactData.testCases.length }}</div>
              <div class="lbl">波及测试用例</div>
            </div>
          </div>

          <!-- 详细分析折叠面板 -->
          <el-collapse v-model="activeNames" class="custom-collapse">
            <el-collapse-item name="modules">
              <template #title>
                <span class="collapse-title">
                  <el-icon><Menu /></el-icon> 业务模块影响分析
                </span>
              </template>
              <div v-for="(mod, index) in impactData.modules" :key="index" class="impact-item">
                <div class="item-header">
                  <el-tag size="small" :type="mod.action === '新增' ? 'success' : 'warning'">
                    {{ mod.action }}功能点
                  </el-tag>
                  <span class="item-title">{{ mod.title }}</span>
                </div>
                <div class="item-desc">{{ mod.reason }}</div>
              </div>
            </el-collapse-item>

            <el-collapse-item name="tasks">
              <template #title>
                <span class="collapse-title">
                  <el-icon><List /></el-icon> 研发任务影响建议
                </span>
              </template>
              <div v-for="(task, index) in impactData.tasks" :key="index" class="impact-item">
                <div class="item-header">
                  <el-tag size="small" :type="task.action === '新增' ? 'success' : 'danger'">
                    {{ task.action }}任务
                  </el-tag>
                  <span class="item-title">{{ task.title }}</span>
                </div>
                <div class="item-desc"><strong>当前状态：</strong>{{ task.status }} | <strong>负责人：</strong>{{ task.assignee || '待定' }}</div>
                <div class="item-desc">{{ task.reason }}</div>
              </div>
            </el-collapse-item>

            <el-collapse-item name="testcases">
              <template #title>
                <span class="collapse-title">
                  <el-icon><CircleCheck /></el-icon> 测试用例及验收影响
                </span>
              </template>
              <div v-for="(tc, index) in impactData.testCases" :key="index" class="impact-item">
                <div class="item-header">
                  <el-tag size="small" :type="tc.action === '需修改' ? 'warning' : 'info'">
                    {{ tc.action }}
                  </el-tag>
                  <span class="item-title">{{ tc.title }}</span>
                </div>
                <div class="item-desc"><strong>关联编号：</strong>{{ tc.id }}</div>
                <div class="item-desc">{{ tc.reason }}</div>
              </div>
            </el-collapse-item>
          </el-collapse>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue';
import { useRoute } from 'vue-router';
import { ElMessage } from 'element-plus';
import { MagicStick, Menu, List, CircleCheck } from '@element-plus/icons-vue';
import { api } from '../../api/client';

const route = useRoute();
// 左侧表单数据
const changeForm = reactive({
  title: '增加飞书扫码登录功能',
  reason: '为了方便企业内部员工快速登录，提升系统易用性，减少密码遗忘导致的客诉。',
  content: '1. 在原有的账号密码登录页面，增加“飞书扫码”切换Tab。\n2. 对接飞书 OAuth2.0 授权接口。\n3. 若扫码的用户在系统内不存在，则自动创建一个无权限的访客账号，并提示“请联系管理员授权”。',
  milestone: 'v1.0'
});

const isAnalyzing = ref(false);
const hasResult = ref(false);
const activeNames = ref(['modules', 'tasks', 'testcases']);
const changes = ref([]);

// 模拟 AI 分析出的结构化影响数据
const impactData = reactive({
  level: '中等 (Medium)',
  modules: [],
  tasks: [],
  testCases: []
});

const loadChanges = async () => {
  try {
    changes.value = await api.listChanges(route.params.id);
  } catch (error) {
    ElMessage.error(error.message || '变更列表加载失败');
  }
};

onMounted(loadChanges);

const handleAnalyze = async () => {
  if (!changeForm.title || !changeForm.content) {
    ElMessage.warning('请完整填写变更标题和具体内容');
    return;
  }

  isAnalyzing.value = true;
  hasResult.value = false;
  try {
    const change = await api.createChange(route.params.id, {
      title: changeForm.title,
      content: `${changeForm.reason}\n\n${changeForm.content}`,
      created_by: ''
    });
    await api.analyzeChange(route.params.id, change.id);
    await loadChanges();
    const analyzed = changes.value.find(item => item.id === change.id);
    const analysis = analyzed?.impact_analysis || {};
    impactData.level = analysis.risk_level || 'medium';
    impactData.modules = (analysis.affected_feature_points || []).map(id => ({ action: '影响', title: id, reason: analysis.summary || 'AI 判断该功能点会受到影响。' }));
    impactData.tasks = (analysis.new_tasks_suggested || []).map(task => ({ action: '新增', title: task.title, status: '待排期', assignee: '', reason: task.description }));
    impactData.testCases = (analysis.affected_test_cases || []).map(item => ({ action: item.action, id: item.test_case_id, title: item.test_case_id, reason: item.reason }));
    hasResult.value = true;
    ElMessage.success('AI 影响分析完成');
  } catch (error) {
    ElMessage.error(error.message || '变更影响分析失败');
  } finally {
    isAnalyzing.value = false;
  }
};

const resetAnalysis = () => {
  hasResult.value = false;
};

const confirmChange = async () => {
  const latest = changes.value[0];
  if (!latest) return;
  try {
    await api.updateChangeStatus(route.params.id, latest.id, 'accepted');
    ElMessage.success('变更已确认');
    await loadChanges();
  } catch (error) {
    ElMessage.error(error.message || '变更确认失败');
  }
};
</script>

<style scoped>
.change-workspace { display: flex; gap: 24px; height: calc(100vh - 110px); }
.left-panel, .right-panel { flex: 1; background: #fff; border-radius: 8px; padding: 20px; display: flex; flex-direction: column; box-shadow: 0 1px 4px rgba(0, 0, 0, 0.05); min-width: 0; }
.panel-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; border-bottom: 1px solid #f0f2f5; padding-bottom: 12px; }
.panel-header h3 { margin: 0; font-size: 16px; color: #1f2f3d; }
.title-with-badge { display: flex; align-items: center; gap: 8px; }

.form-container { flex: 1; overflow-y: auto; padding-right: 8px; }
.change-form :deep(.el-form-item__label) { font-weight: 500; color: #303133; padding-bottom: 4px; }
.content-item { flex: 1; display: flex; flex-direction: column; }
.content-item :deep(.el-form-item__content) { flex: 1; display: flex; }
.content-item :deep(.el-textarea) { flex: 1; display: flex; }
.content-item :deep(.el-textarea__inner) { flex: 1; resize: none; }

.panel-footer { margin-top: 16px; padding-top: 16px; border-top: 1px solid #f0f2f5; display: flex; justify-content: flex-end; }
.analyze-btn { width: 100%; }

.analysis-container { flex: 1; overflow-y: auto; }
.impact-content { display: flex; flex-direction: column; gap: 20px; }

.impact-summary { display: flex; gap: 16px; margin-bottom: 8px; }
.summary-item { flex: 1; background: #f8f9fa; border: 1px solid #ebeef5; border-radius: 6px; padding: 16px; text-align: center; }
.summary-item .val { font-size: 24px; font-weight: bold; color: #409EFF; margin-bottom: 4px; }
.summary-item .val.text-warning { color: #E6A23C; }
.summary-item .val.text-danger { color: #F56C6C; }
.summary-item .lbl { font-size: 13px; color: #606266; }

.custom-collapse { border-top: none; }
.collapse-title { display: flex; align-items: center; gap: 8px; font-size: 15px; font-weight: 500; color: #303133; }
.impact-item { padding: 12px; background: #f8f9fa; border-radius: 4px; margin-bottom: 12px; border: 1px solid #ebeef5; }
.impact-item:last-child { margin-bottom: 0; }
.item-header { display: flex; align-items: center; gap: 8px; margin-bottom: 8px; }
.item-title { font-weight: 500; color: #303133; font-size: 14px; }
.item-desc { font-size: 13px; color: #606266; line-height: 1.5; margin-top: 4px; }
</style>
