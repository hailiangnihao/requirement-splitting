<template>
  <div class="split-container">
    <!-- 左侧：原始需求区 -->
    <div class="left-panel">
      <div class="panel-header">
        <h3>原始需求</h3>
        <el-upload
          :auto-upload="false"
          :show-file-list="false"
          accept=".doc,.docx,.pdf,.txt"
          :on-change="handleFileUpload"
        >
          <el-button type="primary" plain size="small">上传文档 (Word/PDF)</el-button>
        </el-upload>
      </div>
      <div class="input-wrapper">
        <el-input
          v-model="rawRequirement"
          type="textarea"
          placeholder="请输入您的原始需求描述，例如：&#10;我们需要一个用户登录功能，支持账号密码登录，登录后跳转到项目列表。如果没有账号可以进行注册..."
          class="req-input"
        />
      </div>
      <div class="panel-footer">
        <div class="ai-config">
          <span class="label">拆分策略：</span>
          <el-radio-group v-model="splitStrategy" size="small">
            <el-radio-button label="stage1">一阶段 (模块/里程碑)</el-radio-button>
            <el-radio-button label="stage2">全量深度拆分</el-radio-button>
          </el-radio-group>
        </div>
        <el-button type="primary" size="large" @click="handleAiSplit" :loading="isSplitting">
          <el-icon class="el-icon--left"><MagicStick /></el-icon>
          AI 智能拆分
        </el-button>
      </div>
    </div>

    <!-- 右侧：AI 拆分草稿 (结构化树) -->
    <div class="right-panel">
      <div class="panel-header">
        <div class="title-with-badge">
          <h3>AI 拆分草稿</h3>
          <el-tag v-if="hasResult" type="warning" size="small" effect="light">草稿未发布</el-tag>
        </div>
        <div class="actions">
          <el-button size="small" @click="resetTree" :disabled="isSplitting">重置</el-button>
          <el-button type="success" size="small" :disabled="!hasResult" @click="publishPlan">
            确认并发布正式计划
          </el-button>
        </div>
      </div>

      <div class="tree-container" v-loading="isSplitting" :element-loading-text="progressMessage || 'AI 正在深度思考和拆分中...'">
        <el-empty v-if="!hasResult && !isSplitting" description="暂无拆分结果，请在左侧输入需求并点击拆分" />

        <!-- AI思考过程展示 -->
        <div v-if="showThinking && thinkingProcess" class="thinking-box">
          <div class="thinking-header">
            <el-icon class="thinking-icon"><Cpu /></el-icon>
            <span>AI 思考过程</span>
            <el-button link type="primary" size="small" @click="showThinking = !showThinking">
              {{ showThinking ? '收起' : '展开' }}
            </el-button>
          </div>
          <div class="thinking-content" v-if="showThinking">
            <pre>{{ thinkingProcess }}</pre>
          </div>
        </div>

        <!-- 使用卡片式布局替代树形结构 -->
        <div v-if="hasResult" class="plan-cards">
          <div v-for="module in treeData" :key="module.id" class="module-card">
            <div class="module-header">
              <div class="module-title">
                <el-icon><FolderOpened /></el-icon>
                <span>{{ module.label }}</span>
              </div>
              <div class="module-actions">
                <el-button link type="primary" size="small" @click="handleEdit(module)">
                  <el-icon><Edit /></el-icon>
                </el-button>
                <el-button link type="danger" size="small" @click="handleDeleteModule(module)">
                  <el-icon><Delete /></el-icon>
                </el-button>
              </div>
            </div>
            <div class="module-desc" v-if="module.desc">{{ module.desc }}</div>

            <!-- 功能点列表 -->
            <div v-for="feature in module.children" :key="feature.id" class="feature-card">
              <div class="feature-header">
                <div class="feature-title">
                  <el-tag type="success" size="small">功能点</el-tag>
                  <span>{{ feature.label }}</span>
                </div>
                <div class="feature-actions">
                  <el-button link type="primary" size="small" @click="handleEdit(feature)">编辑</el-button>
                  <el-button link type="danger" size="small" @click="handleDeleteItem(module, feature)">删除</el-button>
                </div>
              </div>
              <div class="feature-desc" v-if="feature.desc">{{ feature.desc }}</div>

              <!-- 任务和测试用例 -->
              <div class="items-grid">
                <div v-for="item in feature.children" :key="item.id" class="item-card">
                  <div class="item-header">
                    <el-tag :type="getItemTagType(item.type)" size="small">{{ item.type }}</el-tag>
                    <div class="item-actions">
                      <el-button link type="primary" size="small" @click="handleEdit(item)">
                        <el-icon><Edit /></el-icon>
                      </el-button>
                      <el-button link type="danger" size="small" @click="handleDeleteItem(feature, item)">
                        <el-icon><Delete /></el-icon>
                      </el-button>
                    </div>
                  </div>
                  <div class="item-title">{{ item.label }}</div>
                  <div class="item-desc" v-if="item.desc">{{ item.desc }}</div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { ElMessage, ElMessageBox } from 'element-plus';
import { MagicStick, FolderOpened, Edit, Delete, Cpu } from '@element-plus/icons-vue';
import { api, baseURL } from '../../api/client';

const route = useRoute();
const router = useRouter();
const rawRequirement = ref('');
const splitStrategy = ref('stage2');
const isSplitting = ref(false);
const hasResult = ref(false);
const treeData = ref([]);
const currentDraftId = ref('');
const progressMessage = ref(''); // 进度消息
const thinkingProcess = ref(''); // AI思考过程
const showThinking = ref(false); // 是否显示思考过程
const generatedContent = ref(''); // 已生成的内容

const defaultProps = { children: 'children', label: 'label' };

const buildTreeFromDraft = (output) => {
  const tasksByFeature = new Map();
  const casesByFeature = new Map();
  for (const task of output.dev_tasks || []) {
    const list = tasksByFeature.get(task.feature_point_key) || [];
    list.push({ id: task.key, label: task.title, type: '开发任务', desc: task.description });
    tasksByFeature.set(task.feature_point_key, list);
  }
  for (const testCase of output.test_cases || []) {
    const list = casesByFeature.get(testCase.feature_point_key) || [];
    list.push({ id: testCase.key, label: testCase.title, type: '测试用例', desc: testCase.expected_result });
    casesByFeature.set(testCase.feature_point_key, list);
  }
  const featuresByModule = new Map();
  for (const feature of output.feature_points || []) {
    const children = [
      ...(tasksByFeature.get(feature.key) || []),
      ...(casesByFeature.get(feature.key) || []),
      ...(output.acceptance_items || [])
        .filter(item => item.feature_point_key === feature.key)
        .map(item => ({ id: item.key, label: item.title, type: '验收项', desc: item.pass_criteria }))
    ];
    const list = featuresByModule.get(feature.module_key) || [];
    list.push({ id: feature.key, label: feature.title, type: '功能点', desc: feature.description, children });
    featuresByModule.set(feature.module_key, list);
  }
  return (output.modules || []).map(module => ({
    id: module.key,
    label: module.name,
    type: '模块',
    desc: module.description,
    children: featuresByModule.get(module.key) || []
  }));
};

const loadLatestDraft = async () => {
  const projectId = route.params.id;
  if (!projectId) return;
  try {
    const drafts = await api.listDrafts(projectId);
    const draft = (drafts || [])[0];
    if (draft) {
      currentDraftId.value = draft.id;
      treeData.value = buildTreeFromDraft(draft.output_json || {});
      hasResult.value = treeData.value.length > 0;
    }
    const requirements = await api.listRequirements(projectId);
    if (requirements?.[0]?.content) {
      rawRequirement.value = requirements[0].content;
    }
  } catch (error) {
    ElMessage.error(error.message || '拆分草稿加载失败');
  }
};

onMounted(loadLatestDraft);

const handleAiSplit = async () => {
  if (!rawRequirement.value.trim()) {
    ElMessage.warning('请先输入原始需求');
    return;
  }
  const projectId = route.params.id;

  // 重置状态
  isSplitting.value = true;
  hasResult.value = false;
  progressMessage.value = '准备中...';
  thinkingProcess.value = '';
  generatedContent.value = '';
  showThinking.value = true; // 默认展开思考过程

  try {
    // 先创建需求
    const requirement = await api.createRequirement(projectId, {
      title: '原始需求',
      content: rawRequirement.value,
      source_type: 'manual'
    });

    // 使用流式API
    await splitWithStream(projectId, {
      requirement_id: requirement.id,
      content: rawRequirement.value
    });

  } catch (error) {
    ElMessage.error(error.message || 'AI 拆分失败');
    isSplitting.value = false;
  }
};

// 流式拆分
const splitWithStream = async (projectId, payload) => {
  return new Promise((resolve, reject) => {
    // 由于EventSource不支持POST，我们使用fetch来处理SSE
    fetch(`${baseURL}/api/projects/${projectId}/ai/split-requirement/stream`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(payload)
    })
    .then(response => {
      const reader = response.body.getReader();
      const decoder = new TextDecoder();

      function read() {
        reader.read().then(({ done, value }) => {
          if (done) {
            isSplitting.value = false;
            resolve();
            return;
          }

          const chunk = decoder.decode(value);
          const lines = chunk.split('\n');

          lines.forEach(line => {
            if (line.startsWith('data: ')) {
              try {
                const data = JSON.parse(line.substring(6));

                if (data.type === 'progress') {
                  progressMessage.value = data.message;
                } else if (data.type === 'thinking') {
                  // 更新思考过程
                  thinkingProcess.value = data.thinking;
                  progressMessage.value = data.message;
                } else if (data.type === 'content') {
                  // 更新生成的内容
                  generatedContent.value = data.content;
                  progressMessage.value = data.message;

                  // 尝试实时解析JSON并更新树
                  try {
                    const partialData = JSON.parse(data.content);
                    if (partialData.modules || partialData.milestones) {
                      treeData.value = buildTreeFromDraft(partialData);
                      hasResult.value = true;
                    }
                  } catch (e) {
                    // JSON还不完整，忽略错误
                  }
                } else if (data.type === 'result' || data.type === 'complete') {
                  // 最终结果
                  if (data.data) {
                    treeData.value = buildTreeFromDraft(data.data);
                    hasResult.value = true;
                  }
                  if (data.type === 'complete') {
                    if (data.data?.id) {
                      currentDraftId.value = data.data.id;
                    }
                    ElMessage.success('AI 拆分完成，已生成拆分草稿，请人工确认');
                    isSplitting.value = false;
                  }
                } else if (data.type === 'error') {
                  ElMessage.error(data.message);
                  isSplitting.value = false;
                  reject(new Error(data.message));
                }
              } catch (e) {
                console.error('Failed to parse SSE data:', e);
              }
            }
          });

          read();
        });
      }

      read();
    })
    .catch(error => {
      ElMessage.error('连接失败: ' + error.message);
      isSplitting.value = false;
      reject(error);
    });
  });
};

const resetTree = () => {
  treeData.value = [];
  hasResult.value = false;
};

const publishPlan = async () => {
  if (!currentDraftId.value) {
    ElMessage.warning('没有可发布的草稿');
    return;
  }
  try {
    await api.publishDraft(route.params.id, currentDraftId.value);
    ElMessage.success('正式计划发布成功，可在任务看板和测试页查看');
    router.push(`/project/${route.params.id}/kanban`);
  } catch (error) {
    ElMessage.error(error.message || '正式计划发布失败');
  }
};

// 根据节点类型返回不同的标签颜色
const getTagType = (type) => {
  const map = {
    '模块': 'primary',
    '功能点': 'success',
    '开发任务': 'warning',
    '测试用例': 'info',
    '验收项': 'danger'
  };
  return map[type] || 'info';
};

// 编辑节点
const handleEdit = (data) => {
  ElMessageBox.prompt('', `编辑 ${data.type}`, {
    confirmButtonText: '保存',
    cancelButtonText: '取消',
    inputValue: data.label,
    inputType: 'textarea',
    inputPlaceholder: '请输入内容',
    customClass: 'edit-dialog'
  }).then(({ value }) => {
    if (value && value.trim()) {
      data.label = value.trim();
      ElMessage.success('修改成功');
    }
  }).catch(() => {
    // 用户取消
  });
};

// 删除节点
const handleDelete = (node, data) => {
  ElMessageBox.confirm(
    `确定要删除 ${data.type} "${data.label}" 吗？`,
    '删除确认',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }
  ).then(() => {
    const parent = node.parent;
    const children = parent.data.children || parent.data;
    const index = children.findIndex(d => d.id === data.id);
    if (index !== -1) {
      children.splice(index, 1);
      ElMessage.success('删除成功');
    }
  }).catch(() => {
    // 用户取消
  });
};

// 删除模块
const handleDeleteModule = (module) => {
  ElMessageBox.confirm(
    `确定要删除模块 "${module.label}" 及其所有子项吗？`,
    '删除确认',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }
  ).then(() => {
    const index = treeData.value.findIndex(m => m.id === module.id);
    if (index !== -1) {
      treeData.value.splice(index, 1);
      ElMessage.success('删除成功');
    }
  }).catch(() => {});
};

// 删除子项（功能点、任务、测试用例）
const handleDeleteItem = (parent, item) => {
  ElMessageBox.confirm(
    `确定要删除 "${item.label}" 吗？`,
    '删除确认',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }
  ).then(() => {
    const index = parent.children.findIndex(c => c.id === item.id);
    if (index !== -1) {
      parent.children.splice(index, 1);
      ElMessage.success('删除成功');
    }
  }).catch(() => {});
};

// 获取标签类型
const getItemTagType = (type) => {
  const map = {
    '开发任务': 'warning',
    '测试用例': 'info',
    '验收项': 'danger'
  };
  return map[type] || 'info';
};

// 上传文档
const handleFileUpload = (file) => {
  const fileName = file.name;
  const fileType = fileName.substring(fileName.lastIndexOf('.') + 1).toLowerCase();

  if (!['doc', 'docx', 'pdf', 'txt'].includes(fileType)) {
    ElMessage.error('仅支持 Word、PDF 和 TXT 格式');
    return;
  }

  // TODO: 实现文档解析功能
  // 1. 上传文件到后端
  // 2. 后端解析文档内容（使用 OCR 或文档解析库）
  // 3. 将解析后的文本填充到输入框

  ElMessage.info({
    message: '文档上传功能开发中，敬请期待！',
    duration: 2000
  });

  // 临时方案：读取文本文件内容
  if (fileType === 'txt') {
    const reader = new FileReader();
    reader.onload = (e) => {
      rawRequirement.value = e.target.result;
      ElMessage.success('文本文件读取成功');
    };
    reader.readAsText(file.raw);
  }
};
</script>

<style scoped>
.split-container {
  display: flex;
  gap: 24px;
  /* 撑满外层 page-container 的高度 */
  height: calc(100vh - 110px); 
}
.left-panel, .right-panel {
  flex: 1;
  background: #fff;
  border-radius: 8px;
  padding: 20px;
  display: flex;
  flex-direction: column;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.05);
  min-width: 0; /* 防止 flex 子项溢出 */
}
.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  border-bottom: 1px solid #f0f2f5;
  padding-bottom: 12px;
}
.panel-header h3 { margin: 0; font-size: 16px; color: #1f2f3d; }
.title-with-badge { display: flex; align-items: center; gap: 8px; }
.input-wrapper {
  flex: 1;
  display: flex;
  margin-bottom: 16px;
}
/* 深度选择器让输入框高度撑满 */
.input-wrapper :deep(.el-textarea) { flex: 1; display: flex; }
.input-wrapper :deep(.el-textarea__inner) { flex: 1; resize: none; font-size: 14px; }
.panel-footer { display: flex; justify-content: space-between; align-items: center; }
.ai-config .label { font-size: 13px; color: #606266; margin-right: 8px; }

.tree-container {
  flex: 1;
  overflow-y: auto;
  border: 1px solid #ebeef5;
  border-radius: 4px;
  padding: 16px;
}

/* AI思考过程展示样式 */
.thinking-box {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 12px;
  padding: 20px;
  margin-bottom: 24px;
  box-shadow: 0 4px 6px rgba(102, 126, 234, 0.2);
  animation: fadeIn 0.5s ease-in;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(-10px); }
  to { opacity: 1; transform: translateY(0); }
}

.thinking-header {
  display: flex;
  align-items: center;
  gap: 10px;
  color: white;
  font-weight: 600;
  font-size: 16px;
  margin-bottom: 12px;
}

.thinking-icon {
  font-size: 20px;
  animation: pulse 2s infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

.thinking-content {
  background: rgba(255, 255, 255, 0.95);
  border-radius: 8px;
  padding: 16px;
  max-height: 400px;
  overflow-y: auto;
}

.thinking-content pre {
  margin: 0;
  white-space: pre-wrap;
  word-wrap: break-word;
  font-size: 13px;
  line-height: 1.6;
  color: #303133;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
}

/* 卡片式布局样式 */
.plan-cards {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.module-card {
  background: #fff;
  border: 2px solid #409EFF;
  border-radius: 8px;
  padding: 16px;
}

.module-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.module-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  font-weight: 600;
  color: #409EFF;
}

.module-desc {
  font-size: 13px;
  color: #909399;
  margin-bottom: 16px;
  padding-left: 28px;
}

.module-actions, .feature-actions, .item-actions {
  display: flex;
  gap: 4px;
}

.feature-card {
  background: #f0f9ff;
  border: 1px solid #d9ecff;
  border-radius: 6px;
  padding: 12px;
  margin-bottom: 12px;
}

.feature-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.feature-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  font-weight: 500;
  color: #303133;
}

.feature-desc {
  font-size: 12px;
  color: #606266;
  margin-bottom: 12px;
  padding-left: 60px;
}

.items-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 12px;
  margin-top: 12px;
}

.item-card {
  background: #fff;
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  padding: 10px;
  transition: all 0.2s;
}

.item-card:hover {
  border-color: #409EFF;
  box-shadow: 0 2px 8px rgba(64, 158, 255, 0.1);
}

.item-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.item-title {
  font-size: 13px;
  font-weight: 500;
  color: #303133;
  margin-bottom: 6px;
  line-height: 1.4;
}

.item-desc {
  font-size: 12px;
  color: #909399;
  line-height: 1.4;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

/* 旧的树形样式保留作为备用 */
.custom-tree-node {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 14px;
  padding-right: 8px;
  width: 100%;
}
.node-main { display: flex; align-items: center; }
.node-tag { margin-right: 8px; width: 66px; text-align: center; }
.node-label { font-weight: 500; color: #303133; }
.node-extra { display: flex; align-items: center; }
.node-desc {
  margin-right: 16px;
  font-size: 12px;
  color: #909399;
  max-width: 200px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.node-actions { display: none; }
.custom-tree-node:hover .node-actions { display: block; }

/* 编辑对话框样式 */
:deep(.edit-dialog .el-message-box__input textarea) {
  min-height: 100px;
  resize: vertical;
}
</style>
