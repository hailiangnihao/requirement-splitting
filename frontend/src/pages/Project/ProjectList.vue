<template>
  <div class="project-list-page">
    <!-- 顶部操作区 -->
    <div class="page-header">
      <h2>项目空间</h2>
      <div class="header-actions">
        <el-input
          v-model="searchKey"
          placeholder="搜索项目名称..."
          prefix-icon="Search"
          clearable
          class="search-input"
        />
        <el-select v-model="filterStatus" placeholder="全部状态" clearable class="filter-select">
          <el-option label="待开始" value="pending" />
          <el-option label="进行中" value="doing" />
          <el-option label="已延期" value="delayed" />
          <el-option label="已交付" value="done" />
        </el-select>
        <el-button type="primary" icon="Plus" @click="openWizard">新建项目</el-button>
      </div>
    </div>

    <!-- 项目卡片网格 -->
    <div class="project-grid">
      <el-card 
        v-for="project in filteredProjects" 
        :key="project.id" 
        class="project-card" 
        shadow="hover"
        @click="goToProject(project.id)"
      >
        <!-- 卡片封面/缩略图 -->
        <div :class="['card-cover', project.coverStyle]">
          <div class="cover-content">
            <h3>{{ project.name }}</h3>
            <el-tag :type="getStatusType(project.status)" size="small" effect="dark" class="status-tag">
              {{ project.status }}
            </el-tag>
          </div>
        </div>
        
        <!-- 卡片内容 -->
        <div class="card-body">
          <div class="info-row">
            <span class="label">负责人：</span>
            <span class="value flex-center">
              <el-avatar :size="20" class="mini-avatar">{{ project.manager.charAt(0) }}</el-avatar>
              {{ project.manager }}
            </span>
          </div>
          
          <div class="progress-section">
            <div class="progress-header">
              <span class="label">整体进度</span>
              <span class="percentage">{{ project.progress }}%</span>
            </div>
            <el-progress 
              :percentage="project.progress" 
              :color="getHealthColor(project.health)" 
              :show-text="false" 
            />
          </div>
        </div>
        
        <!-- 卡片底部 -->
        <div class="card-footer">
          <span class="update-time">最后更新: {{ project.updateTime }}</span>
          <el-button link type="primary" @click.stop="goToProject(project.id)">进入工作台 <el-icon><ArrowRight /></el-icon></el-button>
        </div>
      </el-card>

      <!-- 兜底的新建卡片 -->
      <div class="project-card new-project-card" @click="openWizard">
        <el-icon :size="40" color="#909399"><Plus /></el-icon>
        <p>创建新项目</p>
      </div>
    </div>

    <!-- 新建项目向导弹窗 -->
    <el-dialog
      v-model="wizardVisible"
      title="新建 AI 辅助项目"
      width="680px"
      destroy-on-close
      class="wizard-dialog"
    >
      <el-steps :active="activeStep" finish-status="success" align-center class="wizard-steps">
        <el-step title="基础信息" description="定义项目目标" />
        <el-step title="需求录入" description="上传或输入PRD" />
        <el-step title="完成创建" description="触发AI拆分" />
      </el-steps>

      <div class="step-content">
        <!-- 步骤 1：基础信息 -->
        <div v-show="activeStep === 0" class="step-pane">
          <el-form :model="projectForm" label-position="top">
            <el-form-item label="项目名称" required>
              <el-input v-model="projectForm.name" placeholder="请输入项目名称，例如：电商后台重构V2.0" />
            </el-form-item>
            <div class="form-row">
              <el-form-item label="负责人" required class="flex-1">
                <el-select v-model="projectForm.manager" placeholder="选择负责人" style="width: 100%">
                  <el-option label="张三 (PM)" value="张三" />
                  <el-option label="李四 (研发)" value="李四" />
                </el-select>
              </el-form-item>
              <el-form-item label="计划周期" required class="flex-1">
                <el-date-picker
                  v-model="projectForm.dateRange"
                  type="daterange"
                  range-separator="至"
                  start-placeholder="开始日期"
                  end-placeholder="结束日期"
                  style="width: 100%"
                />
              </el-form-item>
            </div>
            <el-form-item label="项目目标 (AI 拆分参考上下文)">
              <el-input v-model="projectForm.goal" type="textarea" :rows="3" placeholder="简述该项目的业务价值和核心目标..." />
            </el-form-item>
          </el-form>
        </div>

        <!-- 步骤 2：需求录入 -->
        <div v-show="activeStep === 1" class="step-pane">
          <div class="upload-area">
            <el-upload
              drag
              action="#"
              :auto-upload="false"
              class="prd-upload"
            >
              <el-icon class="el-icon--upload"><Document /></el-icon>
              <div class="el-upload__text">
                将 PRD 文档拖到此处，或 <em>点击上传</em>
              </div>
              <template #tip>
                <div class="el-upload__tip">
                  支持 Word、PDF、Markdown 格式。上传后系统将自动解析需求。
                </div>
              </template>
            </el-upload>
            <el-divider>或者手动输入</el-divider>
            <el-input v-model="projectForm.rawRequirement" type="textarea" :rows="4" placeholder="在此处粘贴原始需求文本..." />
          </div>
        </div>

        <!-- 步骤 3：完成 -->
        <div v-show="activeStep === 2" class="step-pane success-pane">
          <el-result
            icon="success"
            title="项目创建成功！"
            sub-title="项目基础框架已搭建。是否立即让 AI 为您生成第一版的需求拆分草稿？"
          >
          </el-result>
        </div>
      </div>

      <template #footer>
        <div class="wizard-footer">
          <el-button v-if="activeStep > 0 && activeStep < 2" @click="activeStep--">上一步</el-button>
          <el-button v-if="activeStep === 0" type="primary" @click="activeStep++">下一步</el-button>
          <el-button v-if="activeStep === 1" type="primary" @click="activeStep++">完成创建</el-button>
          <el-button v-if="activeStep === 2" type="primary" :loading="isCreating" @click="handleFinish">立即开始 AI 拆分</el-button>
          <el-button v-if="activeStep === 2" @click="wizardVisible = false">稍后处理</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { ElMessage } from 'element-plus';
import { Search, Plus, ArrowRight, Document } from '@element-plus/icons-vue';
import { api } from '../../api/client';

const router = useRouter();

// 搜索与过滤
const searchKey = ref('');
const filterStatus = ref('');

const projects = ref([]);
const isCreating = ref(false);

const normalizeProject = (project, index = 0) => ({
  id: project.id,
  name: project.name,
  manager: project.owner_id || 'PM',
  status: getStatusText(project.status),
  progress: project.status === 'accepted' || project.status === 'launched' ? 100 : 0,
  health: getHealthValue(project.health),
  coverStyle: ['gradient-blue', 'gradient-purple', 'gradient-orange', 'gradient-green'][index % 4],
  updateTime: formatTime(project.updated_at)
});

const loadProjects = async () => {
  try {
    const data = await api.listProjects();
    projects.value = (data || []).map(normalizeProject);
  } catch (error) {
    ElMessage.error(error.message || '项目列表加载失败');
  }
};

onMounted(loadProjects);

// 过滤逻辑
const filteredProjects = computed(() => {
  return projects.value.filter(p => {
    const matchKey = p.name.includes(searchKey.value);
    const matchStatus = filterStatus.value ? getStatusValue(p.status) === filterStatus.value : true;
    return matchKey && matchStatus;
  });
});

// 向导状态
const wizardVisible = ref(false);
const activeStep = ref(0);
const projectForm = reactive({
  name: '', manager: '', dateRange: [], goal: '', rawRequirement: ''
});

const openWizard = () => {
  activeStep.value = 0;
  projectForm.name = '';
  projectForm.manager = '';
  projectForm.dateRange = [];
  projectForm.goal = '';
  projectForm.rawRequirement = '';
  wizardVisible.value = true;
};

const handleFinish = async () => {
  if (!projectForm.name.trim()) {
    ElMessage.warning('请输入项目名称');
    activeStep.value = 0;
    return;
  }
  isCreating.value = true;
  try {
    const project = await api.createProject({
      name: projectForm.name,
      objective: projectForm.goal,
      scope: projectForm.rawRequirement || projectForm.goal,
      owner_id: '',
      owner_role: 'owner'
    });
    if (projectForm.rawRequirement.trim()) {
      const requirement = await api.createRequirement(project.id, {
        content: projectForm.rawRequirement,
        title: '原始需求',
        source_type: 'manual'
      });
      await api.splitRequirement(project.id, {
        requirement_id: requirement.id,
        content: projectForm.rawRequirement
      });
    }
    wizardVisible.value = false;
    await loadProjects();
    router.push(`/project/${project.id}/split`);
  } catch (error) {
    ElMessage.error(error.message || '项目创建失败');
  } finally {
    isCreating.value = false;
  }
};

const goToProject = (id) => {
  router.push(`/project/${id}/overview`);
};

// 辅助格式化
const getStatusValue = (status) => ({ '待开始': 'pending', '进行中': 'doing', '已暂停': 'paused', '已验收': 'accepted', '已上线': 'done', '已归档': 'done' }[status] || '');
const getStatusType = (status) => ({ '待开始': 'info', '进行中': 'primary', '已暂停': 'warning', '已验收': 'success', '已上线': 'success', '已归档': 'info' }[status] || 'info');
const getHealthColor = (health) => ({ 'danger': '#F56C6C', 'warning': '#E6A23C', 'success': '#67C23A', 'normal': '#409EFF' }[health] || '#409EFF');
const getStatusText = (status) => ({ planning: '待开始', active: '进行中', paused: '已暂停', accepted: '已验收', launched: '已上线', archived: '已归档' }[status] || status);
const getHealthValue = (health) => ({ healthy: 'success', attention: 'warning', risk: 'danger', severe_risk: 'danger' }[health] || 'normal');
const formatTime = (value) => value ? new Date(value).toLocaleString() : '-';
</script>

<style scoped>
.project-list-page { padding: 24px; min-height: calc(100vh - 60px); }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 24px; }
.page-header h2 { margin: 0; font-size: 20px; color: #1f2f3d; }
.header-actions { display: flex; gap: 16px; align-items: center; }
.search-input { width: 240px; }
.filter-select { width: 120px; }

/* 网格与卡片 */
.project-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(320px, 1fr)); gap: 24px; }
.project-card { border-radius: 8px; border: none; overflow: hidden; cursor: pointer; display: flex; flex-direction: column; transition: transform 0.2s, box-shadow 0.2s; }
.project-card:hover { transform: translateY(-4px); box-shadow: 0 8px 16px rgba(0,0,0,0.08) !important; }
.project-card :deep(.el-card__body) { padding: 0; display: flex; flex-direction: column; height: 100%; }

.card-cover { height: 120px; padding: 20px; position: relative; color: white; display: flex; align-items: flex-end; }
.gradient-blue { background: linear-gradient(135deg, #409EFF, #36cfc9); }
.gradient-purple { background: linear-gradient(135deg, #722ed1, #b37feb); }
.gradient-orange { background: linear-gradient(135deg, #E6A23C, #ffc069); }
.gradient-green { background: linear-gradient(135deg, #67C23A, #95de64); }

.cover-content { width: 100%; display: flex; justify-content: space-between; align-items: center; }
.cover-content h3 { margin: 0; font-size: 18px; font-weight: 600; text-shadow: 0 1px 2px rgba(0,0,0,0.2); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; flex: 1; margin-right: 12px; }
.status-tag { box-shadow: 0 1px 2px rgba(0,0,0,0.1); }

.card-body { padding: 20px; flex: 1; }
.info-row { font-size: 14px; color: #606266; margin-bottom: 16px; display: flex; align-items: center; }
.flex-center { display: flex; align-items: center; gap: 8px; font-weight: 500; color: #303133; }
.mini-avatar { background: #c6e2ff; color: #409EFF; font-size: 12px; }

.progress-section { margin-top: auto; }
.progress-header { display: flex; justify-content: space-between; font-size: 13px; color: #606266; margin-bottom: 8px; }
.percentage { font-weight: 600; color: #303133; }

.card-footer { padding: 12px 20px; background: #fafafa; border-top: 1px solid #f0f2f5; display: flex; justify-content: space-between; align-items: center; font-size: 12px; }
.update-time { color: #909399; }

/* 新建按钮兜底卡片 */
.new-project-card { border: 2px dashed #dcdfe6; background: transparent; justify-content: center; align-items: center; color: #909399; min-height: 240px; box-shadow: none !important; }
.new-project-card:hover { border-color: #409EFF; color: #409EFF; background: #ecf5ff; }
.new-project-card p { margin-top: 12px; font-size: 15px; font-weight: 500; }

/* 向导样式 */
.wizard-steps { margin-bottom: 32px; padding: 0 20px; }
.step-content { min-height: 280px; }
.form-row { display: flex; gap: 20px; }
.flex-1 { flex: 1; }
.upload-area { display: flex; flex-direction: column; gap: 16px; }
.prd-upload :deep(.el-upload-dragger) { padding: 32px; }
.success-pane { display: flex; justify-content: center; align-items: center; height: 100%; }
</style>
